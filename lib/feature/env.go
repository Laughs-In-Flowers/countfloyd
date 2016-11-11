package feature

import (
	"io/ioutil"
	"sync"

	"github.com/Laughs-In-Flowers/data"
)

type Env interface {
	Raw
	Constructors
	Features
	Applicator
	Populator
}

type Applicator interface {
	Apply([]string, *data.Vector, ...MapFn) error
	ApplyFor([]string, *data.Vector, int, ...MapFn) error
}

type Populator interface {
	Populate([]byte) error
	PopulateConstructors(...Constructor) error
	PopulateYamlFiles(...string) error
	PopulateGroup(...string) error
}

type env struct {
	Raw
	Constructors
	Features
}

func Empty() Env {
	e := &env{}
	e.Raw = NewRaw(e)
	e.Constructors = internal
	e.Features = NewFeatures(e)
	return e
}

func New(raw []byte, cs ...Constructor) (Env, error) {
	e := Empty()
	if err := e.PopulateConstructors(cs...); err != nil {
		return nil, err
	}
	if err := e.Populate(raw); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *env) Populate(raw []byte) error {
	err := e.queue(raw)
	if err != nil {
		return err
	}
	e.dequeue()
	return nil
}

func (e *env) PopulateConstructors(cs ...Constructor) error {
	e.SetConstructor(cs...)
	return nil
}

func (e *env) PopulateYamlFiles(files ...string) error {
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
