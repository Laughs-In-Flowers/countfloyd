package feature

import "strings"

type Constructor interface {
	Tagger
	Order() int
	Construct(string, *RawFeature, Env) Feature
}

type ConstructorFn func(string, *RawFeature, Env) (Informer, Emitter, Mapper)

type constructor struct {
	tag   string
	order int
	fn    ConstructorFn
}

func DefaultConstructor(tag string, fn ConstructorFn) Constructor {
	c := constructor{tag, 50, fn}
	return c
}

func NewConstructor(tag string, order int, fn ConstructorFn) Constructor {
	c := constructor{tag, order, fn}
	return c
}

func (c constructor) Tag() string {
	return c.tag
}

func (c constructor) Order() int {
	return c.order
}

func (c constructor) Construct(name string, r *RawFeature, e Env) Feature {
	return NewFeature(c.fn(name, r, e))
}

type Constructors interface {
	SetConstructor(...Constructor) error
	GetConstructor(string) (Constructor, bool)
	ListConstructors() []Constructor
}

type constructors struct {
	has map[string]Constructor
}

func NewConstructors() Constructors {
	return &constructors{make(map[string]Constructor)}
}

func SetConstructor(cns ...Constructor) error {
	return internal.SetConstructor(cns...)
}

func (c *constructors) SetConstructor(cns ...Constructor) error {
	for _, cn := range cns {
		tag := cn.Tag()
		if _, exists := c.has[tag]; exists {
			return ExistsError("construct", tag)
		}
		c.has[tag] = cn
	}
	return nil
}

func GetConstructor(key string) (Constructor, bool) {
	return internal.GetConstructor(key)
}

func (c *constructors) GetConstructor(key string) (Constructor, bool) {
	if c, exists := c.has[strings.ToUpper(key)]; exists {
		return c, true
	}
	return nil, false
}

func ListConstructors() []Constructor {
	return internal.ListConstructors()
}

func (c *constructors) ListConstructors() []Constructor {
	var ret []Constructor
	for _, c := range c.has {
		ret = append(ret, c)
	}
	return ret
}

var internal Constructors

func init() {
	internal = NewConstructors()
}
