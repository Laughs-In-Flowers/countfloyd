package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cf "github.com/Laughs-In-Flowers/countfloyd/lib/server"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
)

type sonnect struct {
	formatter, local, socket string
	timeout                  time.Duration
}

var current *sonnect

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
		aOptions: &aOptions{},
	}
}

type sOptions struct {
	LocalPath, SocketPath string
	Timeout               time.Duration
}

func socketFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.LocalPath, "local", o.LocalPath, "Specify a local path for communication to the server.")
	fs.StringVar(&o.SocketPath, "socket", o.SocketPath, "Specify the socket path of the server.")
}

type pOptions struct {
	pFeatureF, pComponentF, pEntityF string
	pFeatureD, pComponentD, pEntityD string
}

func filesFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.pFeatureD, "featuresDir", "", "A directory to locate feature files in.")
	fs.StringVar(&o.pFeatureF, "featuresFiles", "", "A comma delimited list of feature files.")
	fs.StringVar(&o.pComponentD, "componentsDir", "", "A directory to locate component files in.")
	fs.StringVar(&o.pComponentF, "componentsFiles", "", "A comma delimited list of component files.")
	fs.StringVar(&o.pEntityD, "entitiesDir", "", "A directory to locate entity files in.")
	fs.StringVar(&o.pEntityF, "entitiesFiles", "", "A comma delimited list of entity files.")
}

func parseDirFiles(dir, files string) []string {
	var ret []string

	if dir != "" {
		df, err := ioutil.ReadDir(dir)
		if err != nil {
			L.Fatal(err.Error())
		}

		for _, f := range df {
			ret = append(ret, filepath.Join(dir, f.Name()))
		}
	}

	if files != "" {
		lf := strings.Split(files, ",")
		for _, f := range lf {
			ret = append(ret, f)
		}
	}

	return ret
}

func (o *Options) files(tag string) []string {
	switch tag {
	case "features":
		return parseDirFiles(o.pFeatureD, o.pFeatureF)
	case "components":
		return parseDirFiles(o.pComponentD, o.pComponentF)
	case "entities":
		return parseDirFiles(o.pEntityD, o.pEntityF)
	}
	return nil
}

type qOptions struct {
	qFeature string
}

type aOptions struct {
	aNumber                                     int
	aFeatures, aComponent, aComponents, aEntity string
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
		func(c context.Context, a []string) flip.ExitStatus {
			if o.LogFormatter != "" {
				switch o.LogFormatter {
				case "text", "stdout":
					L.SwapFormatter(log.GetFormatter("countfloyd_text"))
				default:
					L.SwapFormatter(log.GetFormatter(o.LogFormatter))
				}
			}
			current = &sonnect{o.LogFormatter, o.LocalPath, o.SocketPath, o.Timeout}
			return flip.ExitNo
		},
		fs,
	)
}

func connect(s *sonnect, service, action string, d *data.Vector) flip.ExitStatus {
	req := cf.NewRequest(
		cf.ByteService(service),
		cf.ByteAction(action),
		d,
	)

	conn, err := connection(s.local, s.socket)
	defer cleanup(conn, s.local)
	if err != nil {
		return onError(err)
	}

	_, err = conn.Write(req.ToByte())
	if err != nil {
		return onError(err)
	}

	resp, err := response(conn, s.timeout)
	if err != nil {
		return onError(err)
	}

	L.Print(resp)

	return flip.ExitSuccess

	return flip.ExitSuccess
}

func onError(err error) flip.ExitStatus {
	L.Printf(err.Error())
	return flip.ExitFailure
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

var ResponseError = Crror("Error getting a response from the countfloyd server: %s").Out

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
		func(c context.Context, a []string) flip.ExitStatus {
			cs := []string{"-socket", current.socket, "-logFormatter", current.formatter}
			cs = getStartPopulate(cs, o)
			cs = append(cs, "start")
			cmd := exec.Command("cfs", cs...)
			cmd.Stdout = os.Stdout
			err := cmd.Start()
			if err != nil {
				return flip.ExitFailure
			}
			return flip.ExitSuccess
		},
		fs,
	)
}

func StopCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("stop", flip.ContinueOnError)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"stop",
		"stop a countfloyd server",
		2,
		func(c context.Context, a []string) flip.ExitStatus {
			return connect(current, "system", "quit", nil)
		},
		fs,
	)
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
		func(c context.Context, a []string) flip.ExitStatus {
			return connect(current, "query", "status", nil)
		},
		fs,
	)
}

func QueryCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("query", flip.ContinueOnError)
		fs.StringVar(&o.qFeature, "feature", o.qFeature, "return information for this specified feature")
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"query",
		"query a countfloyd server for feature information",
		4,
		func(c context.Context, a []string) flip.ExitStatus {
			d := data.New("")
			d.Set(data.NewStringItem("query_feature", o.qFeature))
			return connect(current, "query", "feature", d)
		},
		fs,
	)
}

func populateData(o *Options) *data.Vector {
	d := data.New("")
	fs := data.NewStringsItem("features", o.files("features")...)
	cs := data.NewStringsItem("components", o.files("components")...)
	es := data.NewStringsItem("entities", o.files("entities")...)
	d.Set(fs, cs, es)
	return d
}

func PopulateCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("populate", flip.ContinueOnError)
		filesFlags(o, fs)
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"populate",
		"populate a countfloyd server with features from provided files.",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			d := populateData(o)
			return connect(current, "data", "populate_from_files", d)
		},
		fs,
	)
	return nil
}

type aSwitch struct {
	v map[string]bool
}

func newASwitch() *aSwitch {
	return &aSwitch{
		make(map[string]bool),
	}
}

func evalASwitch(o *Options, d *data.Vector) (string, *data.Vector, error) {
	s := newASwitch()
	return s.state(o, d)
}

var MoreThanAllowableError = Crror("Can only request one of features, component, components, entity: %v").Out

func (a *aSwitch) state(o *Options, d *data.Vector) (string, *data.Vector, error) {
	var action string
	if o.aFeatures != "" {
		a.v["features"] = true
		action = "apply_features"
		spl := strings.Split(o.aFeatures, ",")
		d.Set(data.NewStringsItem("meta.features", spl...))
	}
	if o.aComponent != "" {
		a.v["component"] = true
		action = "apply_component"
		d.Set(data.NewStringItem("meta.component", o.aComponent))
	}
	if o.aComponents != "" {
		a.v["components"] = true
		action = "apply_components"
		spl := strings.Split(o.aComponents, ",")
		d.Set(data.NewStringsItem("meta.components", spl...))
	}
	if o.aEntity != "" {
		a.v["entity"] = true
		action = "apply_entity"
		d.Set(data.NewStringItem("meta.entity", o.aEntity))
	}
	var acount []string
	for k, v := range a.v {
		if v {
			acount = append(acount, k)
		}
	}
	if len(acount) > 1 {
		return "", nil, MoreThanAllowableError(acount)
	}
	return action, d, nil
}

func ApplyCommand() flip.Command {
	o := NewOptions()
	fs := func(o *Options) *flip.FlagSet {
		fs := flip.NewFlagSet("apply", flip.ContinueOnError)
		fs.IntVar(&o.aNumber, "number", 0, "A number value for meta.number")
		fs.StringVar(&o.aFeatures, "features", "", "A comma delimited list of features to apply.")
		fs.StringVar(&o.aComponent, "component", "", "A specific component to apply.")
		fs.StringVar(&o.aComponents, "components", "", "A comma delimited list of components to apply.")
		fs.StringVar(&o.aEntity, "entity", "", "A specific entity to apply.")
		return fs
	}(o)
	return flip.NewCommand(
		"",
		"apply",
		"apply a set of features",
		2,
		func(c context.Context, a []string) flip.ExitStatus {
			d := data.New("")
			d.Set(data.NewIntItem("meta.number", o.aNumber))
			action, d, err := evalASwitch(o, d)
			if err != nil {
				L.Print(err)
				return flip.ExitUsageError
			}
			return connect(current, "data", action, d)
		},
		fs,
	)
}
