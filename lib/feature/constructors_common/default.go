package constructors_common

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

// the default constructor
func Default() feature.Constructor {
	return feature.NewConstructor("DEFAULT", 1, defaultConstructor)
}

func defaultConstructor(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	var val string
	if len(list) >= 1 {
		val = list[0]
	}
	ef := func() data.Item {
		return data.NewStringItem(tag, val)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return construct("default", r.Group, tag, list, list, ef, mf)
}
