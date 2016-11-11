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
	start = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "start")
	start.Stdout = os.Stdout
	populate = exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "stdout", "populate", "-featuresFiles", filepath.Join(cd, "features.yaml"))
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
	//qb := new(bytes.Buffer)
	//query.Stdout = qb
	//query.Run()
	//qr := server.EmptyResponse()
	//json.Unmarshal(qb.Bytes(), qr)
	//if qr.Error == nil && qr.Data != nil {
	//	d := qr.Data
	//	fmt.Println("---------------------")
	//	fmt.Println("QUERY:SOCIETY-ORIENT")
	//	fmt.Println(d.ToString("apply"))
	//	v := d.ToStrings("values")
	//	for _, vi := range v {
	//		fmt.Println(vi)
	//	}
	//	fmt.Println("---------------------")
	//}
	//stat := make(map[string]int)
	//for i := 0; i <= 1000; i += 1 {
	//	b := new(bytes.Buffer)
	//	q := exec.Command("countfloyd", "-socket", socketPath, "-logFormatter", "raw", "apply", "-number", fmt.Sprintf("%d", i), "-features", "self-needs,society-orient")
	//	q.Stdout = b
	//	q.Run()
	//	r := server.EmptyResponse()
	//	err := json.Unmarshal(b.Bytes(), r)
	//	if err != nil {
	//		log.Printf("%s", err.Error())
	//	}
	//	sn := r.Data.ToString("SELF-NEEDS")
	//	stat[sn] = stat[sn] + 1
	//	rb, _ := r.Data.MarshalJSON()
	//	fmt.Println(string(rb))
	//	fmt.Println("---------------------")
	//	b.Reset()
	//}
	//spew.Dump(stat)
	stop.Run()
}
