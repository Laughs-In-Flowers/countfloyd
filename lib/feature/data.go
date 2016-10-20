package feature

import (
	"strconv"

	"github.com/Laughs-In-Flowers/data"
)

type Data struct {
	*data.Container
}

func NewData(n int) *data.Container {
	d := data.NewContainer("FEATURES")
	d.Set(data.NewItem("feature.priority", strconv.Itoa(n)))
	return d
}

func DataFrom(m *data.Container, e Env) *data.Container {
	n := m.Get("meta.number")
	d := NewData(n.ToInt())
	a := m.ToList("meta.features")
	f := &Data{d}
	e.Apply(a, f)
	return f.Container
}
