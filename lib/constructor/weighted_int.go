package constructor

import (
	"math"
	"strconv"
	"strings"

	mr "math/rand"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var WeightedInt feature.Constructor

func weightedInt(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	cs, numbers := wiArgParser(e, r.MustGetValues())

	ef := func() *data.Item {
		i := data.NewItem(tag, "")
		choose, err := cs.Choose()
		if err != nil {
			i.SetString(err.Error())
			return i
		}
		i.SetInt(choose.Int())
		return i
	}

	mf := func(d *feature.Data) {
		d.Set(ef())
	}

	return construct("WEIGHTED_INT", r.Set, tag, numbers, numbers, ef, mf)
}

func wiArgParser(e feature.Env, args []string) (choices, []string) {
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
			numbers = withNumbersAppend(raw, args[4:])
		case "default":
			numbers = evenAppend(raw, 1)
		}

		switch args[3] {
		case "single":
			cs = IntChoices(numbers...)
		case "range":
			cs = IntRangeChoices(numbers...)
		}
	default:
		numbers = args
		l := len(strings.Split(numbers[0], "_"))
		switch l {
		case 2:
			cs = IntChoices(numbers...)
		case 3:
			cs = IntRangeChoices(numbers...)
		}
	}

	return cs, numbers
}

type intchoice [2]int

func (i intchoice) Weight() int {
	return i[1]
}

func (i intchoice) String() string {
	return strconv.Itoa(i.Int())
}

func (i intchoice) Int() int {
	return i[0]
}

func (i intchoice) Float() float64 {
	return 0.0
}

func IntChoice(n int, w int) intchoice {
	return intchoice{n, w}
}

func IntChoices(ints ...string) choices {
	var ret choices
	var err error
	for _, i := range ints {
		var in, weight int
		vals := strings.Split(i, "_")
		if len(vals) > 1 {
			in, err = strconv.Atoi(vals[0])
			if err != nil {
				in = 0
			}
			weight, err = strconv.Atoi(vals[1])
			if err != nil {
				weight = 1
			}
			ret = append(ret, IntChoice(in, weight))
		}
	}
	return ret
}

type intrangechoice [3]int

func (i intrangechoice) Weight() int {
	return i[2]
}

func (i intrangechoice) String() string {
	return strconv.Itoa(i.Int())
}

func (i intrangechoice) Int() int {
	var ret int
	min := i[0]
	max := i[1]
	if min == 0 && max == 0 {
		ret = 0
	}
	if max < 0 {
		mx := int(math.Abs(float64(max)))
		mn := int(math.Abs(float64(min)))
		ret = ((mr.Intn(mx-mn) + 1) + mn) * -1
	}
	if max > 0 {
		ret = (mr.Intn(max-min) + 1) + min
	}
	return ret
}

func (i intrangechoice) Float() float64 {
	return 0.0
}

func IntRangeChoice(min, max, weight int) intrangechoice {
	return intrangechoice{min, max, weight}
}

func IntRangeChoices(i ...string) choices {
	var ret choices
	for _, v := range i {
		spl := strings.Split(v, "_")
		min, err := strconv.Atoi(spl[0])
		max, err := strconv.Atoi(spl[1])
		weight, err := strconv.Atoi(spl[2])
		if err == nil {
			ret = append(ret, IntRangeChoice(min, max, weight))
		}
	}
	return ret
}

func init() {
	WeightedInt = feature.NewConstructor("WEIGHTED_INT", 102, weightedInt)
	feature.SetConstructor(WeightedInt)
}
