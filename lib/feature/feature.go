package feature

import (
	"encoding/json"
	"strings"

	"github.com/Laughs-In-Flowers/data"
)

type Feature interface {
	Informer
	Emitter
	Mapper
}

type feature struct {
	Informer
	Emitter
	Mapper
}

func NewFeature(i Informer, e Emitter, m Mapper) Feature {
	return &feature{i, e, m}
}

type Grouper interface {
	Group() []string
	IsGroup(string) bool
}

type Tagger interface {
	Tag() string
}

type Parenter interface {
	From() string
}

type Detailer interface {
	Grouper
	Tagger
	Parenter
}

type Valuer interface {
	Raw() string
	Values() []string
	Length() int
}

type Transmitter interface {
	RawFeature() RawFeature
	Bytes() ([]byte, error)
	String() string
	json.Marshaler
}

type Informer interface {
	Detailer
	Valuer
	Transmitter
}

type informer struct {
	from   string
	group  []string
	tag    string
	raw    string
	values []string
}

func NewInformer(f string, g []string, t string, r []string, v []string) Informer {
	return &informer{
		from:   f,
		group:  g,
		tag:    t,
		raw:    strings.Join(r, ","),
		values: v,
	}
}

func (i *informer) Group() []string {
	return i.group
}

func (i *informer) IsGroup(g string) bool {
	for _, v := range i.group {
		if g == v {
			return true
		}
	}
	return false
}

func (i *informer) From() string {
	return i.from
}

func (i *informer) Tag() string {
	return i.tag
}

func (i *informer) Raw() string {
	return i.raw
}

func (i *informer) RawFeature() RawFeature {
	return RawFeature{
		i.group,
		i.tag,
		i.from,
		i.values,
		nil,
	}
}

func (i *informer) Values() []string {
	return i.values
}

func (i *informer) Length() int {
	return len(i.values)
}

func (i *informer) Bytes() ([]byte, error) {
	return json.Marshal(i.RawFeature())
}

func (i *informer) MarshalJSON() ([]byte, error) {
	return i.Bytes()
}

func (i *informer) String() string {
	ret, err := i.Bytes()
	if err != nil {
		return err.Error()
	}
	return string(ret)
}

type EmitFn func() *data.Item

type Emitter interface {
	Emit() *data.Item
}

type emitter struct {
	eFn EmitFn
}

func NewEmitter(efn EmitFn) Emitter {
	return &emitter{
		eFn: efn,
	}
}

func (e *emitter) Emit() *data.Item {
	return e.eFn()
}

type MapFn func(*Data)

type Mapper interface {
	Map(*Data)
}

type mapper struct {
	mFn MapFn
}

func NewMapper(mfn MapFn) Mapper {
	return &mapper{mfn}
}

func (m *mapper) Map(d *Data) {
	m.mFn(d)
}
