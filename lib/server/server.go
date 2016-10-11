package server

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/log"
)

type Request struct {
	Action Action
	Data   *data.Container
}

type Action int

const (
	Unknown Action = iota
	Populate
	PopulateFromFiles
	Apply
)

func StringToAction(s string) Action {
	switch s {
	case "populate":
		return Populate
	case "populate_from_files":
		return PopulateFromFiles
	case "apply":
		return Apply
	}
	return Unknown
}

func NewRequest(req []byte) *Request {
	d := data.NewContainer("")
	err := json.Unmarshal(req, &d)
	if err != nil {
		d.Set(data.NewItem("Error", err.Error()))
	}
	var a Action
	if ac := d.Get("action"); ac != nil {
		astr := ac.ToString()
		a = StringToAction(astr)
	}
	return &Request{
		Action: a,
		Data:   d,
	}
}

type Response struct {
	Error error
	Data  *data.Container
}

type settings struct {
	SocketPath string
}

func newSettings() *settings {
	return &settings{
		"/tmp/countfloyd_0_0-socket",
	}
}

type Server struct {
	Configuration
	*settings
	log.Logger
	feature.Env
	*Listener
	interrupt chan os.Signal
}

type Listener struct {
	*net.UnixListener
	process func([]byte) []byte
	stop    chan struct{}
}

func (l *Listener) listen() {
LISTEN:
	for {
		select {
		case <-l.stop:
			break LISTEN
		default:
			conn, err := l.AcceptUnix()
			if err != nil {
				panic(err)
			}
			var buf [1024]byte
			n, err := conn.Read(buf[:])
			if err != nil {
				panic(err)
			}
			req := bytes.Trim(buf[:n], " ")
			resp := l.process(req)
			conn.Write(resp)
			conn.Close()
		}
	}
}

func (l *Listener) stopListening() {
	l.stop <- struct{}{}
}

func New(c ...Config) *Server {
	s := &Server{
		settings:  &settings{},
		interrupt: make(chan os.Signal, 0),
	}

	signal.Notify(
		s.interrupt,
		os.Interrupt,
		os.Kill,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	s.Configuration = newConfiguration(s, c...)

	return s
}

func (s *Server) Serve() {
	go s.listen()

	for {
		select {
		case sig := <-s.interrupt:
			s.SignalHandler(sig)
		}
	}
}

var QUIT = []byte("QUIT")

func isQuit(r []byte) bool {
	if v := bytes.Compare(r, QUIT); v == 0 {
		return true
	}
	return false
}

func (s *Server) process(r []byte) []byte {
	switch {
	case isQuit(r):
		s.Stop()
	default:
		req := NewRequest(r)
		d := req.Data
		resp := &Response{}

		switch req.Action {
		case PopulateFromFiles:
			files := d.ToList("files")
			resp.Error = s.PopulateYamlFiles(files...)
		case Apply:
			resp.Data = feature.DataFrom(d, s)
		}

		rb, _ := json.Marshal(&resp)

		return rb
	}
	return NullResponse
}

func (s *Server) Stop() {
	s.Print("exiting")
	s.Close()
	os.Remove(s.SocketPath)
	os.Exit(0)
}

func (s *Server) SignalHandler(sig os.Signal) {
	s.Printf("received signal %v", sig)
	switch sig {
	case os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM:
		s.Stop()
	}
}

var NullResponse []byte

func init() {
	NullResponse, _ = json.Marshal(&Response{})
}
