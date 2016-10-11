package main

import (
	"context"
	"os"

	"github.com/Laughs-In-Flowers/flip"
	"github.com/Laughs-In-Flowers/log"
)

var (
	C *flip.Commander
	L log.Logger
)

func init() {
	C = flip.BaseWithVersion(versionPackage, versionTag, versionHash, versionDate)
	L = log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
	flip.RegisterGroup("top", -1, TopCommand())
	flip.RegisterGroup("control", 1, StartCommand(), StopCommand())
	flip.RegisterGroup("action", 2, PopulateCommand(), ApplyCommand())
}

func main() {
	ctx := context.Background()
	C.Execute(ctx, os.Args)
	os.Exit(0)
}
