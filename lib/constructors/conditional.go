package constructors

import (
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

// apply the first feature listed with tag, and subsequent features based on
// the result of the first.
func conditionalInitialComposite(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.Values
	firstTag, firstFeature := split(list[0])
	rest := listSplit(list[1:])

	wfn := func() map[string]string {
		cm := make(map[string]string)

		ff := e.MustGetFeature(firstFeature)
		ffI := ff.Emit()
		ffVal := ffI.ToString()
		cm[firstTag] = ffVal

		get := valueFormat(rest, ffVal)

		for k, v := range get {
			gf := e.MustGetFeature(v)
			gi := gf.Emit()
			cm[k] = gi.ToString()
		}

		return cm
	}

	ef := func() *data.Item {
		i := data.NewItem(tag, "")
		m := wfn()
		i.SetMap(m)
		return i
	}

	mf := func(d *feature.Data) {
		d.Set(ef())
	}

	return construct("CONDITIONAL_INITIAL_COMPOSITE", r.Set, tag, list, list, ef, mf)
}
