package constructors

import (
	mr "math/rand"
	"strconv"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var Random feature.Constructor

func random(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	sd, err := strconv.ParseFloat(list[0], 64)
	if err != nil {
		sd = 0.0
	}

	vals := list[1:]
	lv := len(vals)
	var ef func() *data.Item
	if lv > 1 {
		ef = func() *data.Item {
			ret := data.NewItem(tag, "")
			if maybe(sd) {
				ret.SetString(vals[mr.Intn(lv)])
			}
			return ret
		}
	} else {
		ef = func() *data.Item {
			ret := data.NewItem(tag, "")
			if maybe(sd) {
				ret.SetString(vals[0])
			}
			return ret
		}
	}

	mf := func(d *feature.Data) {
		i := ef()
		d.SetItem(i)
	}

	return construct("RANDOM", r.Set, tag, list, list, ef, mf)
}

func init() {
	Random = feature.DefaultConstructor("RANDOM", random)
	feature.SetConstructor(Random)
}
