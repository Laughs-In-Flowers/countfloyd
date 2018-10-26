package constructors_common

import (
	"sort"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

//
func AlphaOrdered() feature.Constructor {
	return feature.NewConstructor(
		"ALPHA_ORDERED",
		6,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			v := r.MustGetValues()
			vals := kToList(v[0], e)
			sort.Strings(vals)
			return listFrom("ALPHA_ORDERED", r.Group, tag, v, vals, e)
		})
}

//
func RoundRobin() feature.Constructor {
	return feature.NewConstructor("ROUND_ROBIN", 50, roundRobin)
}

func roundRobin(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	values := r.MustGetValues()

	ef := func() data.Item {
		return data.NewStringsItem(tag, values...)
	}

	limit := len(values) - 1
	idx := limit
	nxt := func(curr, limit int) int {
		if curr == limit {
			return 0
		}
		return curr + 1
	}

	mf := func(f *data.Vector) {
		idx = nxt(idx, limit)
		v := values[idx]
		f.Set(data.NewStringItem(tag, v))
	}

	return construct("ROUND_ROBIN", r.Group, tag, values, values, ef, mf)
}
