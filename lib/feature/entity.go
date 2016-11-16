package feature

import (
	"log"

	"github.com/Laughs-In-Flowers/data"
)

type Entity interface {
	Tagger
	Components() []string
}

type RawEntity struct {
	Tag        string
	Components []*RawComponent
}

type entity struct {
	tag        string
	components []string
}

func (e *entity) Tag() string {
	return e.tag
}

func (e *entity) Components() []string {
	return e.components
}

type Entities interface {
	SetRawEntity(...*RawEntity) error
	SetEntity(...Entity) error
	GetEntity(int, string) []*data.Vector
	MustGetEntity(int, string) []*data.Vector
	ListEntities() []Entity
}

type entities struct {
	e   *env
	has map[string]Entity
}

func newEntities(e *env) Entities {
	return &entities{
		e, make(map[string]Entity),
	}
}

func (e *entities) SetRawEntity(res ...*RawEntity) error {
	var ex []Entity
	for _, v := range res {
		tag := v.Tag
		var cs []string
		for _, vv := range v.Components {
			cs = append(cs, vv.Tag)
		}
		ent := &entity{tag, cs}
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

func getEntity(e *env, ent Entity, priority int) []*data.Vector {
	id := data.V4Quick()
	comp := ent.Components()
	return e.GetComponent(priority, id, comp...)
}

func (e *entities) GetEntity(priority int, key string) []*data.Vector {
	if ent, exists := e.has[key]; exists {
		return getEntity(e.e, ent, priority)
	}
	return nil
}

func (e *entities) MustGetEntity(priority int, key string) []*data.Vector {
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
