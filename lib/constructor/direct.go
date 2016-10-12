package constructor

import (
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var (
	Direct, DirectNull feature.Constructor
)

func direct(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	return directFromList("DIRECT", r.Set, tag, list, list, e)
}

func directNull(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	f := e.MustGetFeature(list[0])
	i := f.Emit()
	vals := i.ToList()
	vals = append(vals, "NULL")
	return directFromList("DIRECT-NULL", r.Set, tag, r.Values, vals, e)
}

func directFromList(
	from string,
	set []string,
	tag string,
	raw []string,
	values []string,
	e feature.Env,
) (feature.Informer, feature.Emitter, feature.Mapper) {
	var ret string
	if len(values) == 1 {
		ret = values[0]
	} else {
		ret = strings.Join(values, ",")
	}

	ef := func() *data.Item {
		i := data.NewItem(tag, "")
		i.SetString(ret)
		return i
	}

	mf := func(d *feature.Data) {
		d.Set(ef())
	}

	return construct(from, set, tag, raw, values, ef, mf)
}

func init() {
	Direct = feature.NewConstructor("DIRECT", 1, direct)
	DirectNull = feature.NewConstructor("DIRECT_NULL", 2, directNull)
	feature.SetConstructor(Direct, DirectNull)
}
