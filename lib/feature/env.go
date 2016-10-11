package feature

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
)

type Env interface {
	Raw
	Constructors
	Features
	Applicator
	Populator
}

type Applicator interface {
	Apply([]string, *Data, int, ...MapFn) error
}

type Populator interface {
	Populate([]byte) error
	PopulateConstructors(...Constructor) error
	PopulateYamlFiles(...string) error
	PopulateSetValues(...string) error
}

type env struct {
	Raw
	Constructors
	Features
	L log.Logger
}

func Empty() Env {
	e := &env{}
	e.Raw = NewRaw(e)
	e.Constructors = internal
	e.Features = NewFeatures(e)
	return e
}

func New(raw []byte, cs ...Constructor) Env {
	e := Empty()
	e.Populate(raw)
	e.PopulateConstructors(cs...)
	return e
}

func (e *env) Populate(raw []byte) error {
	e.queue(raw)
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

func (e *env) PopulateSetValues(sv ...string) error {
	for _, s := range sv {
		d, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return err
		}
		b := bytes.NewBuffer(d)
		r, err := zlib.NewReader(b)
		if err != nil {
			return err
		}
		fs := new(bytes.Buffer)
		io.Copy(fs, r)
		err = e.Populate(fs.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *env) Apply(list []string, to *Data, pass int, with ...MapFn) error {
	for i := 1; i <= pass; i++ {
		for _, l := range list {
			if f := e.GetFeature(l); f != nil {
				f.Map(to)
			}
		}
		for _, fn := range with {
			fn(to)
		}
	}
	return nil
}
