package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var socketPath string = "/tmp/custom_countfloyd_socket_0"

var start, populate, status, query, stop *exec.Cmd

func init() {
	cd, _ := os.Getwd()
	start = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "raw", "start")
	start.Stdout = os.Stdout
	populate = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "raw", "populate", "-feature", filepath.Join(cd, "features.yaml"))
	populate.Stdout = os.Stdout
	status = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "raw", "status")
	status.Stdout = os.Stdout
	query = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "raw", "query", "-feature", "SOCIETY-ORIENT")
	stop = exec.Command("countfloyd", "-socket", socketPath, "stop")
	stop.Stdout = os.Stdout
}

func main() {
	start.Start()
	duration := time.Second * 2
	time.Sleep(duration)
	populate.Start()
	populate.Wait()
	status.Start()
	status.Wait()
	stop.Run()
}
