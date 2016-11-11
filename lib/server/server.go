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

type settings struct {
	SocketPath string
	Listeners  int
}

func newSettings() *settings {
	return &settings{
		"/tmp/countfloyd_0_0-socket", 5,
	}
}

type Server struct {
	Configuration
	*settings
	log.Logger
	feature.Env
	Listening []*Listener
	interrupt chan os.Signal
	*Handlers
}

type ProcessFunc func([]byte) []byte

type Listener struct {
	Error error
	*net.UnixListener
	process ProcessFunc
	stop    chan struct{}
}

func NewListener(socket string, fn ProcessFunc) *Listener {
	l, err := net.ListenUnix("unix", &net.UnixAddr{socket, "unix"})
	st := make(chan struct{}, 0)
	return &Listener{
		err, l, fn, st,
	}
}

func heard(c *net.UnixConn, l func([]byte) []byte) {
	var buf [1024]byte
	n, err := c.Read(buf[:])
	if err != nil {
		panic(err)
	}
	req := bytes.Trim(buf[:n], " ")
	resp := l(req)
	c.Write(resp)
	c.Close()
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
			go heard(conn, l.process)
		}
	}
}

func (l *Listener) quit() {
	l.Close()
	l.stop <- struct{}{}
}

func New(c ...Config) *Server {
	s := &Server{
		settings:  &settings{},
		Listening: make([]*Listener, 0),
		interrupt: make(chan os.Signal, 0),
		Handlers:  NewHandlers(localHandlers...),
	}

	signal.Notify(
		s.interrupt,
		os.Interrupt,
		os.Kill,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)

	s.Configuration = newConfiguration(s, c...)

	return s
}

func (s *Server) Serve() {
	s.Print("serving....")

	for _, v := range s.Listening {
		go v.listen()
	}

	for {
		select {
		case sig := <-s.interrupt:
			s.SignalHandler(sig)
		}
	}
}

var NoFeatureError = Srror("No feature named %s available.").Out

var (
	NoServiceError     = Srror("no service named %s available").Out
	UnknownActionError = Srror("unknown action %s").Out
)

type Handler struct {
	Service, Action string
	Fn              HandlerFunc
}

func NewHandler(service, action string, fn HandlerFunc) *Handler {
	return &Handler{
		service, action, fn,
	}
}

type HandlerFunc func(*Server, *Request) []byte

type hActions map[string]HandlerFunc

type hServices map[string]hActions

type Handlers struct {
	has hServices
}

func NewHandlers(hn ...*Handler) *Handlers {
	h := &Handlers{
		has: make(hServices),
	}
	for _, v := range services {
		h.has[v.String()] = make(hActions)
	}
	h.SetHandle(hn...)
	return h
}

func (h *Handlers) GetRequestedHandle(r *Request) (HandlerFunc, error) {
	var sv, ac string
	sv = r.Service.String()
	ac = r.Action.String()
	return h.GetHandle(sv, ac)
}

func (h *Handlers) GetHandle(service, action string) (HandlerFunc, error) {
	if sm, ok := h.has[service]; ok {
		if fn, ok := sm[action]; ok {
			return fn, nil
		}
		return nil, UnknownActionError(action)
	}
	return nil, NoServiceError(service)
}

func (h *Handlers) SetHandle(hs ...*Handler) {
	for _, hn := range hs {
		if isService(ByteService(hn.Service)) {
			if service, ok := h.has[hn.Service]; ok {
				if !isAction([]byte(hn.Action)) {
					actions = append(actions, []byte(hn.Action))
				}
				service[hn.Action] = hn.Fn
			}
		}
	}
}

func NewDataFrom(m *data.Vector, e feature.Env) *data.Vector {
	n := m.ToInt("meta.number")
	d := feature.NewData(n)
	a := m.ToStrings("meta.features")
	e.Apply(a, d)
	return d
}

var localHandlers []*Handler = []*Handler{
	NewHandler(
		"system",
		"ping",
		func(*Server, *Request) []byte {
			return NullResponse
		}),
	NewHandler(
		"system",
		"quit",
		func(s *Server, r *Request) []byte {
			s.Quit()
			return nil
		}),
	NewHandler(
		"query",
		"status",
		func(s *Server, r *Request) []byte {
			resp := StatusResponse(s)
			return resp.ToByte()
		}),
	NewHandler(
		"query",
		"feature",
		func(s *Server, r *Request) []byte {
			resp := EmptyResponse()
			d := r.Data
			qf := d.ToString("query_feature")
			f := s.GetFeature(qf)
			if f == nil {
				resp.Error = NoFeatureError(qf)
			}
			if f != nil {
				si := data.NewStringsItem("set", f.Group()...)
				fi := data.NewStringItem("apply", f.From())
				vi := data.NewStringsItem("values", f.Values()...)
				d.Set(si, fi, vi)
				resp.Data = d
			}
			return resp.ToByte()
		}),
	NewHandler(
		"data",
		"populate_from_files",
		func(s *Server, r *Request) []byte {
			resp := EmptyResponse()
			d := r.Data
			files := d.ToStrings("files")
			resp.Error = s.PopulateYamlFiles(files...)
			return resp.ToByte()
		}),
	NewHandler(
		"data",
		"apply",
		func(s *Server, r *Request) []byte {
			resp := EmptyResponse()
			d := r.Data
			resp.Data = NewDataFrom(d, s)
			return resp.ToByte()
		}),
	NewHandler(
		"data",
		"apply_to_file",
		func(s *Server, r *Request) []byte {
			resp := EmptyResponse()
			d := r.Data
			path := d.ToString("file")

			fileData := func(path string) (*os.File, []byte, error) {
				fl, err := data.Open(path)
				if err != nil {
					return nil, nil, err
				}
				var n int64
				if fi, err := fl.Stat(); err == nil {
					if size := fi.Size(); size < 1e9 {
						n = size
					}
				}
				b := make([]byte, n)
				_, err = fl.Read(b)
				if err != nil {
					return nil, nil, err
				}
				return fl, b, nil
			}

			fl, b, err := fileData(path)
			if err != nil {
				resp.Error = err
			}

			err = d.UnmarshalJSON(b)
			if err != nil {
				resp.Error = err
			}

			existing := d.Clone("file", "action")

			resp.Data = NewDataFrom(existing, s)

			existing.Merge(resp.Data)

			var afb []byte
			afb, err = existing.MarshalJSON()
			if err != nil {
				resp.Error = err
			}

			fl.Truncate(0)
			fl.WriteAt(afb, 0)
			fl.Sync()
			return resp.ToByte()
		}),
}

func (s *Server) process(r []byte) []byte {
	req := request(r)
	fn, err := s.GetRequestedHandle(req)
	if fn != nil {
		return fn(s, req)
	}
	return ErrorResponse(err).ToByte()
}

func (s *Server) Close() {
	for _, l := range s.Listening {
		l.quit()
	}
}

func (s *Server) Quit() {
	s.Print("exiting")
	s.Close()
	os.Remove(s.SocketPath)
	os.Exit(0)
}

func (s *Server) SignalHandler(sig os.Signal) {
	s.Printf("received signal %v", sig)
	switch sig {
	case os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
		s.Quit()
	}
}

func init() {
	NullResponse, _ = json.Marshal(&Response{})
}
