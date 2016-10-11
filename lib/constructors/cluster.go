package constructors

import (
	//"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

var (
	Cluster feature.Constructor
)

//func featureCluster(f feature.Feature, length int) *data.Item {
//ret := valueless("??? featureCluster")
//var ss []string
//for i := 0; i <= length; i++ {
//	item := f.Emit()
//	ss = append(ss, item.ToString()) //ret = append(ret, f.Emit())
//}
//ret.SetList(ss...)
//return ret
//}

//func featureClusterNoRepeat(f feature.Feature, length int) *data.Item {
//ret := valueless("????? featureClusterNoRepeat")
//has := make(map[interface{}]struct{})
//for i := 0; i <= length; i++ {
//	ret = append(ret, pick(f, has))
//}
//return ret
//}

//func pick(f feature.Feature, has map[string]struct{}) string {
//	i := f.Emit()
//	g := i.ToString()
//	if _, ok := has[g]; !ok {
//		has[g] = struct{}{}
//		return g
//	}
//	return pick(f, has)
//}

//type ClusterFunc func(feature.Feature, int) []interface{}

//func selectClusterFunc(is []string) ClusterFunc {
//	var ret ClusterFunc = featureCluster
//	l := len(is)
//	if is[l-1] == "true" {
//		ret = featureClusterNoRepeat
//	}
//	return ret
//}

//func mkCluster(from string, tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
//list := r.Values

//source := e.MustGetFeature(list[0])

//l := stringToInt(list[1])

//cfn := selectClusterFunc(list)

//ef := func() *data.Item {
//	return fn(source, l)
//}

//mf := func(f *feature.Data) {
//f.Set("", tag, fn(source, l))
//}

//return construct(from, r.Set, tag, list, source.Values(), ef, mf)
//}

func cluster(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	return nil, nil, nil //mkCluster("cluster", tag, r, e)
}

func clusterArgParser(e feature.Env, args []string) (feature.Feature, int, bool) {
	args = argsMustBeLength(e, args, 3)

	source := e.MustGetFeature(args[0])

	length := stringToInt(args[1])

	noRepeat := stringToBool(args[2])

	return source, length, noRepeat
}

//func clusterRandomLength(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
//	list := r.Values
//	cf := list[0]
//	ft := e.MustGetFeature(cf)
//
//	l := stringToInt(list[1])
//
//	fn := selectClusterFunc(list)
//
//	ef := func() *data.Item {
//		rl := randRange(l)
//		return fn(ft, rl)
//	}
//
//	mf := func(d *feature.Data) {
//		//rl := randRange(l)
//		//f.Set(tag, fn(ft, rl))
//	}
//
//	return construct("CLUSTER_RANDOM_LENGTH", r.Set, tag, list, ft.Values(), ef, mf)
//}

//func clusterPercentageOf(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
//	list := r.Values
//	cf := list[0]
//	ft := e.MustGetFeature(cf)
//
//	var pt float64 = 0.5
//	if p, err := strconv.ParseFloat(list[1], 64); err != nil {
//		pt = p
//	}
//
//	ptNumber := float64(float64(pt)/100) * float64(ft.Length())
//
//	r.Values = []string{cf, strconv.FormatInt(int64(ptNumber), 10)}
//
//	return mkCluster("CLUSTER_PERCENT_OF", tag, r, e)
//}

//func clusterRandomPercentageOf(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
//	list := r.Values
//	source := e.MustGetFeature(list[0])
//
//	var max int64 = 100
//	var min int64 = 1
//	max, _ = strconv.ParseInt(list[1], 10, 64)
//	min, _ = strconv.ParseInt(list[2], 10, 64)
//
//	rdpi := func() int {
//		var p int = 100
//		p = randInRange(int(min), int(max))
//		return int(float64(float64(p)/100) * float64(source.Length()))
//	}
//
//	fn := selectClusterFunc(list)
//
//	ef := func() *data.Item {
//		return fn(source, rdpi())
//	}
//
//	mf := func(d *feature.Data) {
//		//f.Set(tag, fn(source, rdpi()))
//	}
//
//	return construct("CLUSTER_RANDOM_PERCENT_OF", r.Set, tag, list, source.Values(), ef, mf)
//}

func init() {
	Cluster = feature.NewConstructor("Cluster", 800, cluster)
	feature.SetConstructor(Cluster)
}
