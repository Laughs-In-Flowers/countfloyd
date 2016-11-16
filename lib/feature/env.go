package feature

import (
	"io/ioutil"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/Laughs-In-Flowers/data"
)

type Env interface {
	Constructors
	Features
	Components
	Entities
	Applicator
	Populator
}

type Applicator interface {
	Apply([]string, *data.Vector, ...MapFn) error
	ApplyFor([]string, *data.Vector, int, ...MapFn) error
}

type Populator interface {
	Populate([]byte) error
	PopulateYaml(...string) error
	PopulateGroup(...string) error
	PopulateComponentYaml(...string) error
	PopulateEntityYaml(...string) error
}

type env struct {
	*raw
	Constructors
	Features
	Components
	Entities
}

func Empty() Env {
	e := &env{}
	e.raw = newRaw(e)
	e.Constructors = internal
	e.Features = newFeatures(e)
	e.Components = newComponents(e)
	e.Entities = newEntities(e)
	return e
}

func New(r []byte, cs ...Constructor) (Env, error) {
	e := Empty()
	if err := e.SetConstructor(cs...); err != nil {
		return nil, err
	}
	if err := e.Populate(r); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *env) Populate(r []byte) error {
	err := e.queue(r)
	if err != nil {
		return err
	}
	e.dequeue()
	return nil
}

func (e *env) PopulateYaml(files ...string) error {
	for _, file := range files {
		read, err := ioutil.ReadFile(file)
		if err == nil {
			err := e.Populate(read)
			if err != nil {
				return err
			}
			continue
		}
		return err
	}
	return nil
}

func (e *env) PopulateGroup(sv ...string) error {
	for _, s := range sv {
		set, err := DecodeFeatureGroup(s)
		b, err := set.Bytes()
		if err != nil {
			return err
		}
		err = e.Populate(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *env) PopulateComponentYaml(files ...string) error {
	var err error
	for _, file := range files {
		var rcs []*RawComponent
		var read []byte
		read, err = ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(read, &rcs)
		if err != nil {
			return err
		}
		err = deqComponent(e, rcs)
	}
	e.dequeue()
	return err
}

func (e *env) PopulateEntityYaml(files ...string) error {
	var err error
	for _, file := range files {
		var res []*RawEntity
		var read []byte
		read, err = ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(read, &res)
		if err != nil {
			return err
		}
		err = deqEntity(e, res)
	}
	e.dequeue()
	return err
}

func (e *env) Apply(list []string, to *data.Vector, with ...MapFn) error {
	return e.ApplyFor(list, to, 1, with...)
}

func fill(e Env, list []string, to *data.Vector) {
	var wg sync.WaitGroup
	ff := func(s string, to *data.Vector) {
		if ft := e.GetFeature(s); ft != nil {
			ft.Map(to)
		}
		wg.Done()
	}
	for _, l := range list {
		wg.Add(1)
		go ff(l, to)
	}
	wg.Wait()
}

func (e *env) ApplyFor(list []string, to *data.Vector, pass int, with ...MapFn) error {
	for i := 1; i <= pass; i++ {
		fill(e, list, to)
		for _, fn := range with {
			fn(to)
		}
	}
	return nil
}
