package constructor

import (
	cr "crypto/rand"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

type Choice struct {
	Weight int
	Value  interface{}
}

type Choices interface {
	Choose() (*Choice, error)
}

type choices struct {
	v []*Choice
}

var Unreachable = Crror("weighted choices error: unreachable")

func (cs *choices) Choose() (*Choice, error) {
	sum := 0
	for _, choice := range cs.v {
		sum += choice.Weight
	}
	r, err := intSpread(0, sum)
	if err != nil {
		return nil, err
	}
	for _, choice := range cs.v {
		r -= choice.Weight
		if r < 0 {
			return choice, nil
		}
	}
	return nil, Unreachable
}

var LargerMin = Crror("Min cannot be greater than max.")

func intSpread(min, max int) (int, error) {
	var result int
	switch {
	case min > max:
		return result, LargerMin
	case max == min:
		result = max
	case max > min:
		maxRand := max - min
		b, err := cr.Int(cr.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return result, err
		}
		result = min + int(b.Int64())
	}
	return result, nil
}

type wsp struct {
	group   []string
	raw     []string
	numbers []string
	ef      feature.EmitFn
	mf      feature.MapFn
}

type numbersFunc func([]string, ...string) []string

func wsParse(tag string,
	r *feature.RawFeature,
	e feature.Env,
	nfn numbersFunc) *wsp {
	raw := argsMustBeLength(e, r.MustGetValues(), 2)

	baseValues := baseValuesList(raw[0], e)

	numbers := nfn(baseValues, raw[1:]...)

	csr := SplitStringChoices(numbers)

	ef := weightedStringEmitFunction(tag, csr)

	mf := weightedStringMapFunction(tag, ef)

	return &wsp{r.Set, raw, numbers, ef, mf}
}

func weightedStringWith(
	from string,
	group []string,
	tag string,
	raw []string,
	values []string,
	ef func() data.Item,
	mf func(*data.Vector),
) (feature.Informer, feature.Emitter, feature.Mapper) {
	return construct(from, group, tag, raw, values, ef, mf)
}

func baseValuesList(f string, e feature.Env) []string {
	source := e.MustGetFeature(f)
	if i, err := source.EmitStrings(); err == nil {
		return i.ToStrings()
	}
	return nil
}

func weightedStringEmitFunction(tag string, csr Choices) feature.EmitFn {
	return func() data.Item {
		i := data.NewStringItem(tag, "")
		c, err := csr.Choose()
		if err != nil {
			i.SetString(err.Error())
		}
		if cs, ok := c.Value.(string); ok {
			i.SetString(cs)
		}
		return i
	}
}

func weightedStringMapFunction(tag string, fn feature.EmitFn) feature.MapFn {
	return func(d *data.Vector) {
		d.Set(fn())
	}
}

func levelWeighting(in []string, a string) []string {
	var ret []string
	for _, v := range in {
		ts := &tuple{v, a}
		ret = append(ret, ts.String())
	}
	return ret
}

func withNumbersWeighting(in []string, numbers ...string) []string {
	var ret []string

	switch {
	case len(numbers) == 1:
		ret = levelWeighting(in, numbers[0])
	default:
		var inner, outer []string
		var numbered string

		if len(numbers) > len(in) {
			outer = numbers
			inner = in
			numbered = "outer"
		} else {
			outer = in
			inner = numbers
			numbered = "inner"
		}

		li := len(inner)

		var hold []tuple

		for i, x := range outer {
			if i >= li {
				hold = append(hold, tuple{inner[mod(i, li)], x})
			} else {
				hold = append(hold, tuple{inner[i], x})
			}
		}

		for _, h := range hold {
			switch {
			case numbered == "outer":
				ret = append(ret, h.String())
			case numbered == "inner":
				ret = append(ret, h.StringReverse())
			}
		}
	}

	return ret
}

func WeightedStringWithWeights() feature.Constructor {
	return feature.NewConstructor(
		"WEIGHTED_STRING_WITH_WEIGHTS", 150, wsWithWeights,
	)
}

func wsWithWeights(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	wsp := wsParse(tag, r, e, withNumbersWeighting)
	return weightedStringWith("WEIGHTED_STRING_WITH_WEIGHTS",
		wsp.group,
		tag,
		wsp.raw,
		wsp.numbers,
		wsp.ef,
		wsp.mf,
	)
}

func normalizeWeighting(in []string, x ...string) []string {
	mean := float64(len(in) - 1/2)

	sd := .25 * float64(len(in))

	variance := math.Pow(float64(sd), 2)

	var ret []string
	for i, v := range in {
		w := math.Ceil(1000 * math.Exp(-(math.Pow((float64(i)-mean), 2) / (2 * variance))))
		ts := &tuple{v, strconv.Itoa(int(w))}
		ret = append(ret, ts.String())
	}

	return ret
}

func WeightedStringWithNormalizedWeights() feature.Constructor {
	return feature.NewConstructor(
		"WEIGHTED_STRING_WITH_NORMALIZED_WEIGHTS", 150, wsWithNormalizedWeights,
	)
}

func wsWithNormalizedWeights(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	wsp := wsParse(tag, r, e, normalizeWeighting)
	return weightedStringWith("WEIGHTED_STRING_WITH_NORMALIZED_WEIGHTS",
		wsp.group,
		tag,
		wsp.raw,
		wsp.numbers,
		wsp.ef,
		wsp.mf,
	)
}

func SplitStringChoice(s string) *Choice {
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
	return &Choice{w, str}
}

func SplitStringChoices(s []string) Choices {
	ret := make([]*Choice, 0)
	for _, v := range s {
		ret = append(ret, SplitStringChoice(v))
	}
	return &choices{ret}
}
