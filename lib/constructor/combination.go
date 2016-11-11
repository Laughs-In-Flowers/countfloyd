package constructor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
)

type Replacer interface {
	Len() int
	Replace([]int) Replacer
	Value() string
	ValueNoSame() (string, bool)
}

type StringReplacer []string

func (sl StringReplacer) Len() int {
	return len(sl)
}

func (sl StringReplacer) Replace(indices []int) Replacer {
	result := make(StringReplacer, len(indices), len(indices))
	for i, idx := range indices {
		result[i] = sl[idx]
	}
	return result
}

func (sl StringReplacer) Value() string {
	var spl []string
	for _, v := range sl {
		s := strings.Split(v, ":")
		spl = append(spl, s[0])
	}
	return strings.Join(spl, "+")
}

func (sl StringReplacer) ValueNoSame() (string, bool) {
	var spl []string
	var curr string
	for _, v := range sl {
		s := strings.Split(v, ":")
		if s[1] == curr {
			return "", false
		}
		spl = append(spl, s[0])
		curr = s[1]
	}
	return strings.Join(spl, "+"), true
}

func Combinations(list Replacer, selectNum int, repeatable bool, buf int) (c chan Replacer) {
	c = make(chan Replacer, buf)
	index := make([]int, list.Len(), list.Len())
	for i := 0; i < list.Len(); i++ {
		index[i] = i
	}

	var comb_generator func([]int, int, int) chan []int
	if repeatable {
		comb_generator = repeated_combinations
	} else {
		comb_generator = combinations
	}

	go func() {
		defer close(c)
		for comb := range comb_generator(index, selectNum, buf) {
			c <- list.Replace(comb)
		}
	}()

	return
}

func combinations(list []int, select_num, buf int) (c chan []int) {
	c = make(chan []int, buf)
	go func() {
		defer close(c)
		switch {
		case select_num == 0:
			c <- []int{}
		case select_num == len(list):
			c <- list
		case len(list) < select_num:
			return
		default:
			for i := 0; i < len(list); i++ {
				for sub_comb := range combinations(list[i+1:], select_num-1, buf) {
					c <- append([]int{list[i]}, sub_comb...)
				}
			}
		}
	}()
	return
}

func repeated_combinations(list []int, select_num, buf int) (c chan []int) {
	c = make(chan []int, buf)
	go func() {
		defer close(c)
		if select_num == 1 {
			for v := range list {
				c <- []int{v}
			}
			return
		}
		for i := 0; i < len(list); i++ {
			for sub_comb := range repeated_combinations(list[i:], select_num-1, buf) {
				c <- append([]int{list[i]}, sub_comb...)
			}
		}
	}()
	return
}

func CombinationStrings() feature.Constructor {
	return feature.NewConstructor("COMBINATION_STRINGS", 90, combinationOfStrings)
}

func combinationOfStrings(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	rp, num, repeat, same, buf := csArgParser(e, r.MustGetValues())

	var vals []string
	for c := range Combinations(rp, num, repeat, buf) {
		if !same {
			if val, ok := c.ValueNoSame(); ok {
				vals = append(vals, val)
			}
		}
		if same {
			vals = append(vals, c.Value())
		}
	}

	return listFrom("COMBINATION_STRINGS", r.Set, tag, r.Values, vals, e)
}

func csArgParser(e feature.Env, args []string) (Replacer, int, bool, bool, int) {
	args = argsMustBeLength(e, args, 4)
	var err error

	rnum := args[0]
	var num int
	num, err = strconv.Atoi(rnum)
	if err != nil {
		num = 1
	}

	rrepeat := args[1]
	var repeat bool
	repeat, err = strconv.ParseBool(rrepeat)
	if err != nil {
		repeat = false
	}

	rsame := args[2]
	var same bool
	same, err = strconv.ParseBool(rsame)
	if err != nil {
		same = false
	}

	from := args[3:]
	var all []string
	for n, v := range from {
		f := e.MustGetFeature(v)
		if i, err := f.EmitStrings(); err == nil {
			l := i.ToStrings()
			var nl []string
			for _, ll := range l {
				nl = append(nl, fmt.Sprintf("%s:%d", ll, n))
			}
			all = append(all, nl...)
		}
	}
	rp := StringReplacer(all)

	buf := 2 * len(from)

	return rp, num, repeat, same, buf
}
