package feature

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Laughs-In-Flowers/data"
	yaml "gopkg.in/yaml.v2"
)

func testConstructorString(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ckey := fmt.Sprintf("TEST_STRING:%s", tag)

	ef := func() data.Item {
		return data.NewStringItem(tag, ckey)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_STRING", r.Set, tag, r.Values, []string{ckey}),
		NewEmitter(ef),
		NewMapper(mf)
}

func testConstructorStrings(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ckey := fmt.Sprintf("TEST_STRINGS:%s", tag)
	ckeys := []string{ckey, ckey, ckey}

	ef := func() data.Item {
		return data.NewStringsItem(tag, ckeys...)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_STRINGS", r.Set, tag, r.Values, ckeys),
		NewEmitter(ef),
		NewMapper(mf)
}

func testConstructorBool(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ef := func() data.Item {
		return data.NewBoolItem(tag, false)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_BOOL", r.Set, tag, r.Values, []string{"false"}),
		NewEmitter(ef),
		NewMapper(mf)
}

func testConstructorInt(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	values := r.MustGetValues()
	v := values[0]
	vn, _ := strconv.Atoi(v)

	ef := func() data.Item {
		return data.NewIntItem(tag, vn)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_INT", r.Set, tag, r.Values, []string{"9000"}),
		NewEmitter(ef),
		NewMapper(mf)
}

func testConstructorFloat(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	values := r.MustGetValues()
	v := values[0]
	vn, _ := strconv.ParseFloat(v, 64)

	ef := func() data.Item {
		return data.NewFloatItem(tag, vn)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_FLOAT", r.Set, tag, r.Values, []string{"9000.0000001"}),
		NewEmitter(ef),
		NewMapper(mf)
}

func testConstructorMulti(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ckey := fmt.Sprintf("TEST_MULTI:%s", tag)

	ef := func() data.Item {
		d := data.New("")
		d.Set(data.NewStringItem("key", ckey))
		return data.NewMultiItem("multi", d)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_STRING", r.Set, tag, r.Values, []string{ckey}),
		NewEmitter(ef),
		NewMapper(mf)
}

type testFeature struct {
	Set    []string                                             `"yaml:set"`
	Tag    string                                               `"yaml:tag"`
	Apply  string                                               `"yaml:apply"`
	Values []string                                             `"yaml:values"`
	fn     func(*testing.T, *testFeature, Env, *data.Container) `"yaml:-"`
}

func getFeature(t *testing.T, e Env, f *testFeature) Feature {
	tag := strings.ToUpper(f.Tag)
	feature := e.GetFeature(tag)
	if feature == nil {
		t.Errorf("feature %s is nil", tag)
	}
	return feature
}

func listContains(s string, ss []string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func assertLength(t *testing.T, testKey string, have, expect []string) {
	l, le := len(have), len(expect)
	if l != le {
		t.Errorf("%s: expected length %d, but got %d", testKey, le, l)
	}
}

func assertIn(t *testing.T, testKey string, have, expect []string) {
	for _, v := range have {
		if !listContains(v, expect) {
			t.Errorf("%s: %s not in %v, but expected", testKey, v, expect)
		}
	}
}

func assertEqual(t *testing.T, testKey string, have, expect []string) {
	assertLength(t, testKey, have, expect)
	assertIn(t, testKey, have, expect)
}

var (
	loc string = "./features.yaml"

	additionalConstructor Constructor = NewConstructor("CONSTRUCTOR_STRINGS", 100, testConstructorStrings)

	testConstructors []Constructor = []Constructor{
		NewConstructor("CONSTRUCTOR_STRING", 47, testConstructorString),
		NewConstructor("CONSTRUCTOR_BOOL", 700, testConstructorBool),
		NewConstructor("CONSTRUCTOR_INT", 500, testConstructorInt),
		NewConstructor("CONSTRUCTOR_FLOAT", 2000, testConstructorFloat),
		DefaultConstructor("CONSTRUCTOR_MULTI", testConstructorMulti),
	}

	rawTestFeatures []*testFeature = []*testFeature{
		{nil, "feature-string", "constructor_string", []string{"TEST"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				si, err := feature.EmitString()
				if err != nil {
					t.Error(err)
				}
				have := []string{si.ToString()}
				expect := []string{"TEST_STRING:FEATURE-STRING"}
				assertEqual(t, "feature-string", have, expect)
			},
		},
		{nil, "feature-strings", "constructor_strings", []string{"TEST", "TEST", "TEST"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				si, err := feature.EmitStrings()
				if err != nil {
					t.Error(err)
				}
				have := si.ToStrings()
				expect := []string{"TEST_STRINGS:FEATURE-STRINGS", "TEST_STRINGS:FEATURE-STRINGS", "TEST_STRINGS:FEATURE-STRINGS"}
				assertEqual(t, "feature-strings", have, expect)
			},
		},
		{nil, "feature-bool", "constructor_bool", []string{"false"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				bi, err := feature.EmitBool()
				if err != nil {
					t.Error(err)
				}
				have := bi.ToBool()
				expect := false
				if have != expect {
					t.Errorf("feature-bool: have %t expected %t", have, expect)
				}
			},
		},
		{nil, "feature-int", "constructor_int", []string{"9000"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				ii, err := feature.EmitInt()
				if err != nil {
					t.Error(err)
				}
				have := ii.ToInt()
				expect := 9000
				if have != expect {
					t.Errorf("feature-int: have %d, expect %d", have, expect)
				}
			},
		},
		{nil, "feature-float", "constructor_float", []string{"9000.0000001"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				fi, err := feature.EmitFloat()
				if err != nil {
					t.Error(err)
				}
				have := fi.ToFloat()
				expect := 9000.0000001
				if have != expect {
					t.Errorf("feature-float: have %f, expect %f", have, expect)
				}
			},
		},
		{nil, "feature-multi", "constructor_multi", []string{"TEST", "TEST", "TEST"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				mi, err := feature.EmitMulti()
				if err != nil {
					t.Error(err)
				}
				c := mi.ToMulti()
				have := c.ToString("key")
				expect := "TEST_MULTI:FEATURE-MULTI"
				if have != expect {
					t.Errorf("feature-multi: have %s, expect %s", have, expect)
				}
			},
		},
	}

	rawWriteTestFeatures []*testFeature = []*testFeature{
		{[]string{"FILE"}, "feature-file-strings", "constructor_strings", []string{"A", "B", "C", "D", "E"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				si, err := feature.EmitStrings()
				if err != nil {
					t.Error(err)
				}
				have := si.ToStrings()
				expect := []string{"TEST_STRINGS:FEATURE-FILE-STRINGS", "TEST_STRINGS:FEATURE-FILE-STRINGS", "TEST_STRINGS:FEATURE-FILE-STRINGS"}
				assertEqual(t, "feature-file-strings", have, expect)
			},
		},
	}

	rawSetFeatures []*testFeature = []*testFeature{
		{[]string{"SET"}, "feature-set-strings", "constructor_strings", []string{"a", "b", "c", "4"},
			func(t *testing.T, f *testFeature, e Env, d *data.Container) {
				feature := getFeature(t, e, f)
				si, err := feature.EmitStrings()
				if err != nil {
					t.Error(err)
				}
				have := si.ToStrings()
				expect := []string{"TEST_STRINGS:FEATURE-SET-STRINGS", "TEST_STRINGS:FEATURE-SET-STRINGS", "TEST_STRINGS:FEATURE-SET-STRINGS"}
				assertEqual(t, "feature-set-strings", have, expect)
			},
		},
	}

	prePackedFeatureSet string
)

func allFeatures() []*testFeature {
	var ret []*testFeature
	ret = append(ret, rawTestFeatures...)
	ret = append(ret, rawWriteTestFeatures...)
	ret = append(ret, rawSetFeatures...)
	return ret
}

func testable(fs []*testFeature) ([]string, []*testFeature) {
	var reta []string
	var retb []*testFeature
	for _, f := range fs {
		if f.fn != nil {
			reta = append(reta, f.Tag)
			retb = append(retb, f)
		}
	}
	return reta, retb
}

func writeYaml(p string) error {
	f, err := data.Open(loc)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(&rawWriteTestFeatures)
	if err != nil {
		return err
	}

	f.Write(b)
	return nil
}

func deleteYaml(p string) {
	os.Remove(p)
}

type constructorSort []Constructor

func (c constructorSort) Len() int {
	return len(c)
}

func (c constructorSort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c constructorSort) Less(i, j int) bool {
	return c[i].Order() > c[j].Order()
}

func testOfConstructor(t *testing.T, e Env) {
	ck := "CONSTRUCTOR_STRINGS"
	_, exists := GetConstructor(ck)
	if !exists {
		t.Errorf("Constructor does not exist: %s", ck)
	}

	cl1 := ListConstructors()
	cl2 := e.ListConstructors()
	if len(cl1) != len(cl2) {
		t.Error("expected equal constructor lists")
	}
	csl1 := constructorSort(cl1)
	csl2 := constructorSort(cl2)
	sort.Sort(csl1)
	sort.Sort(csl2)
	l := csl1.Len() - 1
	for i := 0; i <= l; i++ {
		t1, t2 := csl1[i].Tag(), csl2[i].Tag()
		if t1 != t2 {
			t.Errorf("Constructor tags should be the same but are not: %s - %s", t1, t2)
		}
		o1, o2 := csl1[i].Order(), csl1[i].Order()
		if o1 != o2 {
			t.Errorf("Constructor order should be the same but are not: %d - %d", o1, o2)
		}
	}
}

func testOfFeature(t *testing.T, e Env) {
	testOfConstructor(t, e)

	f := e.MustGetFeature("feature-set-strings")

	var check []string
	check = append(check, f.Group()...)
	check = append(check, strconv.FormatBool(f.IsGroup("SET")))
	check = append(check, strconv.FormatBool(f.IsGroup("None")))
	check = append(check, f.From())
	check = append(check, f.Tag())
	check = append(check, f.Values()...)
	check = append(check, strconv.Itoa(f.Length()))
	check = append(check, f.Raw())

	expect := []string{"SET", "true", "false", "CONSTRUCTOR_STRINGS",
		"FEATURE-SET-STRINGS", "TEST_STRINGS:FEATURE-SET-STRINGS",
		"TEST_STRINGS:FEATURE-SET-STRINGS", "TEST_STRINGS:FEATURE-SET-STRINGS",
		"3", "a,b,c,4",
	}

	for _, v := range check {
		if !listContains(v, expect) {
			t.Errorf("feature value is unexpected: %s - %v")
		}
	}

	var check2 []error
	_, err := f.EmitString()
	check2 = append(check2, err)
	_, err = f.EmitBool()
	check2 = append(check2, err)
	_, err = f.EmitInt()
	check2 = append(check2, err)
	_, err = f.EmitFloat()
	check2 = append(check2, err)
	_, err = f.EmitMulti()
	check2 = append(check2, err)

	for _, er := range check2 {
		if er == nil {
			t.Error("Emit type provided error is nil: %v", check2)
		}
	}

	if err := e.SetFeature(&RawFeature{Tag: "feature-set-strings"}); err == nil {
		t.Error("Setting duplicate named feature did not return an error.")
	}
}

func testFeatureGroup(t *testing.T) {
	e := Empty()

	e.PopulateConstructors(testConstructors...)

	b, err := yaml.Marshal(&rawSetFeatures)
	if err != nil {
		t.Error(err)
	}
	e.Populate(b)

	testOfFeature(t, e)

	g1 := e.GetGroup("SET")
	g1v := g1.Value()
	g2, err := DecodeFeatureGroup(g1v)
	if err != nil {
		t.Error(err)
	}

	l1, l2, l3 := g1.List(), g2.List(), e.List("SET")
	lt1, lt2, lt3 := len(l1), len(l2), len(l3)
	if lt1 != lt2 || lt2 != lt3 || lt3 != lt1 {
		t.Errorf("group, decoded, and comparison group lengths are unequal: %d, %d, %d", lt1, lt2, lt3)
	}

	for _, v := range l1 {
		for _, vv := range l2 {
			if v.Tag != vv.Tag {
				t.Errorf("group feature and decoded group feature tags are not equal: %s - %s", v, vv)
			}
		}
	}

	prePackedFeatureSet = g1v
	g1, g2 = nil, nil
	e = nil
}

func testEnv(t *testing.T) Env {
	SetConstructor(additionalConstructor)

	testFeatureGroup(t)
	err := writeYaml(loc)
	if err != nil {
		t.Error(err)
	}
	defer deleteYaml(loc)

	b, err := yaml.Marshal(&rawTestFeatures)
	if err != nil {
		t.Error(err)
	}

	e, err := New(b, testConstructors...)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateYamlFiles(loc)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateGroup(prePackedFeatureSet)
	if err != nil {
		t.Error(err)
	}

	return e
}

func TestWhole(t *testing.T) {
	e := testEnv(t)

	a, f := testable(allFeatures())

	for h := 0; h <= 100; h++ {
		for i := 7; i <= 12; i++ {
			d := NewData(i)

			e.Apply(a, d)

			for _, ft := range f {
				ft.fn(t, ft, e, d)
			}
		}
	}

}
