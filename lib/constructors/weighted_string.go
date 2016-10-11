package constructors

import (
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var WeightedString feature.Constructor

//
func weightedString(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	cs, numbers := wsArgParser(e, r.MustGetValues())

	ef := func() *data.Item {
		i := data.NewItem(tag, "")
		choose, err := cs.Choose()
		if err != nil {
			i.SetString(err.Error())
			return i
		}
		i.SetString(choose.String())
		return i
	}

	mf := func(d *feature.Data) {
		d.Set(ef())
	}

	return construct("WEIGHTED_STRING", r.Set, tag, numbers, numbers, ef, mf)
}

func wsArgParser(e feature.Env, args []string) (choices, []string) {
	args = argsMustBeLength(e, args, 2)

	var source feature.Feature
	var numbers []string

	switch {
	case args[0] == "SOURCED":
		source = e.MustGetFeature(args[1])
		i := source.Emit()
		raw := i.ToList()
		switch args[2] {
		case "normalize":
			numbers = normalizeAppend(raw, false)
		case "normalizeShuffle":
			numbers = normalizeAppend(raw, true)
		case "withWeights":
			wwArgs := argsMustBeLength(e, args, 3)
			numbers = withNumbersAppend(raw, wwArgs[3:])
		case "default":
			numbers = evenAppend(raw, 1)
		}
	default:
		numbers = args
	}

	cs := SplitStringChoices(numbers)

	return cs, numbers
}

type stringChoice struct {
	value  string
	weight int
}

func (sc stringChoice) Weight() int {
	return sc.weight
}

func (sc stringChoice) String() string {
	return sc.value
}

func (sc stringChoice) Int() int {
	return 0
}

func (sc stringChoice) Float() float64 {
	return 0.0
}

func SplitStringChoice(s string) stringChoice {
	var str string
	vals := strings.Split(s, "_")
	str = vals[0]
	var w int = 1
	if len(vals) > 1 {
		var err error
		w, err = strconv.Atoi(vals[len(vals)-1])
		if err != nil {
			w = 1
		}
	}
	return stringChoice{str, w}
}

func SplitStringChoices(s []string) choices {
	var ret choices
	for _, v := range s {
		ret = append(ret, SplitStringChoice(v))
	}
	return ret
}

func init() {
	WeightedString = feature.NewConstructor("WEIGHTED_STRING", 100, weightedString)
	feature.SetConstructor(WeightedString)
}
