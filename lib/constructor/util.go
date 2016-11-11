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
