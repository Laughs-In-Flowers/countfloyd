package server

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Laughs-In-Flowers/countfloyd/lib/env"
	"github.com/Laughs-In-Flowers/log"
)

type Server struct {
	Configuration
	*settings
	log.Logger
	env.Env
	*Listener
	interrupt chan os.Signal
	*Handlers
}

func New(c ...Config) *Server {
	s := &Server{
		settings:  &settings{},
		interrupt: make(chan os.Signal, 0),
		Handlers:  NewHandlers(localHandlers...),
	}

	signal.Notify(
		s.interrupt,
		//os.Interrupt,
		//os.Kill,
		//syscall.SIGINT,
		//syscall.SIGTERM,
		//syscall.SIGKILL,
	)

	s.Configuration = newConfiguration(s, c...)

	return s
}

type settings struct {
	SocketPath string
}

func newSettings() *settings {
	return &settings{"/tmp/countfloyd_0_0-socket"}
}

func (s *Server) Serve() {
	s.Print("serving....")

	go s.start()

	for {
		select {
		case sig := <-s.interrupt:
			s.SignalHandler(sig)
		}
	}
}

//var NoItemError = xrr.Xrr("No %s with the tag %s is available.").Out

func (s *Server) process(r []byte) []byte {
	req := request(r)
	fn, err := s.GetRequestedHandle(req)
	if fn != nil {
		return fn(s, req)
	}
	return ErrorResponse(err).ToByte()
}

func (s *Server) Close() {
	s.stop()
}

func (s *Server) Quit() {
	s.Print("exiting")
	s.Close()
	os.Remove(s.SocketPath)
	os.Exit(0)
}

func (s *Server) SignalHandler(sig os.Signal) {
	msg := fmt.Sprintf("received signal %v", sig)
	switch sig {
	case os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
		s.Print(msg)
		s.Quit()
	default:
		s.Print(msg)
	}
}

func init() {
	NullResponse, _ = json.Marshal(&Response{})
}
