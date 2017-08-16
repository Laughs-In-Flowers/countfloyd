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

func stringKey(pre, tag string) string {
	PRE, TAG := strings.ToUpper(pre), strings.ToUpper(tag)
	return fmt.Sprintf("%s:%s", PRE, TAG)
}

func tConstructorString(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ckey := stringKey("test_string", tag)

	ef := func() data.Item {
		return data.NewStringItem(tag, ckey)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_STRING", r.Set, tag, r.Values, []string{ckey}),
		NewEmitter(ef),
		NewMapper(mf)
}

func stringsKeys(l int, pre, tag string) []string {
	var ret []string
	key := stringKey(pre, tag)
	for i := 1; i <= l; i++ {
		ret = append(ret, key)
	}
	return ret
}

func tConstructorStrings(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ckeys := stringsKeys(3, "test_strings", tag)

	ef := func() data.Item {
		return data.NewStringsItem(tag, ckeys...)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_STRINGS", r.Set, tag, r.Values, ckeys),
		NewEmitter(ef),
		NewMapper(mf)
}

func tConstructorBool(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ef := func() data.Item {
		return data.NewBoolItem(tag, false)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_BOOL", r.Set, tag, r.Values, []string{"false"}),
		NewEmitter(ef),
		NewMapper(mf)
}

func tConstructorInt(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	values := r.MustGetValues()
	v := values[0]
	vn, _ := strconv.Atoi(v)

	ef := func() data.Item {
		return data.NewIntItem(tag, vn)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_INT", r.Set, tag, r.Values, []string{"9000"}),
		NewEmitter(ef),
		NewMapper(mf)
}

func tConstructorFloat(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	values := r.MustGetValues()
	v := values[0]
	vn, _ := strconv.ParseFloat(v, 64)

	ef := func() data.Item {
		return data.NewFloatItem(tag, vn)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_FLOAT", r.Set, tag, r.Values, []string{"9000.0000001"}),
		NewEmitter(ef),
		NewMapper(mf)
}

func tConstructorVector(tag string, r *RawFeature, e Env) (Informer, Emitter, Mapper) {
	ckey := stringKey("test_vector", tag)

	ef := func() data.Item {
		d := data.New("")
		d.Set(data.NewStringItem("key", ckey))
		return data.NewVectorItem("vector", d)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return NewInformer("CONSTRUCTOR_VECTOR", r.Set, tag, r.Values, []string{ckey}),
		NewEmitter(ef),
		NewMapper(mf)
}

type tFeature struct {
	Set    []string                                       `"yaml:set"`
	Tag    string                                         `"yaml:tag"`
	Apply  string                                         `"yaml:apply"`
	Values []string                                       `"yaml:values"`
	fn     func(*testing.T, *tFeature, Env, *data.Vector) `"yaml:-"`
}

type tComponent struct {
	Tag      string                             `"yaml:tag"`
	Defines  []*tFeature                        `"yaml:defines"`
	Features []*tFeature                        `"yaml:features"`
	fn       func(*testing.T, *tComponent, Env) `"yaml:-"`
}

func (tc *tComponent) featureKeys() []string {
	var ret []string
	for _, v := range tc.Features {
		ret = append(ret, strings.ToUpper(v.Tag))
	}
	return ret
}

type tEntity struct {
	Tag        string                          `"yaml:tag"`
	Defines    []*tFeature                     `"yaml:defines"`
	Components []*tComponent                   `"yaml:features"`
	fn         func(*testing.T, *tEntity, Env) `"yaml:-"`
}

func (te *tEntity) componentFeatureKeys(c string) []string {
	var ret []string
	for _, v := range te.Components {
		if c == v.Tag {
			ret = v.featureKeys()
		}
	}
	return ret
}

func getFeature(t *testing.T, e Env, f *tFeature) Feature {
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

func loc(i int) string {
	return fmt.Sprintf("./features-%d.yaml", i)
}

var (
	additionalConstructor Constructor = NewConstructor("CONSTRUCTOR_STRINGS", 100, tConstructorStrings)

	testConstructors []Constructor = []Constructor{
		NewConstructor("CONSTRUCTOR_STRING", 47, tConstructorString),
		NewConstructor("CONSTRUCTOR_BOOL", 700, tConstructorBool),
		NewConstructor("CONSTRUCTOR_INT", 500, tConstructorInt),
		NewConstructor("CONSTRUCTOR_FLOAT", 2000, tConstructorFloat),
		DefaultConstructor("CONSTRUCTOR_VECTOR", tConstructorVector),
	}

	fna = func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
		feature := getFeature(t, e, f)
		si, err := feature.EmitString()
		if err != nil {
			t.Error(err)
		}
		sit := si.ToString()
		have := []string{sit, sit}
		sk := stringKey("test_string", f.Tag)
		sd := d.ToString(strings.ToUpper(f.Tag))
		expect := []string{sk, sd}
		assertEqual(t, "feature-string", have, expect)
	}

	fnb = func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
		feature := getFeature(t, e, f)
		si, err := feature.EmitStrings()
		if err != nil {
			t.Error(err)
		}

		have := si.ToStrings()

		expect1 := stringsKeys(3, "test_strings", f.Tag)
		assertEqual(t, "feature-strings", have, expect1)
		expect2 := d.ToStrings(strings.ToUpper(f.Tag))
		assertEqual(t, "feature-strings", have, expect2)
	}

	fnc = func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
		feature := getFeature(t, e, f)
		bi, err := feature.EmitBool()
		if err != nil {
			t.Error(err)
		}
		have := bi.ToBool()
		expect := d.ToBool(strings.ToUpper(f.Tag))

		if have != expect {
			t.Errorf("feature-bool: have %t expected %t", have, expect)
		}
	}

	fnd = func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
		feature := getFeature(t, e, f)
		ii, err := feature.EmitInt()
		if err != nil {
			t.Error(err)
		}
		have := ii.ToInt()
		expect := d.ToInt(strings.ToUpper(f.Tag))
		if have != expect {
			t.Errorf("feature-int: have %d, expect %d", have, expect)
		}
	}

	fne = func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
		feature := getFeature(t, e, f)
		fi, err := feature.EmitFloat()
		if err != nil {
			t.Error(err)
		}
		have := fi.ToFloat()
		expect := d.ToFloat(strings.ToUpper(f.Tag))
		if have != expect {
			t.Errorf("feature-float: have %f, expect %f", have, expect)
		}
	}

	fnf = func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
		feature := getFeature(t, e, f)
		mi, err := feature.EmitVector()
		if err != nil {
			t.Error(err)
		}
		c := mi.ToVector()
		have := c.ToString("key")
		expect := stringKey("test_vector", f.Tag)
		if have != expect {
			t.Errorf("feature-vector: have %s, expect %s", have, expect)
		}
	}

	rawTestFeatures []*tFeature = []*tFeature{
		{nil,
			"feature-string",
			"constructor_string",
			[]string{"TEST"},
			fna,
		},
		{nil,
			"feature-strings",
			"constructor_strings",
			[]string{"TEST", "TEST", "TEST"},
			fnb,
		},
		{nil,
			"feature-bool",
			"constructor_bool",
			[]string{"false"},
			fnc,
		},
		{nil,
			"feature-int",
			"constructor_int",
			[]string{"9000"},
			fnd,
		},
		{nil,
			"feature-float",
			"constructor_float",
			[]string{"9000.0000001"},
			fne,
		},
		{nil,
			"feature-vector",
			"constructor_vector",
			[]string{"TEST", "TEST", "TEST"},
			fnf,
		},
	}

	rawWriteTestFeatures []*tFeature = []*tFeature{
		{[]string{"FILE"},
			"feature-file-strings",
			"constructor_strings",
			[]string{"A", "B", "C", "D", "E"},
			fnb,
		},
	}

	rawSetFeatures []*tFeature = []*tFeature{
		{[]string{"SET"},
			"feature-set-strings",
			"constructor_strings",
			[]string{"a", "b", "c", "4"},
			fnb,
		},
	}

	rawDefines = func(tag string) []*tFeature {
		return []*tFeature{
			{[]string{tag},
				fmt.Sprintf("%s-defines", tag),
				"constructor_strings",
				[]string{"a", "b", "c", "4"},
				func(t *testing.T, f *tFeature, e Env, d *data.Vector) {
					feature := getFeature(t, e, f)
					si, err := feature.EmitStrings()
					if err != nil {
						t.Error(err)
					}

					have := si.ToStrings()
					expect := stringsKeys(3, "test_strings", f.Tag)

					assertEqual(t, "feature-strings", have, expect)
				},
			},
		}
	}

	gc = func(tag string, e Env) Component {
		lc := e.ListComponents()
		for _, v := range lc {
			if v.Tag() == tag {
				return v
			}
		}
		return nil
	}

	ct = func(idx int, name string, fn func(int, string, ...string) []*data.Vector) func(*testing.T, *tComponent, Env) {
		return func(t *testing.T, tc *tComponent, e Env) {
			cs := fn(idx, name, tc.Tag)
			tcs := cs[0]
			assertIn(t, "component", tc.featureKeys(), tcs.Keys())

			var df []string
			for _, v := range tc.Defines {
				df = append(df, v.Tag)
				v.fn(t, v, e, tcs)
			}
			cc := gc(tc.Tag, e)
			if cc == nil {
				t.Error("expected component but got nil")
			}
			assertEqual(t, "component", df, cc.Defines())

			for _, v := range tc.Features {
				v.fn(t, v, e, tcs)
			}
		}
	}

	rawComponents = func(tag string) []*tComponent {
		return []*tComponent{
			{fmt.Sprintf("component-1-%s", tag),
				rawDefines(tag),
				[]*tFeature{
					{[]string{"C-1", tag},
						fmt.Sprintf("c-1-string-%s", tag),
						"constructor_string",
						[]string{"TEST"},
						fna,
					},
					{[]string{"C-1", "testing", tag},
						fmt.Sprintf("c-1-strings-%s", tag),
						"constructor_strings",
						[]string{"a", "b", "c", "4"},
						fnb,
					},
				},
				func(t *testing.T, tc *tComponent, e Env) {
					tfn := ct(1, "TEST", e.GetComponent)
					tfn(t, tc, e)
				},
			},
			{fmt.Sprintf("component-2-%s", tag),
				rawDefines(tag),
				[]*tFeature{
					{[]string{"C-2", tag},
						fmt.Sprintf("c-2-string-%s", tag),
						"constructor_string",
						[]string{"TEST"},
						fna,
					},
					{[]string{"C-2", "testing", tag},
						fmt.Sprintf("c-2-strings-%s", tag),
						"constructor_strings",
						[]string{"a", "b", "c", "4"},
						fnb,
					},
				},
				func(t *testing.T, tc *tComponent, e Env) {
					tfn := ct(1, "TEST", e.MustGetComponent)
					tfn(t, tc, e)
				},
			},
		}
	}

	ge = func(tag string, e Env) Entity {
		le := e.ListEntities()
		for _, v := range le {
			if v.Tag() == tag {
				return v
			}
		}
		return nil
	}

	et = func(idx int, name string, fn func(int, string) []*data.Vector) func(*testing.T, *tEntity, Env) {
		return func(t *testing.T, te *tEntity, e Env) {
			ent := fn(idx, name)
			for _, v := range ent {
				cfk := te.componentFeatureKeys(v.ToString("component.tag"))
				assertIn(t, "entity", cfk, v.Keys())
			}

			var df []string
			for _, v := range te.Defines {
				df = append(df, v.Tag)
				v.fn(t, v, e, nil)
			}
			ce := ge(te.Tag, e)
			if ce == nil {
				t.Error("expected entity but got nil")
			}
			assertEqual(t, "entity", df, ce.Defines())

			for _, v := range te.Components {
				v.fn(t, v, e)
			}
		}
	}

	rawEntities []*tEntity = []*tEntity{
		{"entity-1",
			rawDefines("e1"),
			rawComponents("e1"),
			func(t *testing.T, te *tEntity, e Env) {
				efn := et(1, "entity-1", e.GetEntity)
				efn(t, te, e)
			},
		},
		{"entity-2",
			rawDefines("e2"),
			rawComponents("e2"),
			func(t *testing.T, te *tEntity, e Env) {
				efn := et(2, "entity-2", e.MustGetEntity)
				efn(t, te, e)
			},
		},
	}

	packedFeatureSet string
)

func allFeatures() []*tFeature {
	var ret []*tFeature
	ret = append(ret, rawTestFeatures...)
	ret = append(ret, rawWriteTestFeatures...)
	ret = append(ret, rawSetFeatures...)
	return ret
}

func testable(fs []*tFeature) ([]string, []*tFeature) {
	var reta []string
	var retb []*tFeature
	for _, f := range fs {
		if f.fn != nil {
			reta = append(reta, f.Tag)
			retb = append(retb, f)
		}
	}
	return reta, retb
}

func writeYaml(p string, i interface{}) error {
	f, err := data.Open(p)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(&i)
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
	_, err = f.EmitVector()
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

func testOfFeatureGroup(t *testing.T) {
	e := Empty()

	e.SetConstructor(testConstructors...)

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

	packedFeatureSet = g1v
	g1, g2 = nil, nil
	e = nil
}

func testOfComponent(t *testing.T, e Env) {
	list := e.ListComponents()
	var tags []string
	for _, v := range list {
		tags = append(tags, v.Tag())
	}
	for _, v := range rawComponents("rc") {
		assertIn(t, "component", []string{v.Tag}, tags)
		v.fn(t, v, e)
	}
}

func testOfEntity(t *testing.T, e Env) {
	list := e.ListEntities()
	var tags []string
	for _, v := range list {
		tags = append(tags, v.Tag())
	}
	for _, v := range rawEntities {
		assertIn(t, "entity", []string{v.Tag}, tags)
		v.fn(t, v, e)
	}
}

func testEnv(t *testing.T) Env {
	SetConstructor(additionalConstructor)

	testOfFeatureGroup(t)

	loc1 := loc(1)
	err := writeYaml(loc1, rawWriteTestFeatures)
	if err != nil {
		t.Error(err)
	}
	defer deleteYaml(loc1)

	b, err := yaml.Marshal(&rawTestFeatures)
	if err != nil {
		t.Error(err)
	}

	e, err := New(b)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateYaml(loc1)
	if err != nil {
		t.Error(err)
	}

	err = e.PopulateGroup(packedFeatureSet)
	if err != nil {
		t.Error(err)
	}

	loc2 := loc(2)
	err = writeYaml(loc2, rawComponents("rc"))
	if err != nil {
		t.Error(err)
	}
	defer deleteYaml(loc2)

	err = e.PopulateComponentYaml(loc2)
	if err != nil {
		t.Error(err)
	}

	loc3 := loc(3)
	err = writeYaml(loc3, rawEntities)
	if err != nil {
		t.Error(err)
	}
	defer deleteYaml(loc3)

	err = e.PopulateEntityYaml(loc3)
	if err != nil {
		t.Error(err)
	}

	return e
}

func TestPackage(t *testing.T) {
	e := testEnv(t)

	testOfComponent(t, e)

	testOfEntity(t, e)

	a, f := testable(allFeatures())

	xmfn := func(d *data.Vector) {
		d.SetString("extra", "extra")
	}

	for h := 0; h <= 100; h++ {
		for i := 7; i <= 12; i++ {
			d := NewData(i)

			e.Apply(a, d, xmfn)

			for _, ft := range f {
				ft.fn(t, ft, e, d)
			}
		}
	}
}
