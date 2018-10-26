package env

import (
	"sort"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

// A function taking *env and returning an error.
type ConfigFn func(*env) error

// An interface providing Order & Configure functions.
type Config interface {
	Order() int
	Configure(*env) error
}

type config struct {
	order int
	fn    ConfigFn
}

// Returns a default Config with order of 50 and the provided ConfigFn.
func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

// Returns a Config with the provided order and ConfigFn.
func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

// Returns an integer used for ordering.
func (c config) Order() int {
	return c.order
}

// Provided a *env runs any defined functionality, returning any error.
func (c config) Configure(e *env) error {
	return c.fn(e)
}

type configList []Config

// Len for sort.Sort.
func (c configList) Len() int {
	return len(c)
}

// Swap for sort.Sort.
func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less for sort.Sort.
func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

// An interface providing facility for multiple configuration options.
type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	e          *env
	configured bool
	list       configList
}

func newConfiguration(e *env, conf ...Config) *configuration {
	c := &configuration{
		e:    e,
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

// Adds any number of Config to the Configuration.
func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func configure(e *env, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(e)
		if err != nil {
			return err
		}
	}
	return nil
}

// Runs all configuration for this Configuration, return any encountered error immediately.
func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.e, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

// Returns a boolean indicating if Configuration has run Configure.
func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{}

func SetPopulateFeature(groups []string, p []byte) Config {
	return DefaultConfig(func(e *env) error {
		if err := e.populateFeature(groups, p); err != nil {
			return err
		}
		return nil
	})
}

func SetConstructors(cs ...feature.Constructor) Config {
	return DefaultConfig(func(e *env) error {
		if err := e.SetConstructor(cs...); err != nil {
			return err
		}
		return nil
	})
}

func SetConstructorPlugin(dirs ...string) Config {
	return DefaultConfig(func(e *env) error {
		if err := e.PopulateConstructorPlugin(dirs...); err != nil {
			return err
		}
		return nil
	})
}

func SetFeaturePlugin(groups []string, dirs ...string) Config {
	return DefaultConfig(func(e *env) error {
		if err := e.PopulateFeaturePlugin(groups, dirs...); err != nil {
			return err
		}
		return nil
	})
}
