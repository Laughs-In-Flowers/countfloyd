package server

import (
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/log"
)

type ConfigFn func(*Server) error

type Config interface {
	Order() int
	Configure(*Server) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure(e *Server) error {
	return c.fn(e)
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	s          *Server
	configured bool
	list       configList
}

func newConfiguration(s *Server, conf ...Config) *configuration {
	c := &configuration{
		s:    s,
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure(e *Server, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.s, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1000, sLogger},
	config{1001, sSocketPath},
	config{1002, sListener},
	config{1003, sFeatureEnv},
}

func SetLogger(l log.Logger) Config {
	return DefaultConfig(func(s *Server) error {
		s.Logger = l
		return nil
	})
}

func sLogger(s *Server) error {
	if s.Logger == nil {
		s.Logger = log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
	}
	return nil
}

func SetSocketPath(p string) Config {
	return DefaultConfig(func(s *Server) error {
		s.SocketPath = p
		return nil
	})
}

func sSocketPath(s *Server) error {
	if s.SocketPath == "" {
		s.SocketPath = "/tmp/countfloyd_0_0-socket"
	}
	return nil
}

func sListener(s *Server) error {
	if len(s.Listening) == 0 {
		for i := 0; i <= s.Listeners; i++ {
			lr := NewListener(s.SocketPath, s.process)
			if lr.Error != nil {
				return lr.Error
			}
			s.Listening = append(s.Listening, lr)
		}
	}
	return nil
}

func Listeners(n int) Config {
	return DefaultConfig(func(s *Server) error {
		s.Listeners = n
		return nil
	})
}

func SetFeatureEnvironment(f feature.Env) Config {
	return DefaultConfig(func(s *Server) error {
		s.Env = f
		return nil
	})
}

func sFeatureEnv(s *Server) error {
	if s.Env == nil {
		s.Env = feature.Empty()
	}
	return nil
}

func SetPopulateFeatures(files ...string) Config {
	return NewConfig(2000, func(s *Server) error {
		for _, f := range files {
			s.Printf("loading features from %s", f)
		}
		return s.PopulateYaml(files...)
	})
}

func SetPopulateComponents(files ...string) Config {
	return NewConfig(2001, func(s *Server) error {
		for _, f := range files {
			s.Printf("loading features from %s", f)
		}
		return s.PopulateComponentYaml(files...)
	})
}

func SetPopulateEntities(files ...string) Config {
	return NewConfig(2002, func(s *Server) error {
		for _, f := range files {
			s.Printf("loading features from %s", f)
		}
		return s.PopulateEntityYaml(files...)
	})
}

func SetHandler(hs ...*Handler) Config {
	return NewConfig(2000, func(s *Server) error {
		for _, h := range hs {
			s.SetHandle(h)
		}
		return nil
	})
}
