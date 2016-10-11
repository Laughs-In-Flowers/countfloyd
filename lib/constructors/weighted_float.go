package constructors

import (
	"math"
	mr "math/rand"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var WeightedFloat feature.Constructor

func weightedFloat(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	cs, numbers := wfArgParser(e, r.MustGetValues())

	ef := func() *data.Item {
		i := data.NewItem(tag, "")
		choose, err := cs.Choose()
		if err != nil {
			i.SetString(err.Error())
			return i
		}
		i.SetFloat(choose.Float())
		return i
	}

	mf := func(d *feature.Data) {
		d.Set(ef())
	}

	return construct("WEIGHTED_FLOAT", r.Set, tag, numbers, numbers, ef, mf)
}

func wfArgParser(e feature.Env, args []string) (choices, []string) {
	args = argsMustBeLength(e, args, 3)

	var source feature.Feature
	var cs choices
	var numbers []string

	switch args[0] {
	case "SOURCED":
		source = e.MustGetFeature(args[1])
		i := source.Emit()
		raw := i.ToList()

		switch args[2] {
		case "normalize":
			numbers = normalizeAppend(raw, false)
		case "normalizeShuffle":
			numbers = normalizeAppend(raw, true)
		case "withWeights":
			numbers = withNumbersAppend(raw, args[3:])
		case "default":
			numbers = evenAppend(raw, 1)
		}

		switch args[3] {
		case "single":
			cs = FloatChoices(numbers...)
		case "range":
			cs = FloatRangeChoices(numbers...)
		}
	default:
		numbers = args
		l := len(strings.Split(numbers[0], "_"))
		switch l {
		case 2:
			cs = FloatChoices(numbers...)
		case 3:
			cs = FloatRangeChoices(numbers...)
		}
	}
	return cs, numbers
}

type floatchoice [2]float64

func (i floatchoice) Weight() int {
	return int(i[1])
}

func (i floatchoice) String() string {
	return strconv.FormatFloat(i.Float(), 'f', -1, 64)
}

func (i floatchoice) Int() int {
	return 0
}

func (i floatchoice) Float() float64 {
	return i[0]
}

func FloatChoice(n float64, w float64) floatchoice {
	return floatchoice{n, w}
}

func FloatChoices(floats ...string) choices {
	var ret choices
	var err error
	for _, i := range floats {
		var in, weight float64
		vals := strings.Split(i, "_")
		if len(vals) > 1 {
			in, err = strconv.ParseFloat(vals[0], 64)
			if err != nil {
				in = 0.0
			}
			weight, err = strconv.ParseFloat(vals[1], 64)
			if err != nil {
				weight = 1.0
			}
			ret = append(ret, FloatChoice(in, weight))
		}
	}
	return ret
}

type floatrangechoice [3]float64

func (f floatrangechoice) Weight() int {
	return int(f[2])
}

func modifier(m float64, prec int) float64 {
	var rounder float64
	intermed := m * math.Pow(10, float64(prec))

	if m >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / math.Pow(10, float64(prec))
}

func (f floatrangechoice) String() string {
	return strconv.FormatFloat(f.Float(), 'f', -1, 64)
}

func (f floatrangechoice) Int() int {
	return 0
}

func (f floatrangechoice) Float() float64 {
	min := f[0]
	max := f[1]
	return modifier(mr.Float64()*(max-min)+min, 1)
}

func FloatRangeChoice(min, max, weight float64) floatrangechoice {
	return [3]float64{min, max, weight}
}

func FloatRangeChoices(f ...string) choices {
	var ret choices
	for _, v := range f {
		spl := strings.Split(v, "_")
		min, err := strconv.ParseFloat(spl[0], 64)
		max, err := strconv.ParseFloat(spl[1], 64)
		weight, err := strconv.ParseFloat(spl[2], 64)
		if err == nil {
			ret = append(ret, FloatRangeChoice(min, max, weight))
		}
	}
	return ret
}

func init() {
	WeightedFloat = feature.NewConstructor("WEIGHTED_FLOAT", 102, weightedFloat)
	feature.SetConstructor(WeightedFloat)
}
