package feature

import (
	"log"
	"sort"

	"github.com/Laughs-In-Flowers/xrr"
	"gopkg.in/yaml.v2"
)

type RawFeature struct {
	Group       []string
	Tag         string
	Apply       string
	Values      []string
	Constructor Constructor
}

func (r *RawFeature) MustGetValues() []string {
	list := r.Values
	if len(list) < 1 {
		log.Fatalf("zero length list for %s", r.Tag)
	}
	return list
}

type Raw interface {
	Queue([]byte) error
	Dequeue(...string)
	//DeqComponent([]*RawComponent) error
	//DeqEntity([]*RawEntity) error
	AddRaw(...*RawFeature) error
}

type raw struct {
	e   CEnv
	has []*RawFeature
}

func NewRaw(e CEnv) *raw {
	return &raw{
		e:   e,
		has: make([]*RawFeature, 0),
	}
}

func (r *raw) Len() int {
	return len(r.has)
}

func (r *raw) Swap(i, j int) {
	r.has[i], r.has[j] = r.has[j], r.has[i]
}

func (r *raw) Less(i, j int) bool {
	return r.has[i].Constructor.Order() < r.has[j].Constructor.Order()
}

var NoConstructorError = xrr.Xrror("Constructor with tag %s does not exist.").Out

func applyConstructor(e CEnv, rf *RawFeature) error {
	var c Constructor
	var exists bool
	if c, exists = e.GetConstructor(rf.Apply); !exists {
		c, exists = e.GetConstructor("default")
		if !exists {
			return NoConstructorError(rf.Apply)
		}
	}
	rf.Constructor = c
	return nil
}

func (r *raw) AddRaw(rfs ...*RawFeature) error {
	for _, rf := range rfs {
		err := applyConstructor(r.e, rf)
		if err != nil {
			return err
		}
		r.has = append(r.has, rf)
	}
	return nil
}

func (r *raw) Queue(in []byte) error {
	var rfs []*RawFeature
	err := yaml.Unmarshal(in, &rfs)
	if err != nil {
		return err
	}
	return r.AddRaw(rfs...)
}

func (r *raw) Dequeue(groups ...string) {
	sort.Sort(r)
	for i, rf := range r.has {
		rf.Group = append(rf.Group, groups...)
		r.e.SetFeature(rf)
		r.has[i] = nil
	}
	r.has = nil
}

func DeqComponent(e CEnv, rcs []*RawComponent) error {
	for _, rc := range rcs {
		var fs []*RawFeature
		fs = append(fs, rc.Defines...)
		fs = append(fs, rc.Features...)
		err := e.AddRaw(fs...)
		if err != nil {
			return err
		}
	}
	return e.SetRawComponent(rcs...)
}

func DeqEntity(e CEnv, res []*RawEntity) error {
	for _, re := range res {
		err := e.AddRaw(re.Defines...)
		if err != nil {
			return err
		}
		err = DeqComponent(e, re.Components)
		if err != nil {
			return err
		}
	}
	return e.SetRawEntity(res...)
}
