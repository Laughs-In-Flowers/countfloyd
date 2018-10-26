package main

import (
	"context"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/server"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
)

type Options struct {
	LogFormatter                  string
	Socket                        string
	PGroups                       string
	Pfeature, Pcomponent, Pentity string
	PPcomponent, PPfeature        string
}

func unpackToStrings(f string) []string {
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
	fs.StringVar(&o.Pfeature, "features", o.Pfeature, "Attempt to load features from specified files")
	fs.StringVar(&o.Pcomponent, "components", o.Pcomponent, "Attempt to load components from specified files")
	fs.StringVar(&o.Pentity, "entities", o.Pentity, "Attempt to load entities from specified files")
	fs.StringVar(&o.PPcomponent, "componentPlugin", o.PPcomponent, "Attempt to load constructor plugin from dirs")
	fs.StringVar(&o.PPfeature, "featurePlugin", o.PPfeature, "Attempt to load feature plugin from dirs")
	fs.StringVar(&o.PGroups, "groups", o.PGroups, "Groups parameter applied where features, components, or entities are populated.")
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
		if o.Pfeature != "" {
			S.Add(server.SetPopulateFeatures(
				unpackToStrings(o.PGroups),
				unpackToStrings(o.Pfeature)...))
		}
		if o.Pcomponent != "" {
			S.Add(server.SetPopulateComponents(
				unpackToStrings(o.PGroups),
				unpackToStrings(o.Pcomponent)...))
		}
		if o.Pentity != "" {
			S.Add(server.SetPopulateEntities(
				unpackToStrings(o.PGroups),
				unpackToStrings(o.Pentity)...))
		}
		if o.PPcomponent != "" {
			S.Add(server.SetConstructorPluginDirs(
				unpackToStrings(o.PPcomponent)...))
		}
		if o.PPfeature != "" {
			S.Add(server.SetFeaturePluginDirs(
				unpackToStrings(o.PGroups),
				unpackToStrings(o.PPfeature)...))
		}
	},
}

func TopCommand() flip.Command {
	o := &Options{}
	fs := topFlags(o)
	return flip.NewCommand(
		"",
		"cfs: countfloyd server starter",
		`Top level flag usage.`,
		0,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			for _, tx := range txx {
				tx(o)
			}
			return c, flip.ExitNo
		},
		fs,
	)
}

func StartCommand() flip.Command {
	return flip.NewCommand(
		"",
		"start",
		"start the server",
		1,
		false,
		func(c context.Context, a []string) (context.Context, flip.ExitStatus) {
			err := S.Configure()

			if err != nil {
				L.Fatalf("configuration error: %s", err.Error())
				return c, flip.ExitFailure
			}

			S.Serve()

			return c, flip.ExitSuccess
		},
		flip.NewFlagSet("", flip.ContinueOnError),
	)
}
