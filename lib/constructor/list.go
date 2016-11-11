package constructor

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

func listFrom(
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

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return construct(from, set, tag, raw, values, ef, mf)
}

func List() feature.Constructor {
	return feature.NewConstructor("LIST", 1, list)
}

func list(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	return listFrom("LIST", r.Set, tag, list, list, e)
}

type listModifier func([]string) []string

func fromStringsFeature(f string, e feature.Env) []string {
	ft := e.MustGetFeature(f)
	if i, err := ft.EmitStrings(); err == nil {
		return i.ToStrings()
	}
	return []string{}
}

func ListWithNull() feature.Constructor {
	return feature.NewConstructor("LIST_WITH_NULL", 2, listWithNull)
}

func listWithNull(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	vals := fromStringsFeature(list[0], e)
	mfn := func(ss []string) []string {
		ss = append(ss, "NULL")
		return ss
	}
	return listFrom("LIST_WITH_NULL", r.Set, tag, r.Values, vals, e, mfn)
}

func ListShuffle() feature.Constructor {
	return feature.NewConstructor("LIST_SHUFFLE", 3, listShuffle)
}

func listShuffle(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	vals := fromStringsFeature(list[0], e)
	mfn := func(ss []string) []string {
		shuffleStrings(ss)
		return ss
	}
	return listFrom("LIST_SHUFFLE", r.Set, tag, r.Values, vals, e, mfn)
}

func ListExpandIntRange() feature.Constructor {
	return feature.NewConstructor("LIST_EXPAND_INTRANGE", 5, listExpandIntRange)
}

func listExpandIntRange(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	vals := fromStringsFeature(list[0], e)
	return listFrom("LIST_EXPAND_INTRANGE", r.Set, tag, r.Values, vals, e, listExpand)
}

func ListMirrorInts() feature.Constructor {
	return feature.NewConstructor("LIST_EXPAND_MIRRORINTS", 5, listMirrorInts)
}

func listMirrorInts(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	vals := fromStringsFeature(list[0], e)
	return listFrom("LIST_EXPAND_MIRRORINTS", r.Set, tag, r.Values, vals, e, listMirror)
}
