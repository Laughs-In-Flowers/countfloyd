package constructor

import (
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

func Set() feature.Constructor {
	return feature.NewConstructor("SET", 50, set)
}

type kf struct {
	k string
	f feature.Feature
}

func extractKF(e feature.Env, tag string, r string) *kf {
	spl := strings.Split(r, ";")
	if len(spl) != 2 {
		return nil
	}
	k := strings.Join([]string{tag, spl[0]}, ".")
	f := e.GetFeature(spl[1])
	return &kf{k, f}
}

func extractKFS(e feature.Env, tag string, r ...string) []*kf {
	var ret []*kf
	for _, v := range r {
		if x := extractKF(e, tag, v); x.f != nil {
			ret = append(ret, x)
		}
	}
	return ret
}

func set(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
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
						nk := strings.Join([]string{v.k, o}, ".")
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

	return construct("SET", r.Set, tag, raw, values, ef, mf)
}
