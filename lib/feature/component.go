package feature

import (
	"log"

	"github.com/Laughs-In-Flowers/data"
)

type Component interface {
	Tagger
	Defines() []string
	Features() []string
}

type RawComponent struct {
	Tag      string
	Defines  []*RawFeature
	Features []*RawFeature
}

type component struct {
	tag      string
	defines  []string
	features []string
}

func (c *component) Tag() string {
	return c.tag
}

func (c *component) Defines() []string {
	return c.defines
}

func (c *component) Features() []string {
	return c.features
}

type Components interface {
	SetRawComponent(...*RawComponent) error
	SetComponent(...Component) error
	GetComponent(int, string, ...string) []*data.Vector
	MustGetComponent(int, string, ...string) []*data.Vector
	ListComponents() []Component
}

type components struct {
	e   *env
	has map[string]Component
}

func newComponents(e *env) Components {
	return &components{
		e, make(map[string]Component),
	}
}

func (c *components) SetRawComponent(rcs ...*RawComponent) error {
	var s []Component
	for _, v := range rcs {
		t := v.Tag
		var d, f []string
		for _, vd := range v.Defines {
			d = append(d, vd.Tag)
		}
		for _, vf := range v.Features {
			f = append(f, vf.Tag)
		}
		s = append(s, &component{t, d, f})
	}
	return c.SetComponent(s...)
}

func (c *components) SetComponent(cs ...Component) error {
	for _, v := range cs {
		nt := v.Tag()
		if _, exists := c.has[nt]; exists {
			return ExistsError("component", nt)
		}
		c.has[nt] = v
	}
	return nil
}

func newComponentVector(e Env, cc Component, id string, priority int) *data.Vector {
	d := NewData(priority)
	fs := cc.Features()
	e.Apply(fs, d)
	d.SetString("component.tag", cc.Tag())
	d.SetString("entity", id)
	return d
}

func getComponentVector(c *components, key, id string, priority int) (*data.Vector, error) {
	if cm, exists := c.has[key]; exists {
		return newComponentVector(c.e, cm, id, priority), nil
	}
	return nil, DoesNotExistError("component", key)
}

func (c *components) GetComponent(priority int, id string, k ...string) []*data.Vector {
	var ret []*data.Vector
	for _, key := range k {
		if cm, err := getComponentVector(c, key, id, priority); err == nil {
			ret = append(ret, cm)
		}
	}
	return ret
}

func (c *components) MustGetComponent(priority int, id string, k ...string) []*data.Vector {
	var ret []*data.Vector
	for _, key := range k {
		cm, err := getComponentVector(c, key, id, priority)
		if err != nil {
			logErr := NotFoundError("component", key)
			log.Fatalf(logErr.Error())
		}
		ret = append(ret, cm)
	}
	return ret
}

func (c *components) ListComponents() []Component {
	var ret []Component
	for _, v := range c.has {
		ret = append(ret, v)
	}
	return ret
}
