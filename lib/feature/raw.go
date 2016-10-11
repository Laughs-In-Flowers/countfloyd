package feature

import (
	"log"
	"sort"

	"gopkg.in/yaml.v2"
)

type Raw interface {
	sort.Interface
	queue([]byte)
	dequeue()
}

type RawFeature struct {
	Set         []string
	Tag         string
	Apply       string
	Values      []string
	constructor Constructor
}

func (r *RawFeature) MustGetValues() []string {
	list := r.Values
	if len(list) < 1 {
		log.Fatalf("zero length list for %s", r.Tag)
	}
	return list
}

func NewRawFeature(set []string, tag string, values []string, c Constructor) *RawFeature {
	return &RawFeature{
		Set:         set,
		Tag:         tag,
		Values:      values,
		constructor: c,
	}
}

type raw struct {
	e   *env
	has []*RawFeature
}

func NewRaw(e *env) Raw {
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
	return r.has[i].constructor.Order() < r.has[j].constructor.Order()
}

func (r *raw) queue(in []byte) {
	var rfs []*RawFeature
	yaml.Unmarshal(in, &rfs)
	for _, rf := range rfs {
		if c, exists := r.e.GetConstructor(rf.Apply); exists {
			rf.constructor = c
			r.has = append(r.has, rf)
		}
	}
}

func (r *raw) dequeue() {
	sort.Sort(r)
	for i, rf := range r.has {
		r.e.SetFeature(rf)
		r.has[i] = nil
	}
	r.has = nil
}
