package constructors_common

import (
	"fmt"
	"log"
	"math"
	mr "math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

func construct(
	from string,
	group []string,
	tag string,
	raw []string,
	values []string,
	efn feature.EmitFn,
	mfn feature.MapFn,
) (feature.Informer, feature.Emitter, feature.Mapper) {
	g := []string{""}
	g = append(g, group...)
	return feature.NewInformer(from, g, tag, raw, values),
		feature.NewEmitter(efn),
		feature.NewMapper(mfn)
}

type listModifier func(feature.CEnv, []string) []string

//var lmodify = []listModifier{
//	expander,
//	nullifier,
//	shuffler,
//	mirrorInts,
//}

func expander(e feature.CEnv, in []string) []string {
	exp := []string{}
	for _, k := range in {
		exp = append(exp, kToList(k, e)...)
	}
	return exp
}

var (
	numRangeRx = regexp.MustCompile("[0-9]+\\s*\\-\\s*[0-9]+")
	//letterRangeRx = regexp.MustCompile("\\W[a-zA-z]\\s*\\-\\s*\\W[a-zA-z]")
)

func kToList(k string, e feature.CEnv) []string {
	var ret []string
	var hit bool = false
	var comp = []func(){
		func() {
			if numRangeRx.MatchString(k) {
				ret = expandIntRange(k)
				hit = true
			}
		},
		//func() {
		//	if letterRangeRx.MatchString(k) {
		//		//letter range expansion
		//		hit = true
		//	}
		//},
		func() {
			ft := e.GetFeature(k)
			if ft != nil {
				ret = ftRecurse(ft, e)
				hit = true
			}
		},
		func() {
			fg := e.List(k)
			if len(fg) > 0 {
				for _, v := range fg {
					if ft := e.GetFeature(v.Tag); ft != nil {
						ret = append(ret, ftRecurse(ft, e)...)
					}
				}
				hit = true
			}
		},
		func() {
			if !hit {
				ret = append(ret, k)
			}
		},
	}

	for _, fn := range comp {
		if !hit {
			fn()
		}
	}

	return ret
}

func ftRecurse(f feature.Feature, e feature.CEnv) []string {
	var ret []string
	if s, err := f.EmitString(); err == nil {
		ret = append(ret, s.ToString())
	}
	if l, err := f.EmitStrings(); err == nil {
		v := l.ToStrings()
		for _, vv := range v {
			ret = append(ret, kToList(vv, e)...)
		}
	}
	return ret
}

func nullifier(val string) func(e feature.CEnv, ss []string) []string {
	return func(e feature.CEnv, ss []string) []string {
		ss = append(ss, val)
		return ss
	}
}

func shuffler(e feature.CEnv, ss []string) []string {
	shuffleStrings(ss)
	return ss
}

func mirrorInts(e feature.CEnv, ss []string) []string {
	var nums []int
	var nstr []string
	for _, v := range ss {
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

func expandLetterRange(start, end string) []string {
	return nil
}

func listMappedToFloat64Keys(in []string, step float64) map[float64]string {
	i := float64(0)
	mapped := make(map[float64]string)
	for _, v := range in {
		mapped[i] = v
		i = i + step
	}
	return mapped
}

func floatKeysToString(in map[float64]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range in {
		ret[strconv.FormatFloat(k, 'f', 0, 64)] = v
	}
	return ret
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

func argsMustBeLength(e feature.CEnv, args []string, expects int) []string {
	length := len(args)
	if length < expects {
		log.Printf("provided args %s of length %d, expected at least %d", args, length, expects)
		os.Exit(-1)
	}
	return args
}

type kf struct {
	k string
	f feature.Feature
}

func extractSoloKFS(e feature.CEnv, tag string, r string) []*kf {
	var ret []*kf
	spl := strings.Split(r, ";")
	ll := len(spl)
	switch {
	case ll == 1:
		k := spl[0]
		if f := e.GetFeature(k); f != nil {
			ret = append(ret, &kf{k, f})
		}
		gs := e.List(k)
		if len(gs) > 0 {
			for _, rf := range gs {
				if f := e.GetFeature(rf.Tag); f != nil {
					ret = append(ret, &kf{rf.Tag, f})
				}
			}
		}
	case ll == 2:
		k := strings.Join([]string{tag, spl[0]}, ".")
		f := e.GetFeature(spl[1])
		ret = append(ret, &kf{k, f})
	default:
		return nil
	}
	return ret
}

func extractKFS(e feature.CEnv, tag string, r ...string) []*kf {
	var ret []*kf
	for _, v := range r {
		x := extractSoloKFS(e, tag, v)
		for _, v := range x {
			if v.f != nil {
				ret = append(ret, x...)
			}
		}
	}
	return ret
}

func smoothKey(nk ...string) string {
	var xp []string
	for _, v := range nk {
		spl := strings.Split(v, ".")
		xp = append(xp, spl...)
	}
	var cp []string
	for i, v := range xp {
		if i != len(xp)-1 {
			if v == xp[i+1] {
				continue
			}
		}
		if v == "" {
			continue
		}
		cp = append(cp, v)
	}
	return strings.Join(cp, ".")
}

func init() {
	mr.Seed(time.Now().UTC().UnixNano())
}
