package main

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/countfloyd/lib/server"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"

	_ "github.com/Laughs-In-Flowers/countfloyd/lib/constructor"
)

type Options struct {
	LogFormatter                  string
	Socket                        string
	Listeners                     int
	Pfeature, Pcomponent, Pentity string
	Run                           bool
}

func unpackFiles(f string) []string {
	var ret []string
	lf := strings.Split(f, ",")
	for _, f := range lf {
		ret = append(ret, f)
	}
	return ret
}

func topFlags(o *Options) *flip.FlagSet {
	fs := flip.NewFlagSet("", flip.ContinueOnError)
	fs.StringVar(&o.LogFormatter, "logFormatter", o.LogFormatter, "Sets the environment logger formatter.")
	fs.StringVar(&o.Socket, "socket", o.Socket, "Set the server socket path.")
	fs.IntVar(&o.Listeners, "listeners", o.Listeners, "Set the number of listeners at the server socket.")
	fs.StringVar(&o.Pfeature, "features", o.Pfeature, "Attempt to load features from specified files")
	fs.StringVar(&o.Pcomponent, "components", o.Pcomponent, "Attempt to load components from specified files")
	fs.StringVar(&o.Pentity, "entities", o.Pentity, "Attempt to load entities from specified files")
	return fs
}

type tex func(o *Options)

var txx []tex = []tex{
	func(o *Options) {
		if o.LogFormatter != "" {
			switch o.LogFormatter {
			case "text", "stdout":
				L.SwapFormatter(log.GetFormatter("cfs_text"))
			default:
				L.SwapFormatter(log.GetFormatter(o.LogFormatter))
			}
		}
	},
	func(o *Options) {
		if o.Socket != "" {
			S.Add(server.SetSocketPath(o.Socket))
		}
	},
	func(o *Options) {
		if o.Listeners != 0 {
			S.Add(server.Listeners(o.Listeners))
		}
	},
	func(o *Options) {
		switch {
		case o.Pfeature != "":
			fs := unpackFiles(o.Pfeature)
			S.Add(server.SetPopulateFeatures(fs...))
			fallthrough
		case o.Pcomponent != "":
			fs := unpackFiles(o.Pcomponent)
			S.Add(server.SetPopulateComponents(fs...))
			fallthrough
		case o.Pentity != "":
			fs := unpackFiles(o.Pentity)
			S.Add(server.SetPopulateEntities(fs...))
		}
	},
}

func TopCommand() flip.Command {
	o := &Options{}
	fs := topFlags(o)
	return flip.NewCommand(
		"",
		"cfs",
		`Top level flag usage.`,
		0,
		func(c context.Context, a []string) flip.ExitStatus {
			for _, tx := range txx {
				tx(o)
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

func StartCommand() flip.Command {
	return flip.NewCommand(
		"",
		"start",
		"start the server",
		1,
		func(c context.Context, a []string) flip.ExitStatus {
			err := S.Configure()

			if err != nil {
				L.Fatalf("configuration error: %s", err.Error())
				return flip.ExitFailure
			}

			S.Serve()

			return flip.ExitSuccess
		},
		flip.NewFlagSet("", flip.ContinueOnError),
	)
}

var (
	C *flip.Commander
	L log.Logger
	E feature.Env
	S *server.Server
)

func main() {
	ctx := context.Background()
	C.Execute(ctx, os.Args)
	os.Exit(0)
}

func init() {
	n := path.Base(os.Args[0])
	log.SetFormatter("cfs_text", log.MakeTextFormatter(n))
	C = flip.BaseWithVersion(versionPackage, versionTag, versionHash, versionDate)
	L = log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
	E = feature.Empty()
	S = server.New(
		server.SetLogger(L),
		server.SetFeatureEnvironment(E),
	)
	flip.RegisterGroup("top", -1, TopCommand())
	flip.RegisterGroup("run", 1, StartCommand())
}
