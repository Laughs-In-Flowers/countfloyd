package constructor

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var (
	Direct, DirectNull, DirectShuffle feature.Constructor
)

func direct(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	return directFromList("DIRECT", r.Set, tag, list, list, e)
}

type listModifier func([]string) []string

func directNull(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	f := e.MustGetFeature(list[0])
	if i, err := f.EmitStrings(); err == nil {
		vals := i.ToStrings()
		mfn := func(ss []string) []string {
			ss = append(ss, "NULL")
			return ss
		}
		return directFromList("DIRECT-NULL", r.Set, tag, r.Values, vals, e, mfn)
	}
	return nil, nil, nil
}

func directShuffle(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	f := e.MustGetFeature(list[0])
	if i, err := f.EmitStrings(); err == nil {
		vals := i.ToStrings()
		mfn := func(ss []string) []string {
			shuffleStrings(ss)
			return ss
		}
		return directFromList("DIRECT-SHUFFLE", r.Set, tag, r.Values, vals, e, mfn)
	}
	return nil, nil, nil
}

func directFromList(
	from string,
	set []string,
	tag string,
	raw []string,
	values []string,
	e feature.Env,
	modifiers ...listModifier,
) (feature.Informer, feature.Emitter, feature.Mapper) {
	ef := func() data.Item {
		nv := values
		if len(modifiers) > 0 {
			for _, m := range modifiers {
				nv = m(nv)
			}
		}
		return data.NewStringsItem(tag, nv...)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return construct(from, set, tag, raw, values, ef, mf)
}

func init() {
	Direct = feature.NewConstructor("DIRECT", 1, direct)
	DirectNull = feature.NewConstructor("DIRECT_NULL", 2, directNull)
	DirectShuffle = feature.NewConstructor("DIRECT_SHUFFLE", 3, directShuffle)
	feature.SetConstructor(Direct, DirectNull, DirectShuffle)
}
