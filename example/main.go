package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Laughs-In-Flowers/countfloyd/lib/server"
)

var socketPath string = "/tmp/custom_countfloyd_socket_0"

var start, populate, status, query, stop *exec.Cmd

func init() {
	cd, _ := os.Getwd()
	start = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "start")
	start.Stdout = os.Stdout
	populate = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "populate", "-featuresDir", cd)
	populate.Stdout = os.Stdout
	status = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "status")
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
	qb := new(bytes.Buffer)
	query.Stdout = qb
	query.Run()
	qr := server.EmptyResponse()
	json.Unmarshal(qb.Bytes(), qr)
	if qr.Error == nil && qr.Data != nil {
		d := qr.Data
		fmt.Println("---------------------")
		fmt.Println("QUERY:SOCIETY-ORIENT")
		fmt.Println(d.ToString("apply"))
		v := d.ToString("values")
		vs := strings.Split(v, ",")
		for _, vi := range vs {
			fmt.Println(vi)
		}
		fmt.Println("---------------------")
	} else {
		fmt.Println(qr.Error.Error())
	}
	//for i := 0; i <= 100; i += 1 {
	//	b := new(bytes.Buffer)
	//	q := exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "raw", "apply", "-number", fmt.Sprintf("%d", i), "-features", "motivation")
	//	q.Stdout = b
	//	q.Run()
	//	r := server.EmptyResponse()
	//	err := json.Unmarshal(b.Bytes(), r)
	//	if err != nil {
	//		log.Printf("%s", err.Error())
	//	}
	//	fmt.Println(r.Data)
	//	fmt.Println("---------------------")
	//	b.Reset()
	//}
	stop.Run()
}
