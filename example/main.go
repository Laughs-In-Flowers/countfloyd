package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var start, populate, stop *exec.Cmd

func init() {
	cd, _ := os.Getwd()
	start = exec.Command("countfloyd", "start")
	populate = exec.Command("countfloyd", "-logFormatter", "stdout", "populate", "-featuresDir", cd)
	populate.Stdout = os.Stdout
	stop = exec.Command("countfloyd", "stop")
}

func main() {
	start.Start()
	duration := time.Second * 2
	time.Sleep(duration)
	populate.Start()
	populate.Wait()
	for i := 0; i <= 100; i += 2 {
		q := exec.Command("countfloyd", "-logFormatter", "stdout", "apply", "-number", fmt.Sprintf("%d", i), "-features", "motivation")
		q.Stdout = os.Stdout
		q.Start()
		q.Wait()
	}
	stop.Run()
}
