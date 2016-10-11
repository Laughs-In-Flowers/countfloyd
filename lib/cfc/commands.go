package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cf "github.com/Laughs-In-Flowers/countfloyd/lib/server"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
)

type Options struct {
	LogFormatter          string
	LocalPath, SocketPath string
	Timeout               time.Duration
	fromFiles             bool
	dir, files            string
	MetaNumber            int
	MetaFeatures          string
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
	}
}

func (o *Options) PopulateAction() {
	switch {
	case o.dir != "" || o.files != "":
		o.fromFiles = true
	}
}

func (o *Options) Files() []string {
	var ret []string

	if o.dir == "" {
		o.dir = filepath.Join(currentDir, "features")
	}

	df, err := ioutil.ReadDir(o.dir)
	if err != nil {
		L.Fatal(err.Error())
	}

	for _, f := range df {
		ret = append(ret, filepath.Join(o.dir, f.Name()))
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
	return fs
}

const topUse = `Top level flag usage.`

func TopCommand() flip.Command {
	o := NewOptions()
	fs := topFlags(o)
	return flip.NewCommand(
		"",
		"cfc",
		topUse,
		0,
		func(c context.Context, a []string) flip.ExitStatus {
			if o.LogFormatter != "" {
				switch o.LogFormatter {
				case "text", "stdout":
					L.SwapFormatter(log.GetFormatter("cfc_text"))
				}
				L.Printf("swapping to log formatter: %s", o.LogFormatter)
			}
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

func connectByte(o *Options, b []byte) flip.ExitStatus {
	conn, err := connection(o.LocalPath, o.SocketPath)
	defer cleanup(conn, o)
	if err != nil {
		return onError(err)
	}

	_, err = conn.Write(b)
	if err != nil {
		return onError(err)
	}

	resp, err := response(conn, o.Timeout)
	if err != nil {
		return onError(err)
	}

	L.Print(resp)

	return flip.ExitSuccess
}

func connectData(o *Options, d *data.Container) flip.ExitStatus {
	b, err := d.MarshalJSON()
	if err != nil {
		return onError(err)
	}

	return connectByte(o, b)
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

var ResponseError = Crror("Response error: %s").Out

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

func cleanup(c *net.UnixConn, o *Options) {
	if c != nil {
		c.Close()
	}
	os.Remove(o.LocalPath)
}

func StartCommand() flip.Command {
	return flip.NewCommand(
		"",
		"start",
		"start a verbose countfloyd server",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			cmd := exec.Command("countfloyd", "-logFormatter", "stdout", "start")
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				return flip.ExitFailure
			}
			return flip.ExitSuccess
		},
		flip.NewFlagSet("", flip.ContinueOnError),
	)
}

func StopCommand() flip.Command {
	o := NewOptions()
	return flip.NewCommand(
		"",
		"stop",
		"stop a countfloyd server",
		2,
		func(c context.Context, a []string) flip.ExitStatus {
			return connectByte(o, cf.QUIT)
		},
		flip.NewFlagSet("populate", flip.ContinueOnError),
	)
}

func populateFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("populate", flip.ContinueOnError)
	fs.StringVar(&o.dir, "dir", "", "A directory to locate the provide files in.")
	fs.StringVar(&o.files, "files", "", "")
	return fs
}

func PopulateCommand() flip.Command {
	o := NewOptions()
	fs := populateFlags(o)
	return flip.NewCommand(
		"",
		"populate",
		"populate a countfloyd server with features and/or constructors",
		10,
		func(c context.Context, a []string) flip.ExitStatus {
			o.PopulateAction()
			switch {
			case o.fromFiles:
				d := data.NewContainer("")
				d.Set(data.NewItem("action", "populate_from_files"))
				d.Set(data.NewItem("files", o.FilesString()))
				return connectData(o, d)
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
	return fs
}

func ApplyCommand() flip.Command {
	o := NewOptions()
	fs := applyFlags(o)
	return flip.NewCommand(
		"",
		"apply",
		"apply a set of features",
		20,
		func(c context.Context, a []string) flip.ExitStatus {
			d := data.NewContainer("")
			d.Set(data.NewItem("action", "apply"))
			d.Set(data.NewItem("meta.number", strconv.Itoa(o.MetaNumber)))
			d.Set(data.NewItem("meta.features", o.MetaFeatures))
			return connectData(o, d)
		},
		fs,
	)
}

var currentDir string

func init() {
	currentDir, _ = os.Getwd()
	n := fmt.Sprintf("%s", path.Base(os.Args[0]))
	log.SetFormatter("cfc_text", log.MakeTextFormatter(n))
}
