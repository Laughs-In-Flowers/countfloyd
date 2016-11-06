package constructor

import (
	"fmt"
	"log"
	"math"
	mr "math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/davecgh/go-spew/spew"
)

func construct(
	from string,
	set []string,
	tag string,
	raw []string,
	values []string,
	efn feature.EmitFn,
	mfn feature.MapFn,
) (feature.Informer, feature.Emitter, feature.Mapper) {
	ss := []string{""}
	ss = append(ss, set...)
	return feature.NewInformer(from, ss, tag, raw, values),
		feature.NewEmitter(efn),
		feature.NewMapper(mfn)
}

func maybe(n float64) bool {
	maybe := mr.Float64()
	if maybe <= n {
		return true
	}
	return false
}

func mod(a, b int) int {
	return int(math.Mod(float64(a), float64(b)))
}

func listContains(item string, list []string) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}

func expandIntRange(s string) []string {
	var expanded []string
	if spl := strings.Split(s, "-"); len(spl) > 1 {
		start, err := strconv.ParseInt(spl[0], 10, 64)
		if err != nil {
			expanded = append(expanded, spl[0])
		}
		end, err := strconv.ParseInt(spl[1], 10, 64)
		if err != nil {
			expanded = append(expanded, spl[1])
		}
		for x := start; x <= end; x++ {
			num := strconv.FormatInt(x, 10)
			expanded = append(expanded, num)
		}
	}
	return expanded
}

func listExpand(in []string) []string {
	var expanded []string
	for _, v := range in {
		switch {
		case strings.Contains(v, "-"):
			x := expandIntRange(v)
			expanded = append(expanded, x...)
		default:
			expanded = append(expanded, v)
		}
	}
	return expanded
}

func listMappedToIntKeys(in []string) map[int]string {
	mapped := make(map[int]string)
	for i, v := range in {
		mapped[i+1] = v
	}
	return mapped
}

func intKeysToString(in map[int]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range in {
		ret[strconv.Itoa(k)] = v
	}
	return ret
}

func listMirror(in []string) []string {
	var nums []int
	var nstr []string
	for _, v := range in {
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

func shuffleStrings(s []string) {
	n := len(s)
	for i := n - 1; i > 0; i-- {
		j := mr.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

type tuple struct {
	one string
	two string
}

func (t *tuple) String() string {
	return fmt.Sprintf("%s_%s", t.one, t.two)
}

func (t *tuple) StringReverse() string {
	return fmt.Sprintf("%s_%s", t.two, t.one)
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

func normalizeWeighting(in []string, x ...string) []string {
	mean := float64(len(in) - 1/2)
	spew.Dump(mean)
	sd := .25 * float64(len(in))
	variance := math.Pow(float64(sd), 2)

	spew.Dump(sd, variance)

	for i, v := range in {
		w := math.Exp(-(math.Pow((float64(i)-mean), 2) / (2 * variance)))
		spew.Dump(w, v)
		//	round(max_weight * exp(-(i - mean)^2 / (2 * variance)))
	}

	var ret []string

	/*
		if shuffle {
			shuffleStrings(in)
		}

		l := len(in)
		switch {
		case mod(l, 3) == 0:
			l3 := (l / 3)
			var low, mid, high []string
			for i, v := range in {
				o := i + 1
				switch {
				case o <= l3:
					low = append(low, v)
				case o > l3 && o <= l3*2:
					mid = append(mid, v)
				case o >= l3*2:
					high = append(high, v)
				}
			}

			q1 := 1600 / len(low)
			q2 := 6800 / len(mid)
			q3 := 1600 / len(high)

			for _, v := range low {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q1))
			}

			for _, v := range mid {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q2))
			}

			for _, v := range high {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q3))
			}
		case mod(l, 2) == 0 && l >= 4:
			l4 := (l / 4)
			var first, second, third, fourth []string
			for i, v := range in {
				o := i + 1
				switch {
				case o <= l4:
					first = append(first, v)
				case o > l4 && o <= l4*2:
					second = append(second, v)
				case o > l4*2 && o <= l4*3:
					third = append(third, v)
				case o > l4*3:
					fourth = append(fourth, v)
				}
			}
			q1 := 1600 / len(first)
			q2 := 3400 / len(second)
			q3 := 3400 / len(third)
			q4 := 1600 / len(fourth)

			for _, v := range first {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q1))
			}

			for _, v := range second {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q2))
			}

			for _, v := range third {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q3))
			}

			for _, v := range fourth {
				ret = append(ret, fmt.Sprintf("%s_%d", v, q4))
			}
		default:
			var first, second, third []string
			q1 := (1 + math.Floor(float64(l)*.16))
			q3 := (float64(l) - math.Floor(float64(l)*.16))
			var o float64
			for i, v := range in {
				o = float64(i + 1)
				switch {
				case o <= q1:
					first = append(first, v)
				case o > q1 && o < q3:
					second = append(second, v)
				case o >= q3:
					third = append(third, v)
				}
			}
			lq1 := 1600 / len(first)
			lq2 := 3400 / len(second)
			lq3 := 1600 / len(third)
			for _, v := range first {
				ret = append(ret, fmt.Sprintf("%s_%d", v, lq1))
			}

			for _, v := range second {
				ret = append(ret, fmt.Sprintf("%s_%d", v, lq2))
			}

			for _, v := range third {
				ret = append(ret, fmt.Sprintf("%s_%d", v, lq3))
			}

		}
	*/
	return ret
}

func argsMustBeLength(e feature.Env, args []string, expects int) []string {
	length := len(args)
	if length < expects {
		log.Printf("provided args %s of length %d, expected at least %d", args, length, expects)
		os.Exit(-1)
	}
	return args
}

func init() {
	mr.Seed(time.Now().UTC().UnixNano())
}
