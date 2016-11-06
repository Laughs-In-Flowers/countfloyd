package constructor

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var CollectionMember feature.Constructor

func collectionMember(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values
	ex := listExpand(list)
	mapped := listMappedToIntKeys(ex)

	ef := func() data.Item {
		m := intKeysToString(mapped)
		d := data.New("")
		for k, v := range m {
			d.Set(data.NewStringItem(k, v))
		}
		return data.NewMultiItem(tag, d)
	}

	mf := func(f *data.Container) {
		n := f.ToInt("feature.priority")
		if v, ok := mapped[n]; ok {
			f.Set(data.NewStringItem(tag, v))
		}
	}

	return construct("COLLECTION_MEMBER", r.Set, tag, list, ex, ef, mf)
}

func init() {
	CollectionMember = feature.NewConstructor("COLLECTION_MEMBER", 50, collectionMember)
	feature.SetConstructor(CollectionMember)
}
