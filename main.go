package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Laughs-In-Flowers/countfloyd/lib/server"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
	"github.com/Laughs-In-Flowers/xrr"
)

type Options struct {
	LogFormatter string
	*sOptions
	*pOptions
	*qOptions
	*aOptions
}

func NewOptions() *Options {
	return &Options{
		sOptions: &sOptions{
			LocalPath:  "/tmp/cfc",
			SocketPath: "/tmp/countfloyd_0_0-socket",
			Timeout:    2 * time.Second,
		},
		pOptions: &pOptions{},
		qOptions: &qOptions{},
		aOptions: &aOptions{
			aStore:    "stdout",
			aLocation: currentLoc,
		},
	}
}

type sOptions struct {
	LocalPath, SocketPath string
	Timeout               time.Duration
}

type pOptions struct {
	pFeature, pComponent, pEntity      string
	pConstructorPlugin, pFeaturePlugin string
	pGroup                             string
}

func pathError(e error) bool {
	_, ok := e.(*os.PathError)
	return ok
}

func parseDirFiles(in string) []string {
	var ret []string

	spl := strings.Split(in, ",")
	for _, i := range spl {
		f, oErr := os.Open(i)
		if !pathError(oErr) {
			fi, fErr := f.Stat()
			if !pathError(fErr) {
				isDir := fi.IsDir()
				switch {
				case isDir:
					dn := filepath.Dir(i)
					di, _ := f.Readdir(-1)
					for _, dif := range di {
						if !dif.IsDir() {
							np := filepath.Join(dn, dif.Name())
							ret = append(ret, np)
						}
					}
				case !isDir:
					ret = append(ret, i)
				}
			}
		}
		f.Close()
	}

	return ret
}

func (o *Options) files(tag string) []string {
	switch tag {
	case "constructor-plugin":
		return strings.Split(tag, ",")
	case "feature-plugin":
		return strings.Split(tag, ",")
	case "features":
		return parseDirFiles(o.pFeature)
	case "components":
		return parseDirFiles(o.pComponent)
	case "entities":
		return parseDirFiles(o.pEntity)
	}
	return nil
}

type qOptions struct {
	qFeature, qComponent, qEntity string
}

type aOptions struct {
	aNumber                       float64
	aFeature, aComponent, aEntity string
	aStore, aLocation             string
}

var MoreThanAllowableError = xrr.Xrror("Can only request one of feature, component, or entity: %v").Out

func (o *Options) Act(d *data.Vector, single bool) (string, *data.Vector, error) {
	var aSwitch = make(map[string]bool)
	var action string
	switch {
	case o.aFeature != "":
		action = "apply_feature"
		aSwitch[action] = true
		spl := strings.Split(o.aFeature, ",")
		d.Set(data.NewStringsItem("meta.feature", spl...))
	case o.qFeature != "":
		action = "query_feature"
		aSwitch[action] = true
		d.Set(data.NewStringItem("query_feature", o.qFeature))
	case o.aComponent != "":
		action = "apply_component"
		aSwitch[action] = true
		spl := strings.Split(o.aComponent, ",")
		d.Set(data.NewStringsItem("meta.component", spl...))
	case o.qComponent != "":
		action = "query_component"
		aSwitch[action] = true
		d.Set(data.NewStringItem("query_component", o.qComponent))
	case o.aEntity != "":
		action = "apply_entity"
		aSwitch[action] = true
		d.Set(data.NewStringItem("meta.entity", o.aEntity))
	case o.qEntity != "":
		action = "query_entity"
		aSwitch[action] = true
		d.Set(data.NewStringItem("query_entity", o.qEntity))
	}
	if single {
		var acount []string
		for k, v := range aSwitch {
			if v {
				acount = append(acount, k)
			}
		}
		if len(acount) > 1 {
			return "", nil, MoreThanAllowableError(acount)
		}
	}
	return action, d, nil
}

var (
	currentDir  string
	defaultFile string = "out"
	currentLoc  string
	L           log.Logger
)

func init() {
	currentDir, _ = os.Getwd()
	currentLoc = filepath.Join(currentDir, defaultFile)
	L = log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
}

type sonnect struct {
	formatter, local, socket string
	timeout                  time.Duration
}

var Sonnect *sonnect

func socketFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.LocalPath, "local", o.LocalPath, "Specify a local path for communication to the server.")
	fs.StringVar(&o.SocketPath, "socket", o.SocketPath, "Specify the socket path of the server.")
}

func TopCommand() flip.Command {
	o := NewOptions()

	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("", flip.ContinueOnError)
		fs.StringVar(&o.LogFormatter, "logFormatter", o.LogFormatter, "Sets the environment logger formatter.")
		socketFlags(o, fs)
		return fs
	}(o)

	return flip.NewCommand(
		"",
		"countfloyd",
		`Top level flag usage.`,
		0,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			if o.LogFormatter != "" {
				switch o.LogFormatter {
				case "text", "stdout":
					L.SwapFormatter(log.GetFormatter("countfloyd_text"))
				default:
					L.SwapFormatter(log.GetFormatter(o.LogFormatter))
				}
			}
			Sonnect = &sonnect{o.LogFormatter, o.LocalPath, o.SocketPath, o.Timeout}
			return c, flip.ExitNo
		},
		fs,
	)
}

func logConnect(service, action, is, point, err string) {
	L.Printf("%s    %s    %s    %s    %s",
		service,
		action,
		is,
		point,
		err,
	)
}

func onError(service, action, point string, err error) flip.ExitStatus {
	logConnect(service, action, "failure", point, err.Error())
	return flip.ExitFailure
}

func onSuccess(service, action string) flip.ExitStatus {
	logConnect(service, action, "succcess", "", "")
	return flip.ExitSuccess
}

func isQuit(service, action string) flip.ExitStatus {
	logConnect(service, action, "", "", "")
	return flip.ExitSuccess
}

func connect(s *sonnect, service, action string, d *data.Vector) flip.ExitStatus {
	req := server.NewRequest(
		server.ByteService(service),
		server.ByteAction(action),
		d,
	)

	conn, cErr := connection(s.local, s.socket)
	defer cleanup(conn, s.local)
	if cErr != nil {
		return onError(service, action, "connection", cErr)
	}

	_, wErr := conn.Write(req.ToByte())
	if wErr != nil {
		return onError(service, action, "write", wErr)
	}

	resp, rErr := response(conn, s.timeout)
	if rErr != nil {
		return onError(service, action, "response", rErr)
	}

	if action == "quit" {
		return isQuit(service, action)
	}

	sresp, mErr := unmarshal(resp)
	if mErr != nil {
		return onError(service, action, "unmarshal", mErr)
	}
	if sresp.Error != "" {
		return onError(service, action, "result", errors.New(sresp.Error))
	}

	if sresp.Data != nil {
		sErr := store(sresp.Data)
		if sErr != nil {
			return onError(service, action, "store", sErr)
		}
	}

	return onSuccess(service, action)
}

func connection(local, socket string) (*net.UnixConn, error) {
	t := "unix"
	laddr := net.UnixAddr{local, t}
	conn, err := net.DialUnix(t, &laddr, &net.UnixAddr{socket, t})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

var ResponseError = xrr.Xrror("Error getting a response from the countfloyd server: %s").Out

func response(c io.Reader, timeout time.Duration) ([]byte, error) {
	t := time.After(timeout)
	for {
		select {
		case <-t:
			return nil, ResponseError("time out")
		default:
			buf := new(bytes.Buffer)
			io.Copy(buf, c)
			return buf.Bytes(), nil
		}
	}
	return nil, ResponseError("no response")
}

func unmarshal(b []byte) (*server.Response, error) {
	ret := server.EmptyResponse()
	mErr := json.Unmarshal(b, &ret)
	if mErr != nil {
		return nil, mErr
	}
	return ret, nil
}

func store(v *data.Vector) error {
	rs := v.ToStrings(retrievalKey)
	s, gsErr := data.GetStore(rs[0], rs)
	if gsErr != nil {
		return gsErr
	}

	s.Swap(v)
	_, swErr := s.Out()
	if swErr != nil {
		return swErr
	}

	return nil
}

func cleanup(c *net.UnixConn, local string) {
	if c != nil {
		c.Close()
	}
	os.Remove(local)
}

func getStartPopulate(a []string, o *Options) []string {
	if ff := o.files("features"); len(ff) > 0 {
		a = append(a, "-features", strings.Join(ff, ","))
	}

	if cf := o.files("components"); len(cf) > 0 {
		a = append(a, "-components", strings.Join(cf, ","))
	}

	if ef := o.files("entities"); len(ef) > 0 {
		a = append(a, "-entities", strings.Join(ef, ","))
	}

	return a
}

func filesFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.pFeature, "feature", "", "Populate features from files or directories.")
	fs.StringVar(&o.pComponent, "component", "", "Populate components from files or directories.")
	fs.StringVar(&o.pEntity, "entity", "", "Populate entities from files or directories.")
}

func StartCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("", flip.ContinueOnError)
		filesFlags(o, fs)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"start",
		"start a countfloyd server",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			cs := []string{"-socket", Sonnect.socket, "-logFormatter", Sonnect.formatter}
			cs = getStartPopulate(cs, o)
			cs = append(cs, "start")
			cmd := exec.Command("cfs", cs...)
			cmd.Stdout = os.Stdout
			err := cmd.Start()
			if err != nil {
				return c, flip.ExitFailure
			}
			return c, flip.ExitSuccess
		},
		fs,
	)
}

func end(v string) flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet(v, flip.ContinueOnError)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		v,
		fmt.Sprintf("%s the countfloyd server", v),
		2,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return c, connect(Sonnect, "system", "quit", nil)
		},
		fs,
	)
}

func StopCommand() flip.Command {
	return end("stop")
}

func QuitCommand() flip.Command {
	return end("quit")
}

func StatusCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("status", flip.ContinueOnError)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"status",
		"the status of a countfloyd server",
		3,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			return c, connect(Sonnect, "query", "status", newVector(o))
		},
		fs,
	)
}

func queryFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.qFeature, "feature", o.qFeature, "return information for this specified feature")
	fs.StringVar(&o.qComponent, "component", o.qComponent, "return information for this specified component")
	fs.StringVar(&o.qEntity, "entity", o.qEntity, "return information for this specified entity")
}

func queryVector(o *Options) (string, *data.Vector, error) {
	var action string
	d := newVector(o)
	var err error
	action, d, err = o.Act(d, true)
	return action, d, err
}

func QueryCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("query", flip.ContinueOnError)
		queryFlags(o, fs)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"query",
		"query a countfloyd server for feature information",
		4,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			action, d, err := queryVector(o)
			if err != nil {
				L.Print(err)
				return c, flip.ExitUsageError
			}
			return c, connect(Sonnect, "query", action, d)
		},
		fs,
	)
}

func populateVector(o *Options) (string, *data.Vector, error) {
	d := newVector(o)
	cp := data.NewStringsItem("constructor-plugin", o.files("constructor-plugin")...)
	fp := data.NewStringsItem("feature-plugin", o.files("feature-plugin")...)
	fs := data.NewStringsItem("features", o.files("features")...)
	cs := data.NewStringsItem("components", o.files("components")...)
	es := data.NewStringsItem("entities", o.files("entities")...)
	d.Set(cp, fp, fs, cs, es)
	d.SetStrings("groups", o.pGroup)
	return "populate_from_files", d, nil
}

func PopulateCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("populate", flip.ContinueOnError)
		fs.StringVar(&o.pGroup, "group", o.pGroup, "Comma separated string list of set tags to apply to all features read in with this instance.")
		fs.StringVar(&o.pConstructorPlugin, "constructorPlugin", o.pConstructorPlugin, "Comma separated string list of directories containing Constructor plugins.")
		fs.StringVar(&o.pFeaturePlugin, "featurePlugin", o.pFeaturePlugin, "Comma separated string list of directories containing Feature plugins.")
		filesFlags(o, fs)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"populate",
		"populate a countfloyd server with features from provided files.",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			action, d, _ := populateVector(o)
			return c, connect(Sonnect, "data", action, d)
		},
		fs,
	)
}

func depopulateVector(o *Options) (string, *data.Vector, error) {
	d := newVector(o)
	d.SetStrings("groups", o.pGroup)
	return "depopulate", d, nil
}

func DepopulateCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("depopulate", flip.ContinueOnError)
		fs.StringVar(&o.pGroup, "groups", o.pGroup, "Comma separate string list of group tags to remove.")
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"depopulate",
		"depopulate a countfloyd server by feature grouping.",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			action, d, _ := depopulateVector(o)
			return c, connect(Sonnect, "data", action, d)
		},
		fs,
	)
	return nil
}

func applyFlags(o *Options, fs *flip.FlagSet) {
	fs.Float64Var(&o.aNumber, "priority", 0, "An float64 value for nonspecific use by the feature.")
	fs.StringVar(&o.aFeature, "feature", "", "A comma delimited list of features to apply.")
	fs.StringVar(&o.aComponent, "component", "", "A comma delimited list of components to apply.")
	fs.StringVar(&o.aEntity, "entity", "", "A specific entity to apply.")
	fs.StringVar(&o.aStore, "store", o.aStore, "A data store to use [out, json, jsonf, yaml]")
	fs.StringVar(&o.aLocation, "location", o.aLocation, "The location the store writes to if the store requires a location")
}

func applyVector(o *Options) (string, *data.Vector, error) {
	var action string
	d := newVector(o)
	var err error
	action, d, err = o.Act(d, true)
	return action, d, err
}

func ApplyCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("apply", flip.ContinueOnError)
		applyFlags(o, fs)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"apply",
		"apply a set of features",
		2,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			action, d, err := applyVector(o)
			if err != nil {
				L.Print(err)
				return c, flip.ExitUsageError
			}
			return c, connect(Sonnect, "data", action, d)
		},
		fs,
	)
}

var retrievalKey = "store.retrieval.string"

func newVector(o *Options) *data.Vector {
	d := data.New("")
	d.Set(data.NewFloat64Item("meta.priority", o.aNumber))
	p := filepath.Clean(o.aLocation)
	dir, file := filepath.Dir(p), filepath.Base(p)
	d.Set(data.NewStringsItem(retrievalKey, o.aStore, dir, file))
	return d
}

var (
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
	F              flip.Flpr
)

func init() {
	log.SetFormatter("countfloyd_text", log.MakeTextFormatter(versionPackage))
	F = flip.New("countfloyd")
	F.AddBuiltIn("version", versionPackage, versionTag, versionHash, versionDate).
		AddBuiltIn("help").
		SetGroup("top", -1, TopCommand()).
		SetGroup("control",
			1,
			StartCommand(),
			StopCommand(),
			QuitCommand(),
			StatusCommand(),
			QueryCommand()).
		SetGroup("action",
			2,
			PopulateCommand(),
			DepopulateCommand(),
			ApplyCommand())
}

func main() {
	ctx := context.Background()
	exit := F.Execute(ctx, os.Args)
	os.Exit(exit)
}
