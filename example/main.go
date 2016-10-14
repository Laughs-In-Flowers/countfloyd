package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var socketPath string = "/tmp/custom_countfloyd_socket_0"

var start, populate, status, query, stop *exec.Cmd

func init() {
	cd, _ := os.Getwd()
	start = exec.Command("countfloyd", "-socket", socketPath, "start")
	start.Stdout = os.Stdout
	populate = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "populate", "-featuresDir", cd)
	populate.Stdout = os.Stdout
	status = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "status")
	status.Stdout = os.Stdout
	query = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "query", "-feature", "ORIENTATION-CUE")
	query.Stdout = os.Stdout
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
	query.Start()
	query.Wait()
	for i := 0; i <= 100; i += 2 {
		q := exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "apply", "-number", fmt.Sprintf("%d", i), "-features", "motivation")
		q.Stdout = os.Stdout
		q.Start()
		q.Wait()
	}
	stop.Run()
}
