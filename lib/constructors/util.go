package constructors

import (
	"fmt"
	"log"
	"math"
	mr "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

func construct(
	from string,
	set []string,
	tag string,
	raw []string,
	values []string,
	efn feature.EmitFn,
	mfn feature.MapFn) (feature.Informer, feature.Emitter, feature.Mapper) {
	if set != nil {
		set = append(set, "")
	}
	if set == nil {
		set = []string{""}
	}
	return feature.NewInformer(from, set, tag, raw, values),
		feature.NewEmitter(efn),
		feature.NewMapper(mfn)
}

func randRange(l int) int {
	return (mr.Intn(l-1) + 1)
}

func randInRange(min, max int) int {
	return mr.Intn(max-min) + min
}

func maybe(n float64) bool {
	maybe := mr.Float64()
	if maybe <= n {
		return true
	}
	return false
}

func remain(a, b int) int {
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

func listExpand(in []string) []string {
	var expanded []string
	for _, v := range in {
		if strings.Contains(v, "-") {
			if spl := strings.Split(v, "-"); len(spl) > 1 {
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
		} else {
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

func split(in string) (string, string) {
	spl := strings.Split(in, ":")
	return spl[0], spl[1]
}

func listSplit(l []string) map[string]string {
	ret := make(map[string]string)
	for _, v := range l {
		key, value := split(v)
		ret[key] = value
	}
	return ret
}

func valueFormat(in map[string]string, with interface{}) map[string]string {
	ret := make(map[string]string)
	for k, v := range in {
		ret[k] = fmt.Sprintf(v, with)
	}
	return ret
}

func stringToInt(in string) int {
	var ret int64
	var err error
	ret, err = strconv.ParseInt(in, 10, 64)
	if err != nil {
		ret = 1
	}
	return int(ret)
}

func stringToBool(in string) bool {
	var ret bool
	ret, _ = strconv.ParseBool(in)
	return ret
}

func shuffleStrings(s []string) {
	n := len(s)
	for i := n - 1; i > 0; i-- {
		j := mr.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

func semiColonReplace(in []string) []string {
	var ret []string
	for _, v := range in {
		ret = append(ret, strings.Replace(v, ";", ":", -1))
	}
	return ret
}

func boolOf(s string) bool {
	switch s {
	case "true", "1", "TRUE", "True", "yes", "YES":
		return true
	}
	return false
}

func evenAppend(in []string, i int) []string {
	var ret []string
	for _, v := range in {
		ret = append(ret, fmt.Sprintf("%s_%d", v, i))
	}
	return ret
}

type tuple struct {
	one string
	two string
}

func withNumbersAppend(in []string, numbers []string) []string {
	var ret []string

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
			hold = append(hold, tuple{inner[remain(i, li)], x})
		} else {
			hold = append(hold, tuple{inner[i], x})
		}
	}

	for _, h := range hold {
		if numbered == "outer" {
			ret = append(ret, fmt.Sprintf("%s_%s", h.one, h.two))
		}
		if numbered == "inner" {
			ret = append(ret, fmt.Sprintf("%s_%s", h.two, h.one))
		}
	}

	return ret
}

func normalizeAppend(in []string, shuffle bool) []string {
	var ret []string

	if shuffle {
		shuffleStrings(in)
	}

	l := len(in)
	switch {
	case math.Mod(float64(l), 3) == 0:
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
	case math.Mod(float64(l), 2) == 0 && l >= 4:
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
