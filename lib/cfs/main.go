package main

import (
	"context"
	"os"
	"path"

	"github.com/Laughs-In-Flowers/countfloyd/lib/env"
	"github.com/Laughs-In-Flowers/countfloyd/lib/server"
	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"

	_ "github.com/Laughs-In-Flowers/countfloyd/lib/feature/constructors_common"
)

var (
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
	F              flip.Flpr
	L              log.Logger
	E              env.Env
	S              *server.Server
)

func init() {
	n := path.Base(os.Args[0])
	log.SetFormatter("cfs_text", log.MakeTextFormatter(n))
	F = flip.New("cfs")
	L = log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
	E = env.Empty()
	S = server.New(
		server.SetLogger(L),
		server.SetFeatureEnvironment(E),
	)
	F.AddBuiltIn("version", versionPackage, versionTag, versionHash, versionDate).
		AddBuiltIn("help").
		SetGroup("top", -1, TopCommand()).
		SetGroup("run", 1, StartCommand())
}

func main() {
	ctx := context.Background()
	exit := F.Execute(ctx, os.Args)
	os.Exit(exit)
}
