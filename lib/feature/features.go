package feature

import (
	"log"
	"strings"
)

type Features interface {
	SetFeature(*RawFeature)
	GetFeature(string) Feature
	MustGetFeature(string) Feature
	GetGroup(string) *FeatureSet
	List(string) []RawFeature
	ListString(string) string
	ListKeys(string) []string
}

type features struct {
	e   *env
	has map[string]Feature
}

func NewFeatures(e *env) Features {
	return &features{
		e:   e,
		has: make(map[string]Feature),
	}
}

func (fs *features) SetFeature(rf *RawFeature) {
	TAG := strings.ToUpper(rf.Tag)
	fs.has[TAG] = rf.constructor.Construct(TAG, rf, fs.e)
}

func (fs *features) GetFeature(tag string) Feature {
	if f, ok := fs.has[strings.ToUpper(tag)]; ok {
		return f
	}
	return nil
}

func (fs *features) MustGetFeature(tag string) Feature {
	f := fs.GetFeature(tag)
	if f == nil {
		log.Fatalf("feature: %s not found, exiting", tag)
	}
	return f
}

func (fs *features) GetGroup(g string) *FeatureSet {
	var ret []RawFeature
	for _, f := range fs.has {
		if f.IsGroup(g) {
			ret = append(ret, f.RawFeature())
		}
	}
	return &FeatureSet{list: ret}
}

func (fs *features) List(group string) []RawFeature {
	g := fs.GetGroup(group)
	return g.List()
}

func (fs *features) ListString(group string) string {
	g := fs.GetGroup(group)
	return g.ListString()
}

func (fs *features) ListKeys(group string) []string {
	g := fs.GetGroup("")
	return g.Keys()
}
