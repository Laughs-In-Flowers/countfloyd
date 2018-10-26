package env

import (
	"io/ioutil"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

type Env interface {
	Loader
	feature.Raw
	feature.Constructors
	feature.Features
	feature.Components
	feature.Entities
	Applicator
	Populator
}

type Applicator interface {
	Apply([]string, *data.Vector, ...feature.MapFn) error
	ApplyFor(int, []string, *data.Vector, ...feature.MapFn) error
}

type Populator interface {
	Populate([]byte) error
	PopulateConstructorPlugin(...string) error
	PopulateFeaturePlugin([]string, ...string) error
	PopulateFeatureYaml([]string, ...string) error
	PopulateFeatureGroupString([]string, ...string) error
	PopulateComponentYaml([]string, ...string) error
	PopulateEntityYaml([]string, ...string) error
}

type env struct {
	Loader
	feature.Raw
	feature.Constructors
	feature.Features
	feature.Components
	feature.Entities
}

func Empty() Env {
	return empty()
}

func empty() *env {
	e := &env{}
	e.Raw = feature.NewRaw(e)
	e.Loader, _ = NewPlugins()
	e.Constructors = feature.Internal
	e.Features = feature.NewFeatures(e)
	e.Components = feature.NewComponents(e)
	e.Entities = feature.NewEntities(e)
	return e
}

func New(cnf ...Config) (Env, error) {
	e := empty()
	c := newConfiguration(e, cnf...)
	if err := c.Configure(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *env) Populate(r []byte) error {
	return e.populateFeature([]string{}, r)
}

func (e *env) PopulateConstructorPlugin(dirs ...string) error {
	aErr := e.Loader.AddDirs(dirs...)
	if aErr != nil {
		return aErr
	}
	nc, cErr := e.Loader.LoadConstructor()
	if cErr != nil {
		return cErr
	}
	if sErr := e.Constructors.SetConstructor(nc...); sErr != nil {
		return sErr
	}
	return nil
}

func (e *env) PopulateFeaturePlugin(groups []string, dirs ...string) error {
	aErr := e.Loader.AddDirs(dirs...)
	if aErr != nil {
		return aErr
	}
	nfs, fErr := e.Loader.LoadFeature()
	if fErr != nil {
		return fErr
	}
	e.AddFeature(nfs...)
	return nil
}

func (e *env) populateFeature(g []string, r []byte) error {
	err := e.Queue(r)
	if err != nil {
		return err
	}
	e.Dequeue(g...)
	return nil
}

func (e *env) PopulateFeatureYaml(groups []string, files ...string) error {
	for _, file := range files {
		read, err := ioutil.ReadFile(file)
		if err == nil {
			err := e.populateFeature(groups, read)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (e *env) PopulateFeatureGroupString(groups []string, sv ...string) error {
	for _, s := range sv {
		set, err := feature.DecodeFeatureGroup(s)
		b, err := set.Bytes()
		if err != nil {
			return err
		}
		err = e.populateFeature(groups, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *env) PopulateComponentYaml(groups []string, files ...string) error {
	var err error
	for _, file := range files {
		var rcs []*feature.RawComponent
		var read []byte
		read, err = ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(read, &rcs)
		if err != nil {
			return err
		}
		err = feature.DeqComponent(e, rcs)
	}
	e.Dequeue(groups...)
	return err
}

func (e *env) PopulateEntityYaml(groups []string, files ...string) error {
	var err error
	for _, file := range files {
		var res []*feature.RawEntity
		var read []byte
		read, err = ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(read, &res)
		if err != nil {
			return err
		}
		err = feature.DeqEntity(e, res)
	}
	e.Dequeue(groups...)
	return err
}

// Apply the list of features to the provided data Vector, following up with
// the provided MapFn. Internally is ApplyFor where n = 1.
func (e *env) Apply(list []string, to *data.Vector, with ...feature.MapFn) error {
	return e.ApplyFor(1, list, to, with...)
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

// Apply for n number of passes the provided list of features to the provided data Vector,
// following up with the provided MapFn.
func (e *env) ApplyFor(pass int, list []string, to *data.Vector, with ...feature.MapFn) error {
	for i := 1; i <= pass; i = i + 1 {
		fill(e, list, to)
		for _, fn := range with {
			fn(to)
		}
	}
	return nil
}
