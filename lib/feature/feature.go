package feature

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/xrr"
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
		strings.Split(i.raw, ","),
		nil,
	}
}

func (i *informer) Values() []string {
	return i.values
}

func (i *informer) Length() int {
	return len(i.values)
}

//func (i *informer) Bytes() ([]byte, error) {
//	return json.Marshal(i.RawFeature())
//}

//func (i *informer) MarshalJSON() ([]byte, error) {
//	return i.Bytes()
//}

type EmitFn func() data.Item

type Emitter interface {
	Emit() data.Item
	EmitString() (data.StringItem, error)
	EmitStrings() (data.StringsItem, error)
	EmitBool() (data.BoolItem, error)
	EmitInt() (data.IntItem, error)
	EmitFloat() (data.Float64Item, error)
	EmitVector() (data.VectorItem, error)
}

type emitter struct {
	eFn EmitFn
}

func NewEmitter(efn EmitFn) Emitter {
	return &emitter{
		eFn: efn,
	}
}

func (e *emitter) Emit() data.Item {
	return e.eFn()
}

var EmitTypeError = xrr.Xrror("unable to emit item as %s").Out

func (e *emitter) EmitString() (data.StringItem, error) {
	f := e.Emit()
	if sf, ok := f.(data.StringItem); ok {
		return sf, nil
	}
	return nil, EmitTypeError("string")
}

func (e *emitter) EmitStrings() (data.StringsItem, error) {
	f := e.Emit()
	if sf, ok := f.(data.StringsItem); ok {
		return sf, nil
	}
	return nil, EmitTypeError("strings")
}

func (e *emitter) EmitBool() (data.BoolItem, error) {
	f := e.Emit()
	if sf, ok := f.(data.BoolItem); ok {
		return sf, nil
	}
	return nil, EmitTypeError("bool")
}

func (e *emitter) EmitInt() (data.IntItem, error) {
	f := e.Emit()
	if sf, ok := f.(data.IntItem); ok {
		return sf, nil
	}
	return nil, EmitTypeError("int")
}

func (e *emitter) EmitFloat() (data.Float64Item, error) {
	f := e.Emit()
	if sf, ok := f.(data.Float64Item); ok {
		return sf, nil
	}
	return nil, EmitTypeError("float")
}

func (e *emitter) EmitVector() (data.VectorItem, error) {
	f := e.Emit()
	if sf, ok := f.(data.VectorItem); ok {
		return sf, nil
	}
	return nil, EmitTypeError("multi")
}

type MapFn func(*data.Vector)

type Mapper interface {
	Map(*data.Vector)
}

type mapper struct {
	mFn MapFn
}

func NewMapper(mfn MapFn) Mapper {
	return &mapper{mfn}
}

func (m *mapper) Map(d *data.Vector) {
	m.mFn(d)
}

type FeatureGroup struct {
	value string
	list  []RawFeature
}

func (fg *FeatureGroup) List() []RawFeature {
	return fg.list
}

func (fg *FeatureGroup) Bytes() ([]byte, error) {
	return yaml.Marshal(fg.list)
}

func (fg *FeatureGroup) Compress() *bytes.Buffer {
	b := new(bytes.Buffer)
	w := zlib.NewWriter(b)
	by, _ := fg.Bytes()
	w.Write(by)
	w.Close()
	return b
}

func (fg *FeatureGroup) Value() string {
	return fg.base64Encode()
}

func (fg *FeatureGroup) base64Encode() string {
	if fg.value == "" {
		b := fg.Compress()
		fg.value = base64.StdEncoding.EncodeToString(b.Bytes())
	}
	return fg.value
}

func DecodeFeatureGroup(s string) (*FeatureGroup, error) {
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(d)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	fs := new(bytes.Buffer)
	io.Copy(fs, r)
	var rf []RawFeature
	err = yaml.Unmarshal(fs.Bytes(), &rf)
	if err != nil {
		return nil, err
	}
	return &FeatureGroup{"", rf}, nil
}

type Features interface {
	AddFeature(...Feature)
	SetFeature(*RawFeature) error
	GetFeature(string) Feature
	MustGetFeature(string) Feature
	GetGroup(string) *FeatureGroup
	List(string) []RawFeature
	Remove(...string) error
}

type features struct {
	e   CEnv
	has map[string]Feature
}

func NewFeatures(e CEnv) Features {
	return &features{
		e:   e,
		has: make(map[string]Feature),
	}
}

func (fs *features) AddFeature(nfs ...Feature) {
	for _, nf := range nfs {
		fs.has[nf.Tag()] = nf
	}
}

func (fs *features) SetFeature(rf *RawFeature) error {
	KEY := strings.ToUpper(rf.Tag)
	if _, exists := fs.has[KEY]; exists {
		return ExistsError("feature", KEY)
	}
	if rf.Constructor == nil {
		err := applyConstructor(fs.e, rf)
		if err != nil {
			return err
		}
	}
	fs.has[KEY] = rf.Constructor.Construct(KEY, rf, fs.e)
	return nil
}

func (fs *features) GetFeature(key string) Feature {
	if f, ok := fs.has[strings.ToUpper(key)]; ok {
		return f
	}
	return nil
}

func (fs *features) MustGetFeature(key string) Feature {
	f := fs.GetFeature(key)
	if f == nil {
		logErr := NotFoundError("feature", key)
		fmt.Fprintf(os.Stderr, "FATAL: %s\n", logErr)
		os.Exit(-9000)
	}
	return f
}

func (fs *features) GetGroup(g string) *FeatureGroup {
	var ret []RawFeature
	for _, f := range fs.has {
		if f.IsGroup(g) {
			ret = append(ret, f.RawFeature())
		}
	}
	return &FeatureGroup{list: ret}
}

func (fs *features) List(group string) []RawFeature {
	g := fs.GetGroup(group)
	return g.List()
}

func (fs *features) remove(group string) error {
	for k, f := range fs.has {
		if f.IsGroup(group) {
			delete(fs.has, k)
		}
	}
	return nil
}

func (fs *features) Remove(groups ...string) error {
	for _, group := range groups {
		err := fs.remove(group)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewData(n float64) *data.Vector {
	d := data.New("")
	d.Set(data.NewFloat64Item("meta.priority", n))
	return d
}
