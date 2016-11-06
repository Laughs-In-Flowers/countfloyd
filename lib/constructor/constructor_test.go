package constructor

import (
	"os"
	"strconv"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

type testFeatureFn func(*testing.T, *testFeature, feature.Env, *data.Container)

type testFeature struct {
	Set    []string      `"yaml:set"`
	Tag    string        `"yaml:tag"`
	Apply  string        `"yaml:apply"`
	Values []string      `"yaml:values"`
	fn     testFeatureFn `"yaml:-"`
}

func assertLength(t *testing.T, f *testFeature, have, expect []string) {
	l, le := len(have), len(expect)
	if l != le {
		t.Errorf("%s:%s: expected length %d, but got %d", f.Tag, f.Apply, le, l)
	}
}

func assertIn(t *testing.T, f *testFeature, have, expect []string) {
	for _, v := range have {
		if !listContains(v, expect) {
			t.Errorf("%s:%s: %s not in %v, but expected", f.Tag, f.Apply, v, expect)
		}
	}
}

func assertNotIn(t *testing.T, f *testFeature, have, unexpected []string) {
	for _, v := range have {
		if listContains(v, unexpected) {
			t.Errorf("%s:%s: %s is in %v, but should not be", f.Tag, f.Apply, v, unexpected)
		}
	}
}

func orderDifference(have, unexpect []string) bool {
	for i, _ := range have {
		if have[i] != unexpect[i] {
			return true
		}
	}
	return false
}

func (f *testFeature) compareStringFromData(t *testing.T, d *data.Container, expect, unexpect []string) {
	dataKey := strings.ToUpper(f.Tag)
	have := []string{d.ToString(dataKey)}
	if expect != nil {
		assertIn(t, f, have, expect)
	}
	if unexpect != nil {
		assertNotIn(t, f, have, unexpect)
	}
}

func (f *testFeature) compareStringsFromData(t *testing.T, d *data.Container, expect []string) {
	dataKey := strings.ToUpper(f.Tag)
	have := d.ToStrings(dataKey)
	assertLength(t, f, have, expect)
	assertIn(t, f, have, expect)
}

func getFeature(t *testing.T, e feature.Env, tf *testFeature) feature.Feature {
	tag := strings.ToUpper(tf.Tag)
	feature := e.GetFeature(tag)
	if feature == nil {
		t.Errorf("feature %s is nil", tag)
	}
	return feature
}

func stringInMulti(t *testing.T, e feature.Env, tf *testFeature, key string) string {
	feature := getFeature(t, e, tf)
	m, err := feature.EmitMulti()
	if err != nil {
		t.Error(err)
	}
	mc := m.ToMulti()
	return mc.ToString(key)
}

func (f *testFeature) compareStringsFromFeatureStrings(t *testing.T, e feature.Env, expect []string) {
	feature := getFeature(t, e, f)
	si, err := feature.EmitStrings()
	if err != nil {
		t.Error(err)
	}
	have := si.ToStrings()
	assertLength(t, f, have, expect)
	assertIn(t, f, have, expect)
}

func (f *testFeature) compareStringsFromFeatureStringsSplit(t *testing.T, e feature.Env, split string, expect1, expect2 []string) {
	testKey := f.Apply
	feature := getFeature(t, e, f)
	si, err := feature.EmitStrings()
	if err != nil {
		t.Error(err)
	}
	have := si.ToStrings()
	for _, v := range have {
		spl := strings.Split(v, split)
		if len(spl) != 2 {
			t.Error("%s: unexpected split value in tested value", testKey)
		}
		assertIn(t, f, []string{spl[0]}, expect1)
		assertLength(t, f, have, expect1)
		assertIn(t, f, []string{spl[1]}, expect2)
		assertLength(t, f, have, expect2)
	}
}

var (
	loc string = "./features.yaml"

	customConstructor = feature.DefaultConstructor("TEST_CONSTRUCTOR", customConstructorFn)

	fruits1 = []string{"apple", "orange"}

	fruits2 = []string{"grapes", "pear", "banana"}

	fruitsNoRepeat = []string{
		"apple+grapes", "apple+pear", "apple+banana", "apple+NULL",
		"orange+grapes", "orange+pear", "orange+banana", "orange+NULL",
		"NULL+grapes", "NULL+pear", "NULL+banana", "NULL+NULL",
	}

	fruitsRepeat = []string{
		"apple+apple", "apple+orange", "apple+NULL",
		"apple+grapes", "apple+pear", "apple+banana", "apple+NULL",
		"orange+apple", "orange+orange", "orange+NULL", "orange+grapes",
		"orange+pear", "orange+banana", "NULL+apple", "NULL+orange",
		"NULL+NULL", "NULL+grapes", "NULL+pear", "grapes+apple",
		"grapes+orange", "grapes+NULL", "grapes+grapes", "pear+apple",
		"pear+orange", "pear+NULL", "banana+apple", "banana+orange",
		"NULL+apple",
	}

	rawTestFeatures []*testFeature = []*testFeature{
		{nil, "direct-a", "direct", []string{"a", "b", "c", "d", "e"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"a", "b", "c", "d", "e"})
			},
		},
		{nil, "direct-b", "direct_null", []string{"direct-a"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"a", "b", "c", "d", "e", "NULL"})
			},
		},
		{nil, "direct-c", "direct_shuffle", []string{"direct-a"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				ft := getFeature(t, e, f)
				var o []bool
				unexpect := []string{"a", "b", "c", "d", "e"}
				for i := 1; i <= 100; i++ {
					si, err := ft.EmitStrings()
					if err != nil {
						t.Error(err)
					}
					have := si.ToStrings()
					o = append(o, orderDifference(have, unexpect))
				}
				oc := make(map[bool]int)
				for _, v := range o {
					oc[v] = oc[v] + 1
				}
				if oc[false] >= 10 {
					t.Errorf("%s: Expected same order percentage where different order is expected exceeds 10%", f.Tag)
				}
			},
		},
		{nil, "collection-member-a", "collection_member", []string{"ace", "2-10", "jack", "queen", "king"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				have := d.ToString("COLLECTION-MEMBER-A")

				nv := d.ToInt("feature.priority")
				key := strconv.Itoa(nv)
				expect := stringInMulti(t, e, f, key)

				if have != expect {
					t.Errorf("collection_member: have %s, expected %s", have, expect)
				}
			},
		},
		{nil, "list-a", "direct", fruits1, nil},
		{nil, "list-b", "direct_null", []string{"list-a"}, nil},
		{nil, "list-c", "direct", fruits2, nil},
		{nil, "list-d", "direct_null", []string{"list-c"}, nil},
		{nil, "combination-strings-a", "combination_strings", []string{"2", "false", "false", "list-b", "list-d"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromFeatureStrings(t, e, fruitsNoRepeat)
			},
		},
		{nil, "combination-strings-b", "combination_strings", []string{"2", "true", "true", "list-b", "list-d"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromFeatureStrings(t, e, fruitsRepeat)
			},
		},
		//{nil, "select-combination-a", "same_weighted_string", []string{"combination-strings-a", "1"},
		//	func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
		//		f.compareStringFromData(t, d, fruitsNoRepeat, nil)
		//	},
		//},
		//{"combination-int-a", "combination_int", []string{}, nil},
		{nil, "exa", "direct", []string{"1-10"}, nil},
		{nil, "expand-ints-a", "expand_ints", []string{"exa", "range"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})
			},
		},
		{nil, "exb", "direct", []string{"1", "2", "5"}, nil},
		{nil, "expand-ints-b", "expand_ints", []string{"exb", "mirror"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"-5", "-2", "-1", "1", "2", "5"})
			},
		},
		{nil, "random-a", "simple_random", []string{"1", "on", "off"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringFromData(t, d, []string{"on", "off"}, []string{""})
			},
		},
		{nil, "random-b", "simple_random", []string{"0.5", "yes"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringFromData(t, d, []string{"yes", ""}, nil)
			},
		},
		{nil, "random-c", "sourced_random", []string{"1", "list-c"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringFromData(t, d, fruits2, nil)
			},
		},
		//{nil, "weighted-string-a", "weighted_string", []string{"SOURCED", "direct-b", "normalize"},
		//	func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
		//		expect := []string{"a_800", "b_800", "c_3400", "d_3400", "e_800", "NULL_800"}
		//		f.compareStringsFromFeatureStrings(t, e, expect)
		//	},
		//},
		//{nil, "weighted-string-b", "weighted_string", []string{"SOURCED", "direct-b", "normalizeShuffle"},
		//	func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
		//		expect1 := []string{"a", "b", "c", "d", "e", "NULL"}
		//		expect2 := []string{"800", "800", "3400", "3400", "800", "800"}
		//		f.compareStringsFromFeatureStringsSplit(t, e, "_", expect1, expect2)
		//	},
		//},
		{nil, "weighted-string-c", "weighted_string_with_weights", []string{"direct-b", "5", "500", "10"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				expect := []string{"a_5", "b_500", "c_10", "d_5", "e_500", "NULL_10"}
				f.compareStringsFromFeatureStrings(t, e, expect)
			},
		},
		{nil, "weighted-string-d", "weighted_string_with_weights", []string{"direct-b", "1"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				expect := []string{"a_1", "b_1", "c_1", "d_1", "e_1", "NULL_1"}
				f.compareStringsFromFeatureStrings(t, e, expect)
			},
		},
		//{"", "weighted_int", []string{}, nil},
		//{"", "weighted_float", []string{}, nil},
		//{"", "set", []string{},nil},
		//{"", "set", []string{},nil},
		//{"", "set", []string{}, nil},
		{nil, "from-custom-constructor", "test_constructor", []string{"TEST", "TEST", "TEST"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				expect := []string{"TEST", "TEST", "TEST"}
				f.compareStringsFromFeatureStrings(t, e, expect)
			},
		},
	}

	rawWriteTestFeatures []*testFeature = []*testFeature{
		{[]string{"FILE"}, "direct-file-a", "direct", []string{"A", "B", "C", "D", "E"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"A", "B", "C", "D", "E"})
			},
		},
		{[]string{"FILE"}, "direct-file-d", "direct_null", []string{"direct-file-a"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"A", "B", "C", "D", "E", "NULL"})
			},
		},
	}

	rawSetFeatures []*testFeature = []*testFeature{
		{[]string{"SET"}, "direct-in-set", "direct", []string{"a", "b", "c", "4"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				f.compareStringsFromData(t, d, []string{"a", "b", "c", "4"})
			},
		},
		{[]string{"SET"}, "weighted-string-in-set", "weighted_string_with_weights", []string{"direct-in-set", "1"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Container) {
				expect := []string{"a_1", "b_1", "c_1", "4_1"}
				f.compareStringsFromFeatureStrings(t, e, expect)
			},
		},
	}

	featureGroup string
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

func customConstructorFn(tag string, r *feature.RawFeature, e feature.Env) (feature.Informer, feature.Emitter, feature.Mapper) {
	list := r.MustGetValues()

	ef := func() data.Item {
		return data.NewStringsItem(tag, list...)
	}

	mf := func(d *data.Container) {
		d.Set(ef())
	}

	return construct("TEST_CONSTRUCTOR", r.Set, tag, list, list, ef, mf)
}

func createGroupValue(t *testing.T) {
	e := feature.Empty()

	b, err := yaml.Marshal(&rawSetFeatures)
	if err != nil {
		t.Error(err)
	}
	e.Populate(b)

	g := e.GetGroup("SET")
	gv := g.Value()
	featureGroup = gv
	e = nil
}

func testEnv(t *testing.T) feature.Env {
	createGroupValue(t)

	err := writeYaml(loc)
	if err != nil {
		t.Error(err)
	}
	defer deleteYaml(loc)

	e := feature.Empty()

	b, err := yaml.Marshal(&rawTestFeatures)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateConstructors(customConstructor)

	err = e.Populate(b)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateYamlFiles(loc)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateGroup(featureGroup)
	if err != nil {
		t.Error(err)
	}

	return e
}

func TestConstructors(t *testing.T) {
	e := testEnv(t)

	a, f := testable(allFeatures())

	for h := 0; h <= 100; h++ {
		for i := 7; i <= 12; i++ {
			d := feature.NewData(i)

			e.Apply(a, d)

			for _, ft := range f {
				ft.fn(t, ft, e, d)
			}
		}
	}
}
