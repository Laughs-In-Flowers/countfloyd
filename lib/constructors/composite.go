package constructors

/*
import (
	"github.com/Laughs-In-Flowers/ifriit/lib/feature"
)

// a composite feature returning a map of the features in the list
func composite(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values

	mf := func(d *feature.Data) {
		//var comp = make(map[string]interface{})
		//for _, v := range list {
		//	ft := e.MustGetFeature(v)
		//	comp[v] = ft.Emit()
		//}
		//f.Set(tag, comp)
	}

	ef := func() interface{} {
		var comp = make(map[string]interface{})
		for _, v := range list {
			ft := e.MustGetFeature(v)
			comp[v] = ft.Emit()
		}
		return comp
	}

	return construct("COMPOSITE", r.Set, tag, list, list, ef, mf)
}

// two labeled clusters from a source list
func compositeClusterSplit(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values

	source := e.MustGetFeature(list[0])
	l := source.Length()

	k1, k2 := list[1], list[2]

	splFn := func() map[string]interface{} {
		splList := source.Values()
		shuffleStrings(splList)

		ret := make(map[string]interface{})
		var s1, s2 []interface{}
		splAt := randRange(l)
		for i, v := range splList {
			if i <= splAt {
				s1 = append(s1, v)
			}
		}
		ret[k1] = s1
		for i, v := range splList {
			if i > splAt {
				s2 = append(s2, v)
			}
		}
		ret[k2] = s2

		return ret
	}

	ef := func() interface{} {
		return splFn()
	}

	mf := func(f *feature.Data) {
		//f.Set(tag, splFn())
	}

	return construct("COMPOSITE_CLUSTER_SPLIT", r.Set, tag, list, source.Values(), ef, mf)
}
*/
