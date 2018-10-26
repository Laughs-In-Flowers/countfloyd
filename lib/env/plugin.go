package env

import (
	"os"
	"path/filepath"

	p "plugin"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/xrr"
)

// An interface for plugin loading.
type Loader interface {
	AddDirs(...string) error
	ListPlugin() (map[string][]string, error)
	LoadConstructor() ([]feature.Constructor, error)
	LoadFeature() ([]feature.Feature, error)
}

type loaders struct {
	has []Loader
}

// Provides a new, multiple directory handling Loader.
func NewPlugins(dirs ...string) (*loaders, error) {
	def := make([]Loader, 0)
	ret := &loaders{def}
	err := ret.AddDirs(dirs...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Adds a new directory.
func (l *loaders) AddDirs(dirs ...string) error {
	for _, dir := range dirs {
		nl, err := newLoader(dir)
		if err != nil {
			return err
		}
		l.has = append(l.has, nl)
	}
	return nil
}

//
func (l *loaders) ListPlugin() (map[string][]string, error) {
	ret := make(map[string][]string)
	for _, sl := range l.has {
		ps, _ := sl.ListPlugin()
		for k, v := range ps {
			ret[k] = v
		}
	}
	return ret, nil
}

// Loads, returning an array of Constructor and any error.
func (l *loaders) LoadConstructor() ([]feature.Constructor, error) {
	var ret []feature.Constructor
	for _, ld := range l.has {
		nc, err := ld.LoadConstructor()
		if err != nil {
			return nil, err
		}
		ret = append(ret, nc...)
	}
	return ret, nil
}

// Loads, returning an array of Feature and any error.
func (l *loaders) LoadFeature() ([]feature.Feature, error) {
	var ret []feature.Feature
	for _, ld := range l.has {
		nc, err := ld.LoadFeature()
		if err != nil {
			return nil, err
		}
		ret = append(ret, nc...)
	}
	return ret, nil
}

type loader struct {
	dir  string
	lfn  func(*loader) (map[string][]string, error)
	llfn func(*loader, bool, bool) ([]feature.Constructor, []feature.Feature, error)
}

func loaderDir(d string) string {
	pth, _ := filepath.Abs(d)
	return pth
}

func newLoader(dir string) (*loader, error) {
	return &loader{
		loaderDir(dir),
		defaultLister,
		defaultLoader,
	}, nil
}

// Does not add a directory, single directory is specified at instantiation.
func (l *loader) AddDirs(...string) error { return nil }

func defaultLister(l *loader) (map[string][]string, error) {
	dir, err := os.Open(l.dir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	ret := make(map[string][]string)
	var res []string
	for _, name := range names {
		if filepath.Ext(name) == ".so" {
			res = append(res, name)
		}
	}
	ret[l.dir] = res

	return ret, nil
}

// Satisfies the interface Loader.Plugins function for this *loader
func (l *loader) ListPlugin() (map[string][]string, error) {
	return l.lfn(l)
}

func defaultLoader(l *loader, c, f bool) ([]feature.Constructor, []feature.Feature, error) {
	plugins, err := l.ListPlugin()
	if err != nil {
		return nil, nil, nil // pass through here and do nothing, its less mess if a dir doesnt exist
	}
	var rcs []feature.Constructor
	var rfs []feature.Feature
	for _, v := range plugins {
		var srcPath string
		for _, plugin := range v {
			srcPath = filepath.Join(l.dir, plugin)
			cs, fs, err := loadPath(l, srcPath, c, f)
			if err != nil {
				return nil, nil, err
			}
			rcs = append(rcs, cs...)
			rfs = append(rfs, fs...)
		}
	}
	return rcs, rfs, nil
}

var (
	//
	OpenPluginError = xrr.Xrror("Unable to open plugin at %s:\n\t for %s").Out
	//
	DoesntExistError = xrr.Xrror("Plugin at %s has no %s.").Out
)

func loadPath(l *loader, path string, c, f bool) ([]feature.Constructor, []feature.Feature, error) {
	pl, err := p.Open(path)
	if err != nil {
		return nil, nil, OpenPluginError(path, err)
	}

	if c {
		luc, err := pl.Lookup("Constructors")
		if err != nil {
			return nil, nil, DoesntExistError(path, "Constructors (func() []feature.Constructor)")
		}
		var lcfn func() []feature.Constructor
		var ok bool
		if lcfn, ok = luc.(func() []feature.Constructor); !ok {
			return nil, nil, OpenPluginError(path, "error with plugin Constructors (func() []feature.Constructor)")
		}
		return lcfn(), nil, nil
	}

	if f {
		luf, err := pl.Lookup("Features")
		if err != nil {
			return nil, nil, DoesntExistError(path, "Features (func() []feature.Feature)")
		}
		var llfn func() []feature.Feature
		var ok bool
		if llfn, ok = luf.(func() []feature.Feature); !ok {
			return nil, nil, OpenPluginError(path, "error with plugin Features (func() []feature.Feature)")
		}
		return nil, llfn(), nil
	}

	return nil, nil, nil
}

// Satisfies the interface Loader.LoadConstructor function for this *loader
func (l *loader) LoadConstructor() ([]feature.Constructor, error) {
	c, _, err := l.llfn(l, true, false)
	return c, err
}

//
func (l *loader) LoadFeature() ([]feature.Feature, error) {
	_, f, err := l.llfn(l, false, true)
	return f, err
}
