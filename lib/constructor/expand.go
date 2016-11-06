package constructor

import "github.com/Laughs-In-Flowers/countfloyd/lib/feature"

var ExpandInts feature.Constructor

func expandInts(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	numbers := eiArgParser(e, r.MustGetValues())
	return directFromList("EXPANDED", r.Set, tag, r.Values, numbers, e)
}

func eiArgParser(e feature.Env, args []string) []string {
	args = argsMustBeLength(e, args, 2)
	var source feature.Feature
	source = e.MustGetFeature(args[0])
	if i, err := source.EmitStrings(); err == nil {
		ss := i.ToStrings()
		switch args[1] {
		case "range":
			return listExpand(ss)
		case "mirror":
			return listMirror(ss)
		}
	}
	return []string{"nothing to expand"}
}

func init() {
	ExpandInts = feature.NewConstructor("EXPAND_INTS", 3, expandInts)
	feature.SetConstructor(ExpandInts)
}
