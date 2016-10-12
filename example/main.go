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
	start = exec.Command("cfc", "start")
	populate = exec.Command("cfc", "-logFormatter", "stdout", "populate", "-dir", cd)
	populate.Stdout = os.Stdout
	stop = exec.Command("cfc", "stop")
}

func main() {
	start.Start()
	duration := time.Second * 2
	time.Sleep(duration)
	populate.Start()
	populate.Wait()
	for i := 0; i <= 100; i += 2 {
		q := exec.Command("cfc", "-logFormatter", "stdout", "apply", "-number", fmt.Sprintf("%d", i), "-features", "motivation")
		q.Stdout = os.Stdout
		q.Start()
		q.Wait()
	}
	stop.Run()
}
