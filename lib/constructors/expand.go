package constructors

import (
	"sort"
	"strconv"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

var ExpandInts feature.Constructor

func expandInts(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	numbers := eiArgParser(e, r.MustGetValues())
	return directFromList("EXPANDED", r.Set, tag, r.Values, numbers, e)
}

func eiArgParser(e feature.Env, args []string) []string {
	args = argsMustBeLength(e, args, 2)
	var source feature.Feature
	source = e.MustGetFeature(args[0])
	i := source.Emit()
	str := i.ToList()
	switch args[1] {
	case "mirror":
		var nums []int
		var nstr []string
		for _, v := range str {
			num, err := strconv.Atoi(v)
			if err == nil {
				nums = append(nums, num)
				nums = append(nums, num-(num*2))
			}
		}
		sort.Ints(nums)
		for _, v := range nums {
			nstr = append(nstr, strconv.Itoa(v))
		}
		return nstr
	}
	return []string{"nothing to expand"}
}

func init() {
	ExpandInts = feature.NewConstructor("EXPAND_INTS", 3, expandInts)
	feature.SetConstructor(ExpandInts)
}
