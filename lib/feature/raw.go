package feature

import (
	"log"
	"sort"

	"gopkg.in/yaml.v2"
)

type Raw interface {
	sort.Interface
	queue([]byte) error
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

var NoConstructorError = Frror("Constructor with tag %s does not exist.").Out

func (r *raw) queue(in []byte) error {
	var rfs []*RawFeature
	err := yaml.Unmarshal(in, &rfs)
	if err != nil {
		return err
	}
	for _, rf := range rfs {
		var c Constructor
		var exists bool
		if c, exists = r.e.GetConstructor(rf.Apply); !exists {
			return NoConstructorError(rf.Apply)
		}
		rf.constructor = c
		r.has = append(r.has, rf)
	}
	return nil
}

func (r *raw) dequeue() {
	sort.Sort(r)
	for i, rf := range r.has {
		r.e.SetFeature(rf)
		r.has[i] = nil
	}
	r.has = nil
}
