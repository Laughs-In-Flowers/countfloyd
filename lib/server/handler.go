package server

import (
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/env"
	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
	"github.com/Laughs-In-Flowers/xrr"
)

type Handler struct {
	Service, Action string
	Fn              HandlerFunc
}

func NewHandler(service, action string, fn HandlerFunc) *Handler {
	return &Handler{
		service, action, fn,
	}
}

type HandlerFunc func(*Server, *Request) []byte

type hActions map[string]HandlerFunc

type hServices map[string]hActions

type Handlers struct {
	has hServices
}

func NewHandlers(hn ...*Handler) *Handlers {
	h := &Handlers{
		has: make(hServices),
	}
	for _, v := range services {
		h.has[v.String()] = make(hActions)
	}
	h.SetHandle(hn...)
	return h
}

func (h *Handlers) GetRequestedHandle(r *Request) (HandlerFunc, error) {
	var sv, ac string
	sv = r.Service.String()
	ac = r.Action.String()
	return h.GetHandle(sv, ac)
}

func (h *Handlers) GetHandle(service, action string) (HandlerFunc, error) {
	if sm, ok := h.has[service]; ok {
		if fn, ok := sm[action]; ok {
			return fn, nil
		}
		return nil, UnknownActionError(action)
	}
	return nil, NoServiceError(service)
}

func (h *Handlers) SetHandle(hs ...*Handler) {
	for _, hn := range hs {
		if isService(ByteService(hn.Service)) {
			if service, ok := h.has[hn.Service]; ok {
				if !isAction([]byte(hn.Action)) {
					actions = append(actions, []byte(hn.Action))
				}
				service[hn.Action] = hn.Fn
			}
		}
	}
}

func pingRespond(*Server, *Request) []byte {
	return NullResponse
}

func quitRespond(s *Server, r *Request) []byte {
	s.Quit()
	return nil
}

func applyRespond(a Action) HandlerFunc {
	return func(s *Server, r *Request) []byte {
		resp := EmptyResponse()
		d := r.Data
		resp.Data = applyDataFrom(a, d, s)
		return resp.ToByte()
	}
}

func applyDataFrom(a Action, m *data.Vector, e env.Env) *data.Vector {
	n := m.ToFloat64("meta.priority")
	switch {
	case actionIs(a, APPLYFEATURE):
		f := m.ToStrings("meta.feature")
		e.Apply(f, m)
	case actionIs(a, APPLYCOMPONENT):
		id := m.ToString("meta.id")
		cs := m.ToStrings("meta.component")
		css := e.GetComponent(n, id, cs...)
		for _, v := range css {
			m.SetVector(v.ToString("component.id"), v)
		}
	case actionIs(a, APPLYENTITY):
		en := m.ToString("meta.entity")
		ent := e.GetEntity(n, en)
		for _, v := range ent {
			m.SetVector(v.ToString("component.id"), v)
		}
	}

	return m
}

func populateRespond(s *Server, r *Request) []byte {
	resp := EmptyResponse()
	d := r.Data
	groups := d.ToStrings("groups")
	var noneOf bool = true
	if cc := d.ToStrings("constructor-plugin"); cc != nil && len(cc) > 0 {
		resp.Error = rErrFmt(s.PopulateConstructorPlugin(cc...))
	}
	if cf := d.ToStrings("feature-plugin"); cf != nil && len(cf) > 0 {
		resp.Error = rErrFmt(s.PopulateFeaturePlugin(groups, cf...))
	}
	if fs := d.ToStrings("features"); fs != nil && len(fs) > 0 {
		resp.Error = rErrFmt(s.PopulateFeatureYaml(groups, fs...))
		noneOf = false
	}
	if cs := d.ToStrings("components"); cs != nil && len(cs) > 0 {
		resp.Error = rErrFmt(s.PopulateComponentYaml(groups, cs...))
		noneOf = false
	}
	if es := d.ToStrings("entities"); es != nil && len(es) > 0 {
		resp.Error = rErrFmt(s.PopulateEntityYaml(groups, es...))
		noneOf = false
	}
	if noneOf {
		resp.Error = "nothing to populate"
	}
	resp.Data = d
	return resp.ToByte()
}

func depopulateRespond(s *Server, r *Request) []byte {
	resp := EmptyResponse()
	d := r.Data
	groups := d.ToStrings("groups")
	err := s.Remove(groups...)
	if err != nil {
		resp.Error = err.Error()
	}
	resp.Data = d
	return resp.ToByte()
}

func queryRespond(a Action) HandlerFunc {
	return func(s *Server, r *Request) []byte {
		resp := EmptyResponse()
		d := r.Data
		resp.Data = queryDataFrom(a, s, d)
		return resp.ToByte()
	}
}

func queryDataFrom(a Action, s *Server, d *data.Vector) *data.Vector {
	switch {
	case actionIs(a, STATUS):
		return statusData(s, d)
	case actionIs(a, QUERYFEATURE):
		return featureData(s, d)
	case actionIs(a, QUERYCOMPONENT):
		return componentData(s, d)
	case actionIs(a, QUERYENTITY):
		return entityData(s, d)
	}

	return d
}

func statusData(s *Server, d *data.Vector) *data.Vector {
	d.Set(data.NewStringItem("socket", s.SocketPath))
	d.Set(data.NewStringItem("services", servicesString()))
	d.Set(data.NewStringItem("actions", actionsString()))

	lc := s.ListConstructors()
	cs := taggedFromConstructor(lc...)
	d.Set(data.NewStringItem("constructors", strings.Join(cs, ",")))

	lf := s.List("")
	fs := taggedFromRawFeature(lf...)
	d.Set(data.NewStringItem("features", strings.Join(fs, ",")))

	cm := s.ListComponents()
	var cml []string
	for _, v := range cm {
		cml = append(cml, v.Tag())
	}
	d.Set(data.NewStringItem("components", strings.Join(cml, ",")))

	el := s.ListEntities()
	var etl []string
	for _, v := range el {
		etl = append(etl, v.Tag())
	}
	d.Set(data.NewStringItem("entities", strings.Join(etl, ",")))

	return d
}

func featureData(e env.Env, d *data.Vector) *data.Vector {
	q := d.ToString("query_feature")
	f := e.GetFeature(q)
	if f != nil {
		si := data.NewStringsItem("groups", f.Group()...)
		fi := data.NewStringItem("apply", f.From())
		vi := data.NewStringsItem("values", f.Values()...)
		d.Set(si, fi, vi)
	}
	return d
}

func componentData(e env.Env, d *data.Vector) *data.Vector {
	q := d.ToString("query_component")
	l := e.ListComponents()
	var c feature.Component
	for _, v := range l {
		if q == v.Tag() {
			c = v
		}
	}
	if c != nil {
		di := data.NewStringsItem("component.defines", c.Defines()...)
		fi := data.NewStringsItem("component.has_features", c.Features()...)
		d.Set(di, fi)
	}
	return d
}

func entityData(e env.Env, d *data.Vector) *data.Vector {
	q := d.ToString("query_entity")
	l := e.ListEntities()
	var ee feature.Entity
	for _, v := range l {
		if q == v.Tag() {
			ee = v
		}
	}
	if e != nil {
		di := data.NewStringsItem("entity.defines", ee.Defines()...)
		ci := data.NewStringsItem("entity.has_components", ee.Components()...)
		d.Set(di, ci)
	}
	return d
}

var localHandlers []*Handler = []*Handler{
	NewHandler(
		"system",
		"ping",
		pingRespond,
	),
	NewHandler(
		"system",
		"quit",
		quitRespond,
	),
	NewHandler(
		"query",
		"status",
		queryRespond(STATUS),
	),
	NewHandler(
		"query",
		"query_feature",
		queryRespond(QUERYFEATURE),
	),
	NewHandler(
		"query",
		"query_component",
		queryRespond(QUERYCOMPONENT),
	),
	NewHandler(
		"query",
		"query_entity",
		queryRespond(QUERYENTITY),
	),
	NewHandler(
		"data",
		"populate_from_files",
		populateRespond,
	),
	NewHandler(
		"data",
		"depopulate",
		depopulateRespond,
	),
	NewHandler(
		"data",
		"apply_feature",
		applyRespond(APPLYFEATURE),
	),
	NewHandler(
		"data",
		"apply_component",
		applyRespond(APPLYCOMPONENT),
	),
	NewHandler(
		"data",
		"apply_entity",
		applyRespond(APPLYENTITY),
	),
}

var (
	NoServiceError     = xrr.Xrror("no service named %s available").Out
	UnknownActionError = xrr.Xrror("unknown action %s").Out
)
