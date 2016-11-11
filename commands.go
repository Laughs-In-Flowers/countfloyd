package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
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
	LogFormatter          string
	LocalPath, SocketPath string
	Timeout               time.Duration
	fromFiles             bool
	dir, files            string
	MetaNumber            int
	MetaFeatures          string
	applyToFile           bool
	applyToFileName       string
	QueryFeature          string
}

func NewOptions() *Options {
	return &Options{
		"",
		"/tmp/cfc",
		"/tmp/countfloyd_0_0-socket",
		2 * time.Second,
		false,
		"",
		"",
		0,
		"",
		false,
		"",
		"",
	}
}

func (o *Options) PopulateAction() {
	switch {
	case o.dir != "" || o.files != "":
		o.fromFiles = true
	}
}

func filesFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.dir, "featuresDir", "", "A directory to locate feature files in.")
	fs.StringVar(&o.files, "featuresFiles", "", "A comma delimited list of feature files.")
}

func socketFlags(o *Options, fs *flip.FlagSet) {
	fs.StringVar(&o.LocalPath, "local", o.LocalPath, "Specify a local path for communication to the server.")
	fs.StringVar(&o.SocketPath, "socket", o.SocketPath, "Specify the socket path of the server.")
}

func (o *Options) Files() []string {
	var ret []string

	if o.dir != "" {
		df, err := ioutil.ReadDir(o.dir)
		if err != nil {
			L.Fatal(err.Error())
		}

		for _, f := range df {
			ret = append(ret, filepath.Join(o.dir, f.Name()))
		}
	}

	if o.files != "" {
		lf := strings.Split(o.files, ",")
		for _, f := range lf {
			ret = append(ret, f)
		}
	}

	return ret
}

func (o *Options) FilesString() string {
	fs := o.Files()
	return strings.Join(fs, ",")
}

func topFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("", flip.ContinueOnError)
	fs.StringVar(&o.LogFormatter, "logFormatter", o.LogFormatter, "Sets the environment logger formatter.")
	socketFlags(o, fs)
	return fs
}

const topUse = `Top level flag usage.`

func TopCommand() flip.Command {
	o := NewOptions()
	fs := topFlags(o)
	return flip.NewCommand(
		"",
		"countfloyd",
		topUse,
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

var (
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
)

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

func startFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("", flip.ContinueOnError)
	return fs
}

func StartCommand() flip.Command {
	o := NewOptions()
	fs := startFlags(o)
	filesFlags(o, fs)
	return flip.NewCommand(
		"",
		"start",
		"start a countfloyd server",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			cs := []string{"-socket", current.socket, "-logFormatter", current.formatter}
			if ff := o.FilesString(); ff != "" {
				cs = append(cs, "-populateFiles", ff)
			}
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

func stopFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("stop", flip.ContinueOnError)
	return fs
}

func StopCommand() flip.Command {
	o := NewOptions()
	fs := stopFlags(o)
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

func statusFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("status", flip.ContinueOnError)
	return fs
}

func StatusCommand() flip.Command {
	o := NewOptions()
	fs := statusFlags(o)
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

func queryFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("query", flip.ContinueOnError)
	fs.StringVar(&o.QueryFeature, "feature", o.QueryFeature, "return information for this specified feature")
	return fs
}

func QueryCommand() flip.Command {
	o := NewOptions()
	fs := queryFlags(o)
	return flip.NewCommand(
		"",
		"query",
		"query a countfloyd server for feature information",
		4,
		func(c context.Context, a []string) flip.ExitStatus {
			d := data.New("")
			d.Set(data.NewStringItem("query_feature", o.QueryFeature))
			return connect(current, "query", "feature", d)
		},
		fs,
	)
}

func populateFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("populate", flip.ContinueOnError)
	filesFlags(o, fs)
	return fs
}

func PopulateCommand() flip.Command {
	o := NewOptions()
	fs := populateFlags(o)
	return flip.NewCommand(
		"",
		"populate",
		"populate a countfloyd server with features from provided files.",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			o.PopulateAction()
			switch {
			case o.fromFiles:
				d := data.New("")
				d.Set(data.NewStringsItem("files", o.Files()...))
				return connect(current, "data", "populate_from_files", d)
			}
			return flip.ExitUsageError
		},
		fs,
	)
	return nil
}

func applyFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("apply", flip.ContinueOnError)
	fs.IntVar(&o.MetaNumber, "number", 0, "A number value for meta.number")
	fs.StringVar(&o.MetaFeatures, "features", "", "A comma delimited list of features to apply.")
	fs.BoolVar(&o.applyToFile, "file", false, "Apply features to a file from data in that file.")
	fs.StringVar(&o.applyToFileName, "name", "", "The file name to read for data, as well as write out applied feature information.")
	socketFlags(o, fs)
	return fs
}

func ApplyCommand() flip.Command {
	o := NewOptions()
	fs := applyFlags(o)
	return flip.NewCommand(
		"",
		"apply",
		"apply a set of features",
		2,
		func(c context.Context, a []string) flip.ExitStatus {
			d := data.New("")
			var action string
			switch {
			case o.applyToFile:
				d.Set(data.NewStringItem("file", o.applyToFileName))
				action = "apply_to_file"
			default:
				d.Set(data.NewIntItem("meta.number", o.MetaNumber))
				spl := strings.Split(o.MetaFeatures, ",")
				d.Set(data.NewStringsItem("meta.features", spl...))
				action = "apply"
			}
			return connect(current, "data", action, d)
		},
		fs,
	)
}
