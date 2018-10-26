package env_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Laughs-In-Flowers/countfloyd/lib/env"
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/xrr"
	yaml "gopkg.in/yaml.v2"
)

func stringKey(pre, tag string) string {
	PRE, TAG := strings.ToUpper(pre), strings.ToUpper(tag)
	return fmt.Sprintf("%s:%s", PRE, TAG)
}

func tConstructorString(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	ckey := stringKey("test_string", tag)

	ef := func() data.Item {
		return data.NewStringItem(tag, ckey)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return feature.NewInformer("CONSTRUCTOR_STRING", r.Group, tag, r.Values, []string{ckey}),
		feature.NewEmitter(ef),
		feature.NewMapper(mf)
}

func stringsKeys(l int, pre, tag string) []string {
	var ret []string
	key := stringKey(pre, tag)
	for i := 1; i <= l; i++ {
		ret = append(ret, key)
	}
	return ret
}

func tConstructorStrings(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	ckeys := stringsKeys(3, "test_strings", tag)

	ef := func() data.Item {
		return data.NewStringsItem(tag, ckeys...)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return feature.NewInformer("CONSTRUCTOR_STRINGS", r.Group, tag, r.Values, ckeys),
		feature.NewEmitter(ef),
		feature.NewMapper(mf)
}

func tConstructorBool(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	ef := func() data.Item {
		return data.NewBoolItem(tag, false)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return feature.NewInformer("CONSTRUCTOR_BOOL", r.Group, tag, r.Values, []string{"false"}),
		feature.NewEmitter(ef),
		feature.NewMapper(mf)
}

func tConstructorInt(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	values := r.MustGetValues()
	v := values[0]
	vn, _ := strconv.Atoi(v)

	ef := func() data.Item {
		return data.NewIntItem(tag, vn)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return feature.NewInformer("CONSTRUCTOR_INT", r.Group, tag, r.Values, []string{"9000"}),
		feature.NewEmitter(ef),
		feature.NewMapper(mf)
}

func tConstructorFloat(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	values := r.MustGetValues()
	v := values[0]
	vn, _ := strconv.ParseFloat(v, 64)

	ef := func() data.Item {
		return data.NewFloat64Item(tag, vn)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return feature.NewInformer("CONSTRUCTOR_FLOAT", r.Group, tag, r.Values, []string{"9000.0000001"}),
		feature.NewEmitter(ef),
		feature.NewMapper(mf)
}

func tConstructorVector(tag string, r *feature.RawFeature, e feature.CEnv) (feature.Informer, feature.Emitter, feature.Mapper) {
	ckey := stringKey("test_vector", tag)

	ef := func() data.Item {
		d := data.New("")
		d.Set(data.NewStringItem("key", ckey))
		return data.NewVectorItem("vector", d)
	}

	mf := func(d *data.Vector) {
		d.Set(ef())
	}

	return feature.NewInformer("CONSTRUCTOR_VECTOR", r.Group, tag, r.Values, []string{ckey}),
		feature.NewEmitter(ef),
		feature.NewMapper(mf)
}

type tFeature struct {
	Group  []string                                                `"yaml:set"`
	Tag    string                                                  `"yaml:tag"`
	Apply  string                                                  `"yaml:apply"`
	Values []string                                                `"yaml:values"`
	fn     func(*testing.T, *tFeature, feature.CEnv, *data.Vector) `"yaml:-"`
}

type tComponent struct {
	Tag      string                                      `"yaml:tag"`
	Defines  []*tFeature                                 `"yaml:defines"`
	Features []*tFeature                                 `"yaml:features"`
	fn       func(*testing.T, *tComponent, feature.CEnv) `"yaml:-"`
}

func (tc *tComponent) featureKeys() []string {
	var ret []string
	for _, v := range tc.Features {
		ret = append(ret, strings.ToUpper(v.Tag))
	}
	return ret
}

type tEntity struct {
	Tag        string                                   `"yaml:tag"`
	Defines    []*tFeature                              `"yaml:defines"`
	Components []*tComponent                            `"yaml:features"`
	fn         func(*testing.T, *tEntity, feature.CEnv) `"yaml:-"`
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

func getFeature(t *testing.T, e feature.CEnv, f *tFeature) feature.Feature {
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
	additionalConstructor feature.Constructor = feature.NewConstructor("CONSTRUCTOR_STRINGS", 100, tConstructorStrings)

	testConstructors []feature.Constructor = []feature.Constructor{
		feature.NewConstructor("CONSTRUCTOR_STRING", 47, tConstructorString),
		feature.NewConstructor("CONSTRUCTOR_BOOL", 700, tConstructorBool),
		feature.NewConstructor("CONSTRUCTOR_INT", 500, tConstructorInt),
		feature.NewConstructor("CONSTRUCTOR_FLOAT", 2000, tConstructorFloat),
		feature.DefaultConstructor("CONSTRUCTOR_VECTOR", tConstructorVector),
	}

	fna = func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
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

	fnb = func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
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

	fnc = func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
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

	fnd = func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
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

	fne = func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
		feature := getFeature(t, e, f)
		fi, err := feature.EmitFloat()
		if err != nil {
			t.Error(err)
		}
		have := fi.ToFloat64()
		expect := d.ToFloat64(strings.ToUpper(f.Tag))
		if have != expect {
			t.Errorf("feature-float: have %f, expect %f", have, expect)
		}
	}

	fnf = func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
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

	//in progress
	rawPluginConstructors []*tFeature = []*tFeature{
		{[]string{"PLUGIN"}, // feature from constructer plugin
			"plugin-constructor",
			"test_constructor_one",
			[]string{},
			nil, //tbd
		},
	}

	//in progress
	rawPluginFeatures []*tFeature = []*tFeature{
		{[]string{"PLUGIN"}, // feature of feature plugin
			"plugin-feature",
			"",
			[]string{""},
			nil, //tbd
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
				func(t *testing.T, f *tFeature, e feature.CEnv, d *data.Vector) {
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

	gc = func(tag string, e feature.CEnv) feature.Component {
		lc := e.ListComponents()
		for _, v := range lc {
			if v.Tag() == tag {
				return v
			}
		}
		return nil
	}

	ct = func(idx float64, name string, fn func(float64, string, ...string) []*data.Vector) func(*testing.T, *tComponent, feature.CEnv) {
		return func(t *testing.T, tc *tComponent, e feature.CEnv) {
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
				func(t *testing.T, tc *tComponent, e feature.CEnv) {
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
				func(t *testing.T, tc *tComponent, e feature.CEnv) {
					tfn := ct(1, "TEST", e.MustGetComponent)
					tfn(t, tc, e)
				},
			},
		}
	}

	ge = func(tag string, e feature.CEnv) feature.Entity {
		le := e.ListEntities()
		for _, v := range le {
			if v.Tag() == tag {
				return v
			}
		}
		return nil
	}

	et = func(idx float64, name string, fn func(float64, string) []*data.Vector) func(*testing.T, *tEntity, feature.CEnv) {
		return func(t *testing.T, te *tEntity, e feature.CEnv) {
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
			func(t *testing.T, te *tEntity, e feature.CEnv) {
				efn := et(1, "entity-1", e.GetEntity)
				efn(t, te, e)
			},
		},
		{"entity-2",
			rawDefines("e2"),
			rawComponents("e2"),
			func(t *testing.T, te *tEntity, e feature.CEnv) {
				efn := et(2, "entity-2", e.MustGetEntity)
				efn(t, te, e)
			},
		},
	}

	packedFeatureSet string

	rootDir string = "/tmp/cf"

	plgnDir string = filepath.Join(rootDir, "plgn")

	plgnFile string = "plgn.go"

	plgnCLoc string = filepath.Join(plgnDir, plgnFile)

	plgnPlgn string = "plgn.so"

	plgnLoc string = filepath.Join(plgnDir, plgnPlgn)

	plgn string = `package main
		import (
			"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
			"github.com/Laughs-In-Flowers/data"
		)

		type TC1 struct {}

		func (t TC1) Tag() string {
			return "TEST_CONSTRUCTOR_ONE"
		}

		func (t TC1) Order() int {
			return 666
		}

		func (t TC1) Construct(tag string, rf *feature.RawFeature, e feature.CEnv) feature.Feature {
			return NewTF1("tf1_from_CONSTRUCTOR")
		}

		func Constructors() []feature.Constructor {
			return []feature.Constructor{TC1{}}
		}

		type TF1 struct {
			tag string
			feature.Informer
			feature.Emitter
			feature.Mapper
		}

		func ef(v string) func() data.Item {
			return func() data.Item {
				return data.NewStringItem("plugin-key", v)
			}
		}

		func NewTF1(tag string) *TF1 {
			return &TF1{
				tag: tag,
				Informer:feature.NewInformer(
					tag,
					[]string{},
					tag,
					[]string{tag, "plugin", "feature"},
					[]string{tag, "plugin", "feature"}),
				Emitter: feature.NewEmitter(ef(tag)),
				Mapper:feature.NewMapper(func(d *data.Vector){ d.Set(ef(tag)()) }),
			}
		}

		func Features() []feature.Feature {
			return []feature.Feature{
				NewTF1("tf1-instance-FEATURE"),
			}
		}
		`
)

func allFeatures() []*tFeature {
	var ret []*tFeature
	ret = append(ret, rawTestFeatures...)
	ret = append(ret, rawPluginConstructors...)
	ret = append(ret, rawPluginFeatures...)
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

func loc(i int) string {
	name := fmt.Sprintf("features-%d.yaml", i)
	return filepath.Join(rootDir, name)
}

func exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

var openError = xrr.Xrror("unable to find or open file %s, provided %s").Out

func open(path string) (*os.File, error) {
	p := filepath.Clean(path)
	dir, name := filepath.Split(p)
	var fp string
	var err error
	switch dir {
	case "":
		fp, err = filepath.Abs(name)
	default:
		exist(dir)
		fp, err = filepath.Abs(p)
	}

	if err != nil {
		return nil, err
	}

	if file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0660); err == nil {
		return file, nil
	}

	return nil, openError(fp, path)
}

func writeYaml(p string, i interface{}) error {
	f, err := open(p)
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

func writePlugin(p string) error {
	f, err := open(p)
	if err != nil {
		return err
	}
	f.WriteString(plgn)
	f.Close()
	return nil
}

func compilePlugin(p string) error {
	cmd := exec.Command(
		"go",
		"build",
		"-buildmode=plugin",
		"-o",
		plgnLoc,
		p,
	)
	o, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("%s: %s", err.Error(), string(o)))
	}
	return nil
}

func deleteFile(p string) {
	os.Remove(p)
}

type constructorSort []feature.Constructor

func (c constructorSort) Len() int {
	return len(c)
}

func (c constructorSort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c constructorSort) Less(i, j int) bool {
	return c[i].Order() > c[j].Order()
}

func testOfConstructor(t *testing.T, e feature.CEnv) {
	ck := "CONSTRUCTOR_STRINGS"
	_, exists := feature.GetConstructor(ck)
	if !exists {
		t.Errorf("Constructor does not exist: %s", ck)
	}

	cl1 := feature.ListConstructors()
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

func testOfFeature(t *testing.T, e feature.CEnv) {
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
			t.Errorf("feature value is unexpected: %s - %v", v, expect)
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
			t.Errorf("Emit type provided error is nil: %v", check2)
		}
	}

	if err := e.SetFeature(&feature.RawFeature{Tag: "feature-set-strings"}); err == nil {
		t.Error("Setting duplicate named feature did not return an error.")
	}
}

func testOfFeatureGroup(t *testing.T) {
	e := env.Empty()

	e.SetConstructor(testConstructors...)

	b, err := yaml.Marshal(&rawSetFeatures)
	if err != nil {
		t.Error(err)
	}
	e.Populate(b)

	testOfFeature(t, e)

	g1 := e.GetGroup("SET")
	g1v := g1.Value()
	g2, err := feature.DecodeFeatureGroup(g1v)
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

func testOfComponent(t *testing.T, e feature.CEnv) {
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

func testOfEntity(t *testing.T, e feature.CEnv) {
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

func errIf(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func testEnv(t *testing.T) feature.CEnv {
	var err error
	feature.SetConstructor(additionalConstructor)

	testOfFeatureGroup(t)

	loc1 := loc(1)
	errIf(t, writeYaml(loc1, rawWriteTestFeatures))
	defer deleteFile(loc1)

	var b []byte
	b, err = yaml.Marshal(&rawTestFeatures)
	errIf(t, err)

	var e env.Env
	e, err = env.New()
	errIf(t, err)

	errIf(t, e.Populate(b))

	errIf(t, writePlugin(plgnCLoc))

	errIf(t, compilePlugin(plgnCLoc))

	errIf(t, e.PopulateConstructorPlugin(plgnDir))

	var cb []byte
	cb, err = yaml.Marshal(&rawPluginConstructors)
	errIf(t, e.Populate(cb))

	errIf(t, e.PopulateFeaturePlugin([]string{}, plgnDir))

	defer deleteFile(plgnCLoc)
	defer deleteFile(plgnLoc)

	errIf(t, e.PopulateFeatureYaml([]string{}, loc1))

	errIf(t, e.PopulateFeatureGroupString([]string{}, packedFeatureSet))

	loc2 := loc(2)
	errIf(t, writeYaml(loc2, rawComponents("rc")))
	defer deleteFile(loc2)

	errIf(t, e.PopulateComponentYaml([]string{}, loc2))

	loc3 := loc(3)
	errIf(t, writeYaml(loc3, rawEntities))
	defer deleteFile(loc3)

	errIf(t, e.PopulateEntityYaml([]string{}, loc3))

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
			d := feature.NewData(float64(i))

			e.Apply(a, d, xmfn)

			for _, ft := range f {
				ft.fn(t, ft, e, d)
			}
		}
	}

	os.RemoveAll(rootDir)
}
