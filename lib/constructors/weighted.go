package constructors

import (
	cr "crypto/rand"
	"math/big"
)

type Choice interface {
	Weight() int
	Int() int
	Float() float64
	String() string
}

type Choices interface {
	Choose() (Choice, error)
}

type choices []Choice

func mkChoices(c ...Choice) choices {
	var ret choices
	for _, v := range c {
		ret = append(ret, v)
	}
	return ret
}

var Unreachable = Crror("weighted choices error: unreachable")

func (cs choices) Choose() (Choice, error) {
	sum := 0
	for _, choice := range cs {
		sum += choice.Weight()
	}
	r, err := intSpread(0, sum)
	if err != nil {
		return nil, err
	}
	for _, choice := range cs {
		r -= choice.Weight()
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
