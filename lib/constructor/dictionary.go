package constructor

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var (
	DictionarySingle, DictionaryEach,
	DictionaryForEach, DictionaryForEachFromSetKey feature.Constructor
)

func dictionarySingleFeature(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	args := r.MustGetValues()
	source := e.MustGetFeature(args[0])
	list := listExpand(args[1:])

	ef := func() *data.Item {
		ret := data.NewItem(tag, "")
		var set map[string]string
		for _, v := range list {
			i := source.Emit()
			set[v] = i.ToString()
		}
		ret.SetMap(set)
		return ret
	}

	mf := func(d *feature.Data) {
		var i []*data.Item
		for _, v := range list {
			item := source.Emit()
			item.Key = v
			i = append(i, item)
		}
		d.SetItem(i...)
	}

	return construct("DICTIONARY_SINGLE", r.Set, tag, r.Values, list, ef, mf)
}

func dictionaryEach(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	args := r.MustGetValues()

	ef := func() *data.Item {
		ret := data.NewItem(tag, "")
		set := make(map[string]string)
		for _, v := range args {
			f := e.MustGetFeature(v)
			i := f.Emit()
			set[v] = i.ToString()
		}
		ret.SetMap(set)
		return ret
	}

	mf := func(d *feature.Data) {
		var i []*data.Item
		for _, v := range args {
			f := e.MustGetFeature(v)
			item := f.Emit()
			i = append(i, item)
		}
		d.SetItem(i...)
	}

	return construct("DICTIONARY_EACH", r.Set, tag, args, args, ef, mf)
}

func dictionaryForEach(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	args := r.MustGetValues()
	apply := listSplit(args)

	ef := func() *data.Item {
		ret := data.NewItem(tag, "")
		set := make(map[string]string)
		for k, v := range apply {
			f := e.GetFeature(v)
			if f != nil {
				i := f.Emit()
				set[k] = i.ToString()
			} else {
				set[k] = v
			}
		}
		ret.SetMap(set)
		return ret
	}

	mf := func(d *feature.Data) {
		var i []*data.Item
		for k, v := range apply {
			item := data.NewItem(k, "")
			f := e.GetFeature(v)
			if f != nil {
				fi := f.Emit()
				i = append(i, fi)
			} else {
				item.SetString(v)
				i = append(i, item)
			}
			item = nil
		}
		d.SetItem(i...)
	}

	return construct("DICTIONARY_FOR_EACH", r.Set, tag, args, args, ef, mf)
}

func dictionaryForEachFromSetKey(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	args := r.MustGetValues()
	setKey := args[0]
	apply := listSplit(args[1:])

	ef := func() *data.Item {
		ret := data.NewItem(tag, "")
		set := make(map[string]string)
		sf := e.GetFeature(setKey)
		si := sf.Emit()
		t := si.ToMap()
		for k, v := range apply {
			f := e.MustGetFeature(t[v])
			fi := f.Emit()
			set[k] = fi.ToString()
		}
		ret.SetMap(set)
		return ret
	}

	mf := func(d *feature.Data) {
		sf := e.GetFeature(setKey)
		si := sf.Emit()
		t := si.ToMap()
		var i []*data.Item
		for k, v := range apply {
			f := e.MustGetFeature(t[v])
			ni := f.Emit()
			ni.Key = k
			i = append(i, ni)
		}
		d.SetItem(i...)
	}

	return construct("DICTIONARY_FOR_EACH_FROM_SET_KEY", r.Set, tag, args, args, ef, mf)
}

func init() {
	DictionarySingle = feature.NewConstructor("DICTIONARY_SINGLE", 200, dictionarySingleFeature)
	DictionaryEach = feature.NewConstructor("DICTIONARY_EACH", 201, dictionaryEach)
	DictionaryForEach = feature.NewConstructor("DICTIONARY_FOR_EACH", 203, dictionaryForEach)
	DictionaryForEachFromSetKey = feature.NewConstructor("DICTIONARY_FOR_EACH_FROM_SET_KEY", 9001, dictionaryForEachFromSetKey)
	feature.SetConstructor(DictionarySingle, DictionaryEach, DictionaryForEach, DictionaryForEachFromSetKey)
}
