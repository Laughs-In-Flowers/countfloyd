package constructor

import (
	cr "crypto/rand"
	"math/big"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

var (
	WeightedStringWithWeights           feature.Constructor
	WeightedStringWithNormalizedWeights feature.Constructor
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

	ef := weightedStringEmitFunction(tag, numbers)

	csr := SplitStringChoices(numbers)

	mf := weightedStringMapFunction(tag, csr)

	return &wsp{r.Set, raw, numbers, ef, mf}
}

func weightedStringWith(
	from string,
	group []string,
	tag string,
	raw []string,
	values []string,
	ef func() data.Item,
	mf func(*data.Container),
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

func weightedStringEmitFunction(tag string, numbers []string) feature.EmitFn {
	return func() data.Item {
		return data.NewStringsItem(tag, numbers...)
	}
}

func weightedStringMapFunction(tag string, csr Choices) feature.MapFn {
	return func(d *data.Container) {
		i := data.NewStringItem(tag, "")
		c, err := csr.Choose()
		if err != nil {
			i.SetString(err.Error())
		}
		if cs, ok := c.Value.(string); ok {
			i.SetString(cs)
			d.Set(i)
		}
	}
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

func init() {
	WeightedStringWithWeights = feature.NewConstructor(
		"WEIGHTED_STRING_WITH_WEIGHTS", 101, wsWithWeights,
	)
	WeightedStringWithNormalizedWeights = feature.NewConstructor(
		"WEIGHTED_STRING_WITH_NORMALIZED_WEIGHTS", 102, wsWithNormalizedWeights,
	)
	feature.SetConstructor(
		WeightedStringWithWeights,
		WeightedStringWithNormalizedWeights,
	)
}
