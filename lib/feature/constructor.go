package feature

type Constructor interface {
	Tag() string
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
