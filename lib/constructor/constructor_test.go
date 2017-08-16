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

type testFeatureFn func(*testing.T, *testFeature, feature.Env, *data.Vector)

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

func (f *testFeature) compareStringFromData(t *testing.T, d *data.Vector, expect, unexpect []string) {
	dataKey := strings.ToUpper(f.Tag)
	have := []string{d.ToString(dataKey)}
	if expect != nil {
		assertIn(t, f, have, expect)
	}
	if unexpect != nil {
		assertNotIn(t, f, have, unexpect)
	}
}

func (f *testFeature) compareStringsFromData(t *testing.T, d *data.Vector, expect []string) {
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

func stringInVector(t *testing.T, e feature.Env, tf *testFeature, key string) string {
	feature := getFeature(t, e, tf)
	m, err := feature.EmitVector()
	if err != nil {
		t.Error(err)
	}
	mc := m.ToVector()
	return mc.ToString(key)
}

func (f *testFeature) compareStringsToFeatureValues(t *testing.T, e feature.Env, expect []string) {
	feature := getFeature(t, e, f)
	vs := feature.Values()
	assertLength(t, f, vs, expect)
	assertIn(t, f, vs, expect)
}

func (f *testFeature) compareMultipleStringsToFeatureValues(t *testing.T, e feature.Env, split string, expect1, expect2 []string) {
	feature := getFeature(t, e, f)
	have := feature.Values()
	for _, v := range have {
		spl := strings.Split(v, split)
		if len(spl) != 2 {
			t.Error("%s: unexpected split value in tested value", f.Apply)
		}
		assertIn(t, f, []string{spl[0]}, expect1)
		assertLength(t, f, have, expect1)
		assertIn(t, f, []string{spl[1]}, expect2)
		assertLength(t, f, have, expect2)
	}
}

func (f *testFeature) compareStringInVectorItem(t *testing.T, e feature.Env, key string, expects []string) {
	feature := getFeature(t, e, f)
	vi, err := feature.EmitVector()
	if err != nil {
		t.Error(err)
	}
	vii := vi.ToVector()
	have := vii.ToString(key)
	assertIn(t, f, []string{have}, expects)
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
		{nil, "list-a", "list", []string{"a", "b", "c", "d", "e"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"a", "b", "c", "d", "e"})
			},
		},
		{nil, "list-b", "list_with_null", []string{"list-a"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"a", "b", "c", "d", "e", "NULL"})
			},
		},
		{nil, "list-c", "list_shuffle", []string{"list-b"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				ft := getFeature(t, e, f)
				var o []bool
				unexpect := []string{"a", "b", "c", "d", "e", "NULL"}
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
		{nil, "exa", "list", []string{"1-10"}, nil},
		{nil, "expand-a", "list_expand_intrange", []string{"exa"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"})
			},
		},
		{nil, "exb", "list", []string{"1", "2", "5"}, nil},
		{nil, "expand-b", "list_expand_mirrorints", []string{"exb"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"-5", "-2", "-1", "1", "2", "5"})
			},
		},
		{nil, "collection-member-a", "collection_member", []string{"ace", "2-10", "jack", "queen", "king"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				have := d.ToString("COLLECTION-MEMBER-A")

				nv := d.ToInt("feature.priority")
				key := strconv.Itoa(nv)
				expect := stringInVector(t, e, f, key)

				if have != expect {
					t.Errorf("collection_member: have %s, expected %s", have, expect)
				}
			},
		},
		{nil, "fruits-a", "list", fruits1, nil},
		{nil, "fruits-b", "list_with_null", []string{"fruits-a"}, nil},
		{nil, "fruits-c", "list", fruits2, nil},
		{nil, "fruits-d", "list_with_null", []string{"fruits-c"}, nil},
		{nil, "combination-strings-a", "combination_strings", []string{"2", "false", "false", "fruits-b", "fruits-d"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromFeatureStrings(t, e, fruitsNoRepeat)
			},
		},
		{nil, "combination-strings-b", "combination_strings", []string{"2", "true", "true", "fruits-b", "fruits-d"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromFeatureStrings(t, e, fruitsRepeat)
			},
		},
		{nil, "select-combination-a", "weighted_string_with_weights", []string{"combination-strings-a", "x"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringFromData(t, d, fruitsNoRepeat, nil)
			},
		},
		{nil, "random-a", "simple_random", []string{"1", "on", "off"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringFromData(t, d, []string{"on", "off"}, []string{""})
			},
		},
		{nil, "random-b", "simple_random", []string{"0.5", "yes"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringFromData(t, d, []string{"yes", ""}, nil)
			},
		},
		{nil, "random-c", "sourced_random", []string{"1", "fruits-c"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringFromData(t, d, fruits2, nil)
			},
		},
		{nil, "weighted-string-a", "weighted_string_with_weights", []string{"list-b", "5", "500", "10"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expect := []string{"a_5", "b_500", "c_10", "d_5", "e_500", "NULL_10"}
				f.compareStringsToFeatureValues(t, e, expect)
			},
		},
		{nil, "weighted-string-b", "weighted_string_with_weights", []string{"list-b", "5", "500", "10", "1", "1000", "100", "9", "9", "9"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expect := []string{"a_5", "b_500", "c_10", "d_1", "e_1000", "NULL_100", "a_9", "b_9", "c_9"}
				f.compareStringsToFeatureValues(t, e, expect)
			},
		},
		{nil, "weighted-string-c", "weighted_string_with_normalized_weights", []string{"list-b", ""},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expect := []string{"a_1", "b_4", "c_29", "d_136", "e_412", "NULL_801"}
				f.compareStringsToFeatureValues(t, e, expect)
			},
		},
		{nil, "weighted-string-d", "weighted_string_with_normalized_weights", []string{"list-c", ""},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expect1 := []string{"a", "b", "c", "d", "e", "NULL"}
				expect2 := []string{"1", "4", "29", "136", "412", "801"}
				f.compareMultipleStringsToFeatureValues(t, e, "_", expect1, expect2)
			},
		},
		{nil, "set-a", "set", []string{"v1;random-a", "v2;random-c", "v3;weighted-string-c"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expects := []string{"a", "b", "c", "d", "e", "NULL"}
				f.compareStringInVectorItem(t, e, "set-a.v3", expects)
			},
		},
		{nil, "set-b", "set", []string{"v1;set-a", "v2;set-a", "v3;set-a"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expects := []string{"a", "b", "c", "d", "e", "NULL"}
				f.compareStringInVectorItem(t, e, "set-b.v3.set-a.v3", expects)
			},
		},
		{nil, "from-custom-constructor", "test_constructor", []string{"TEST", "TEST", "TEST"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expect := []string{"TEST", "TEST", "TEST"}
				f.compareStringsFromFeatureStrings(t, e, expect)
			},
		},
	}

	rawWriteTestFeatures []*testFeature = []*testFeature{
		{[]string{"FILE"}, "list-file-a", "list", []string{"A", "B", "C", "D", "E"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"A", "B", "C", "D", "E"})
			},
		},
		{[]string{"FILE"}, "list-file-b", "list_with_null", []string{"list-file-a"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"A", "B", "C", "D", "E", "NULL"})
			},
		},
	}

	rawSetFeatures []*testFeature = []*testFeature{
		{[]string{"SET"}, "derp", "list", []string{"a", "b", "c", "4"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				f.compareStringsFromData(t, d, []string{"a", "b", "c", "4"})
			},
		},
		{[]string{"SET"}, "weighted-string-in-set", "weighted_string_with_weights", []string{"derp", "1"},
			func(t *testing.T, f *testFeature, e feature.Env, d *data.Vector) {
				expect := []string{"a_1", "b_1", "c_1", "4_1"}
				f.compareStringsToFeatureValues(t, e, expect)
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

	mf := func(d *data.Vector) {
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

	err = e.SetConstructor(customConstructor)

	err = e.Populate(b)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateYaml(loc)
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
