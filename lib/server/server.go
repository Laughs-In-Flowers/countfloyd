package server

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/ifriit/lib/file"
	"github.com/Laughs-In-Flowers/log"
)

type Request struct {
	Service Service
	Action  Action
	Data    *data.Container
}

type Service []byte

func (s Service) String() string {
	return string(s)
}

func (s Service) Action() Action {
	return StringToAction(s.String())
}

var (
	NONE   = []byte("none")
	PING   = []byte("ping")
	STATUS = []byte("status")
	QUERY  = []byte("query")
	DATA   = []byte("data")
	QUIT   = []byte("quit")

	services []Service = []Service{
		NONE, PING, STATUS, QUERY, DATA, QUIT,
	}

	directServices []Service = []Service{
		PING, STATUS, QUIT,
	}
)

func servicesString() string {
	var l []string
	for _, v := range services {
		l = append(l, v.String())
	}
	return strings.Join(l, ",")
}

type Action int

const (
	Unknown Action = iota
	Ping
	Status
	QueryFeature
	PopulateFromFiles
	Apply
	ApplyToFile
	Quit
)

var actions []Action = []Action{
	Unknown, Ping, Status, QueryFeature, PopulateFromFiles, Apply, Quit,
}

func actionsString() string {
	var l []string
	for _, v := range actions {
		l = append(l, v.String())
	}
	return strings.Join(l, ",")
}

func StringToAction(s string) Action {
	switch s {
	case "ping":
		return Ping
	case "status":
		return Status
	case "query_feature":
		return QueryFeature
	case "populate_from_files":
		return PopulateFromFiles
	case "apply":
		return Apply
	case "apply_to_file":
		return ApplyToFile
	case "quit":
		return Quit
	}
	return Unknown
}

func (a Action) String() string {
	switch a {
	case Ping:
		return "ping"
	case Status:
		return "status"
	case QueryFeature:
		return "query_feature"
	case PopulateFromFiles:
		return "populate_from_files"
	case Apply:
		return "apply"
	case ApplyToFile:
		return "apply_to_file"
	case Quit:
		return "quit"
	}
	return "unknown"
}

func request(req []byte) *Request {
	switch {
	case directService(req):
		return directRequest(req)
	case featureQuery(req):
		return featureRequest(req)
	default:
		return dataRequest(req)
	}
	return &Request{NONE, Unknown, nil}
}

func directRequest(s Service) *Request {
	return &Request{
		Service: s,
		Action:  s.Action(),
	}
}

func directService(r []byte) bool {
	for _, v := range directServices {
		if bytes.Compare(r, v) == 0 {
			return true
		}
	}
	return false
}

func featureQuery(r []byte) bool {
	return bytes.Contains(r, QUERY)
}

func featureRequest(r []byte) *Request {
	fd := bytes.Fields(r)
	var fss string
	if len(fd) > 1 {
		fs := fd[1]
		fss = string(fs)
	}
	d := data.NewContainer("")
	d.Set(data.NewItem("tag", fss))
	return &Request{
		Service: QUERY,
		Action:  QueryFeature,
		Data:    d,
	}
}

func dataRequest(r []byte) *Request {
	d := data.NewContainer("")
	err := json.Unmarshal(r, &d)
	if err != nil {
		d.Set(data.NewItem("Error", err.Error()))
	}
	var a Action
	if ac := d.Get("action"); ac != nil {
		astr := ac.ToString()
		a = StringToAction(astr)
	}
	return &Request{
		Service: DATA,
		Action:  a,
		Data:    d,
	}
}

type Response struct {
	Error error
	Data  *data.Container
}

var EmptyResponse Response = Response{nil, data.NewContainer("")}

func (s *Server) StatusResponse() *Response {
	d := data.NewContainer("")

	d.Set(data.NewItem("socket", s.SocketPath))
	d.Set(data.NewItem("services", servicesString()))
	d.Set(data.NewItem("actions", actionsString()))
	d.Set(data.NewItem("features", strings.Join(s.ListKeys(""), ",")))

	return &Response{
		nil, d,
	}
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
		syscall.SIGKILL,
	)

	s.Configuration = newConfiguration(s, c...)

	return s
}

func (s *Server) Serve() {
	s.Print("serving....")

	go s.listen()

	for {
		select {
		case sig := <-s.interrupt:
			s.SignalHandler(sig)
		}
	}
}

var NoFeatureError = Srror("No feature named %s available.").Out

func fileData(path string) (*os.File, []byte, error) {
	fl, err := file.Open(path)
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

func (s *Server) process(r []byte) []byte {
	req := request(r)
	d := req.Data
	resp := &Response{}

	switch req.Action {
	case Ping:
		return NullResponse
	case Status:
		resp = s.StatusResponse()
	case Quit:
		s.Stop()
	case QueryFeature:
		tag := d.ToString("tag")
		f := s.GetFeature(tag)
		if f == nil {
			resp.Error = NoFeatureError(tag)
		}
		if f != nil {
			si := data.NewItem("set", "")
			si.SetList(f.Group()...)
			d.Set(si)
			d.Set(data.NewItem("apply", f.From()))
			vi := data.NewItem("values", "")
			vi.SetList(f.Values()...)
			d.Set(vi)
			resp.Data = d
		}
	case PopulateFromFiles:
		files := d.ToList("files")
		resp.Error = s.PopulateYamlFiles(files...)
	case Apply:
		resp.Data = feature.DataFrom(d, s)
	case ApplyToFile:
		path := d.ToString("file")

		fl, b, err := fileData(path)
		if err != nil {
			resp.Error = err
		}

		err = d.UnmarshalJSON(b)
		if err != nil {
			resp.Error = err
		}

		existing := d.Clone("file", "action")

		resp.Data = feature.DataFrom(existing, s)

		wo := data.Merge(existing, resp.Data)

		var afb []byte
		afb, err = wo.MarshalJSON()
		if err != nil {
			resp.Error = err
		}

		fl.Truncate(0)
		fl.WriteAt(afb, 0)
		fl.Sync()
	}

	rb, _ := json.Marshal(&resp)

	return rb
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
	case os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
		s.Stop()
	}
}

var NullResponse []byte

func init() {
	NullResponse, _ = json.Marshal(&Response{})
}
