package feature

import "strings"

type Constructors interface {
	AddConstructors(...Constructors)
	SetConstructor(...Constructor)
	GetConstructor(string) (Constructor, bool)
	ListConstructors() []Constructor
}

type constructors struct {
	has map[string]Constructor
}

func NewConstructors() Constructors {
	return &constructors{make(map[string]Constructor)}
}

func AddConstructors(cd ...Constructors) {
	internal.AddConstructors(cd...)
}

func (c *constructors) AddConstructors(cd ...Constructors) {
	for _, cs := range cd {
		for _, constructor := range cs.ListConstructors() {
			c.SetConstructor(constructor)
		}
	}
}

func SetConstructor(cns ...Constructor) {
	internal.SetConstructor(cns...)
}

func (c *constructors) SetConstructor(cns ...Constructor) {
	for _, cn := range cns {
		c.has[cn.Tag()] = cn
	}
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
