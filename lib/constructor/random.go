package constructor

import (
	mr "math/rand"
	"strconv"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

func random(from, tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	sd, err := strconv.ParseFloat(list[0], 64)
	if err != nil {
		sd = 0.0
	}

	vals := list[1:]
	lv := len(vals)
	var ssv func() string
	switch {
	case lv > 1:
		ssv = func() string { return vals[mr.Intn(lv)] }
	default:
		ssv = func() string { return vals[0] }
	}

	ef := func() data.Item {
		ret := data.NewStringItem(tag, "")
		if maybe(sd) {
			ret.SetString(ssv())
		}
		return ret
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return construct(from, r.Set, tag, list, list, ef, mf)
}

func SimpleRandom() feature.Constructor {
	return feature.DefaultConstructor("SIMPLE_RANDOM", simpleRandom)
}

func simpleRandom(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	return random("RANDOM", tag, r, e)
}

func SourcedRandom() feature.Constructor {
	return feature.NewConstructor("SOURCED_RANDOM", 10000, sourcedRandom)
}

func sourcedRandom(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	if len(list) != 2 {
		return nil, nil, nil
	}
	f := e.GetFeature(list[1])
	if fv, err := f.EmitStrings(); err == nil {
		fvs := fv.ToStrings()
		var nv []string
		nv = append(nv, list[0])
		nv = append(nv, fvs...)
		r.Values = nv
		return random("SOURCED_RANDOM", tag, r, e)
	}
	return nil, nil, nil
}
