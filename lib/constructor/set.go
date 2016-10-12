package constructor

import (
	"fmt"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

var GenerateSetKey feature.Constructor

func replaceWithLocals(l map[string]string, vals []string) []string {
	var ret []string
	for _, v := range vals {
		spl := strings.Split(v, ":")
		if len(spl) > 1 {
			if rep, exists := l[spl[1]]; exists {
				ret = append(ret, fmt.Sprintf("%s:%s", spl[0], rep))
			} else {
				ret = append(ret, fmt.Sprintf("%s:%s", spl[0], spl[1]))
			}
		} else {
			ret = append(ret, v)
		}
	}
	return ret
}

func generateSetKey(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()
	key := list[0]
	rest := list[1:]
	var setList []string
	locals := make(map[string]string)
	for _, v := range rest {
		spl := strings.Split(v, ":")
		var h string
		fKey := spl[0]
		withConstructor := spl[1]
		if withConstructor != "NO" {
			localFeatureName := fmt.Sprintf("%s-%s", key, fKey)
			vals := semiColonReplace(strings.Split(spl[2], ","))
			lvals := replaceWithLocals(locals, vals)
			cn, _ := e.GetConstructor(withConstructor)
			e.SetFeature(feature.NewRawFeature([]string{key}, localFeatureName, lvals, cn))
			h = fmt.Sprintf("%s:%s", fKey, localFeatureName)
			locals[fKey] = localFeatureName
		} else {
			h = fmt.Sprintf("%s:%s", fKey, fKey)
			locals[fKey] = fKey
		}
		setList = append(setList, h)
	}
	return directFromList("GENERATE_SET_KEY", []string{key}, tag, setList, setList, e)
}

func init() {
	GenerateSetKey = feature.NewConstructor("GENERATE_SET_KEY", 9000, generateSetKey)
	feature.SetConstructor(GenerateSetKey)
}
