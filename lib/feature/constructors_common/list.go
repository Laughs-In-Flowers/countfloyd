package constructors_common

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

func listFrom(
	from string,
	group []string,
	tag string,
	raw []string,
	values []string,
	e feature.CEnv,
	modifiers ...listModifier,
) (feature.Informer, feature.Emitter, feature.Mapper) {
	ef := func() data.Item {
		nv := values
		if len(modifiers) > 0 {
			for _, m := range modifiers {
				nv = m(e, nv)
			}
		}
		return data.NewStringsItem(tag, nv...)
	}
	mf := func(d *data.Vector) {
		d.Set(ef())
	}
	return construct(from, group, tag, raw, values, ef, mf)
}

//
func List() feature.Constructor {
	return feature.NewConstructor(
		"LIST",
		1,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			list := r.MustGetValues()
			return listFrom("LIST", r.Group, tag, list, list, e)
		})
}

//
func ListExpand() feature.Constructor {
	return feature.NewConstructor(
		"LIST_EXPAND",
		5,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			values := r.MustGetValues()
			ef := func() data.Item {
				return data.NewStringsItem(tag, values...)
			}
			mf := func(d *data.Vector) {
				exp := expander(e, values)
				i := data.NewStringsItem(tag, exp...)
				d.Set(i)
			}
			return construct("LIST_EXPAND", r.Group, tag, values, values, ef, mf)
		})
}

//
func ListWithNull() feature.Constructor {
	return feature.NewConstructor(
		"LIST_WITH_NULL",
		2,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			list := argsMustBeLength(e, r.MustGetValues(), 2)
			null := list[0]
			vals := kToList(list[1], e)
			return listFrom("LIST_WITH_NULL", r.Group, tag, r.Values, vals, e, nullifier(null))
		})
}

//
func ListShuffle() feature.Constructor {
	return feature.NewConstructor(
		"LIST_SHUFFLE",
		3,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			list := r.MustGetValues()
			vals := kToList(list[0], e)
			return listFrom("LIST_SHUFFLE", r.Group, tag, r.Values, vals, e, shuffler)
		})
}

//
func ListExpandIntRange() feature.Constructor {
	return feature.NewConstructor(
		"LIST_EXPAND_INTRANGE",
		4,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			list := r.MustGetValues()
			vals := kToList(list[0], e)
			return listFrom("LIST_EXPAND_INTRANGE", r.Group, tag, r.Values, vals, e, expander)
		})
}

//
func ListMirrorInts() feature.Constructor {
	return feature.NewConstructor(
		"LIST_EXPAND_MIRRORINTS",
		6,
		func(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
			list := r.MustGetValues()
			vals := kToList(list[0], e)
			return listFrom("LIST_EXPAND_MIRRORINTS", r.Group, tag, r.Values, vals, e, mirrorInts)
		})
}
