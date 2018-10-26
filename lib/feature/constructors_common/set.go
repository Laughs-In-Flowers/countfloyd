package constructors_common

import (
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

// A set constructor that takes provided keys and links them to multiple features in a return map.
func Set() feature.Constructor {
	return feature.NewConstructor("SET", 50, set)
}

func set(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	raw := r.MustGetValues()

	values := argsMustBeLength(e, raw, 1)

	ef := func() data.Item {
		d := data.New("")
		kfs := extractKFS(e, strings.ToLower(tag), values...)
		for _, v := range kfs {
			i := v.f.Emit()
			i.NewKey(v.k)
			switch i.(type) {
			case data.VectorItem:
				if dv, err := v.f.EmitVector(); err == nil {
					vi := dv.ToVector()
					ldv := vi.List()
					for _, ldvi := range ldv {
						o := ldvi.Key()
						nk := smoothKey(v.k, o)
						ldvi.NewKey(nk)
					}
					d.Set(ldv...)
				}
			default:
				d.Set(i)
			}
		}
		return data.NewVectorItem(tag, d)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return construct("SET", r.Group, tag, raw, values, ef, mf)
}
