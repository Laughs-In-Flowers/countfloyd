package constructors_common

import (
	"fmt"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

func CollectionMember() feature.Constructor {
	return feature.NewConstructor("COLLECTION_MEMBER", 50, collectionMember)
}

func collectionMember(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values
	ex := expander(e, list)
	mapped := listMappedToFloat64Keys(ex, 1)

	ef := func() data.Item {
		m := floatKeysToString(mapped)
		d := data.New("")
		for k, v := range m {
			d.Set(data.NewStringItem(k, v))
		}
		return data.NewVectorItem(tag, d)
	}

	mf := func(f *data.Vector) {
		n := f.ToFloat64("meta.priority")
		if v, ok := mapped[n]; ok {
			f.Set(data.NewStringItem(tag, v))
		}
	}

	return construct("COLLECTION_MEMBER", r.Group, tag, list, ex, ef, mf)
}

func CollectionMemberIndexed() feature.Constructor {
	return feature.NewConstructor("COLLECTION_MEMBER_INDEXED", 10000, collectionMemberIndexed)
}

func collectionMemberIndexed(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values
	ex := expander(e, list)
	mapped := listMappedToFloat64Keys(ex, 1)

	var nf []*feature.RawFeature
	for i, v := range mapped {
		nf = append(nf, &feature.RawFeature{
			Group:       r.Group,
			Tag:         fmt.Sprintf("%s_%d", r.Tag, int(i)),
			Apply:       "default",
			Values:      []string{v},
			Constructor: Default(),
		})
	}
	for _, ndf := range nf {
		e.SetFeature(ndf)
	}

	ef := func() data.Item {
		d := data.New("")
		for _, v := range nf {
			d.Set(data.NewStringItem(v.Tag, v.Values[0]))
		}
		return data.NewVectorItem(tag, d)
	}

	mf := func(f *data.Vector) {
		f.Set(ef())
	}

	return construct("COLLECTION_MEMBER_INDEXED", r.Group, tag, list, ex, ef, mf)
}
