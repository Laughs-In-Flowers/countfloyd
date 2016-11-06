package constructor

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var Set feature.Constructor

func set(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()

	ef := func() data.Item {
		//i := make(map[string]string)
		//for _, v := range list {
		//	spl := strings.Split(v, ";")
		//	if len(spl) == 2 {
		//		ft := e.GetFeature(spl[1])
		//		if ft != nil {
		//			ei := ft.Emit()
		//			eik := strings.ToLower(strings.Join([]string{tag, spl[0]}, "."))
		//			i[eik] = ei.ToString()
		//		}
		//	}
		//}
		//ret := data.NewItem(tag, "")
		//ret.SetMap(i)
		//return ret
		return nil
	}

	mf := func(d *data.Container) {
		//i := ef()
		//mi := i.ToMap()
		//var nis []data.Item
		//for k, v := range mi {
		//	key := strings.Join([]string{tag, k}, ".")
		//	ni := data.NewItem(key, v)
		//	nis = append(nis, ni)
		//}
		//d.Set(nis...)
	}

	return construct("SET", r.Set, tag, list, list, ef, mf)
}

func init() {
	Set = feature.NewConstructor("SET", 50, set)
	feature.SetConstructor(Set)
}
