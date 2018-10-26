package feature

import (
	"log"

	"github.com/Laughs-In-Flowers/data"
)

type Entity interface {
	Tagger
	Defines() []string
	Components() []string
}

type RawEntity struct {
	Tag        string
	Defines    []*RawFeature
	Components []*RawComponent
}

type entity struct {
	tag        string
	defines    []string
	components []string
}

func (e *entity) Tag() string {
	return e.tag
}

func (e *entity) Defines() []string {
	return e.defines
}

func (e *entity) Components() []string {
	return e.components
}

type Entities interface {
	SetRawEntity(...*RawEntity) error
	SetEntity(...Entity) error
	GetEntity(float64, string) []*data.Vector
	MustGetEntity(float64, string) []*data.Vector
	ListEntities() []Entity
}

type entities struct {
	e   CEnv
	has map[string]Entity
}

func NewEntities(e CEnv) Entities {
	return &entities{
		e, make(map[string]Entity),
	}
}

func (e *entities) SetRawEntity(res ...*RawEntity) error {
	var ex []Entity
	for _, v := range res {
		tag := v.Tag
		var d []string
		for _, vd := range v.Defines {
			d = append(d, vd.Tag)
		}
		var cs []string
		for _, vv := range v.Components {
			cs = append(cs, vv.Tag)
		}
		ent := &entity{tag, d, cs}
		ex = append(ex, ent)
	}
	return e.SetEntity(ex...)
}

func (e *entities) SetEntity(es ...Entity) error {
	for _, v := range es {
		nt := v.Tag()
		if _, exists := e.has[nt]; exists {
			return ExistsError("entity", nt)
		}
		e.has[nt] = v
	}
	return nil
}

func getEntity(e CEnv, ent Entity, priority float64) []*data.Vector {
	id := genUUID()
	comp := ent.Components()
	return e.GetComponent(priority, id, comp...)
}

func (e *entities) GetEntity(priority float64, key string) []*data.Vector {
	if ent, exists := e.has[key]; exists {
		return getEntity(e.e, ent, priority)
	}
	return nil
}

func (e *entities) MustGetEntity(priority float64, key string) []*data.Vector {
	ent, exists := e.has[key]
	if !exists {
		logErr := NotFoundError("entity", key)
		log.Fatalf(logErr.Error())
	}
	return getEntity(e.e, ent, priority)
}

func (e *entities) ListEntities() []Entity {
	var ret []Entity
	for _, v := range e.has {
		ret = append(ret, v)
	}
	return ret
}
