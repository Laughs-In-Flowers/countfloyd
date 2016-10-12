package constructor

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var CollectionMember feature.Constructor

// apply the feature relative to (numbered) position as collection member across
// multiple generated instances
// ("myTag", []string{"1-3", "a",}, Env)
// feature generates: myTag: 1, myTag: 2, myTag: 3, myTag: a
func collectionMember(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values
	ex := listExpand(list)
	mapped := listMappedToIntKeys(ex)

	ef := func() *data.Item {
		i := data.NewItem(tag, "")
		i.SetMap(intKeysToString(mapped))
		return i
	}

	mf := func(f *feature.Data) {
		n := f.ToInt("feature.priority")
		i := data.NewItem(tag, mapped[n])
		i.SetString(mapped[n])
		f.Set(i)
	}

	return construct("COLLECTION_MEMBER", r.Set, tag, list, ex, ef, mf)
}

func init() {
	CollectionMember = feature.NewConstructor("COLLECTION_MEMBER", 50, collectionMember)
	feature.SetConstructor(CollectionMember)
}
