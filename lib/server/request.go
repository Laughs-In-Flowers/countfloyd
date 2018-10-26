package server

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/Laughs-In-Flowers/data"
)

type Request struct {
	Service Service
	Action  Action
	Data    *data.Vector
}

func NewRequest(s Service, a Action, d *data.Vector) *Request {
	return &Request{
		s, a, d,
	}
}

func request(req []byte) *Request {
	s := NewSpace(req)
	r := parse(s)
	return r
}

func parse(s Space) *Request {
	return &Request{
		s.Service(), s.Action(), s.Data(),
	}
}

var Sep []byte = []byte("++")

func (r *Request) ToByte() []byte {
	l := make([][]byte, 0)
	l = append(l, r.Service)
	l = append(l, r.Action)
	b, err := json.Marshal(r.Data)
	if err != nil {
		l = append(l, []byte(err.Error()))
	}
	l = append(l, b)
	return bytes.Join(l, Sep)
}

type Space [][]byte

func NewSpace(in []byte) Space {
	ret := make(Space, 0)
	fs := bytes.Split(in, Sep)
	if len(fs) == 3 {
		for _, v := range fs {
			ret = append(ret, v)
		}
		return ret
	}
	return ret
}

func (s Space) Service() Service {
	return s[0]
}

func (s Space) Action() Action {
	return s[1]
}

func (s Space) Data() *data.Vector {
	d := data.New("")
	err := json.Unmarshal(s[2], &d)
	if err != nil {
		d.Set(data.NewStringItem("Error", err.Error()))
	}
	return d
}

type Service []byte

func ByteService(s string) Service {
	b := []byte(s)
	if isService(b) {
		return b
	}
	return NONE
}

func isService(b []byte) bool {
	for _, v := range services {
		if bytes.Compare(b, v) == 0 {
			return true
		}
	}
	return false
}

func (s Service) String() string {
	return string(s)
}

var (
	NONE   = []byte("none")
	SYSTEM = []byte("system")
	QUERY  = []byte("query")
	DATA   = []byte("data")

	services []Service = []Service{
		SYSTEM, QUERY, DATA,
	}
)

func servicesString() string {
	var l []string
	for _, v := range services {
		l = append(l, v.String())
	}
	return strings.Join(l, ",")
}

type Action []byte

func ByteAction(s string) Action {
	b := []byte(s)
	if isAction(b) {
		return b
	}
	return UNKNOWN
}

func isAction(b []byte) bool {
	for _, v := range actions {
		if bytes.Compare(b, v) == 0 {
			return true
		}
	}
	return false
}

func actionIs(a, b Action) bool {
	if bytes.Compare(a, b) == 0 {
		return true
	}
	return false
}

func (a Action) String() string {
	return string(a)
}

var (
	UNKNOWN           = []byte("unknown")
	PING              = []byte("ping")
	QUIT              = []byte("quit")
	STATUS            = []byte("status")
	QUERYFEATURE      = []byte("query_feature")
	QUERYCOMPONENT    = []byte("query_component")
	QUERYENTITY       = []byte("query_entity")
	POPULATEFROMFILES = []byte("populate_from_files")
	DEPOPULATE        = []byte("depopulate")
	APPLYFEATURE      = []byte("apply_feature")
	APPLYCOMPONENT    = []byte("apply_component")
	APPLYENTITY       = []byte("apply_entity")

	actions []Action = []Action{
		PING,
		QUIT,
		STATUS,
		QUERYFEATURE,
		QUERYCOMPONENT,
		QUERYENTITY,
		POPULATEFROMFILES,
		DEPOPULATE,
		APPLYFEATURE,
		APPLYCOMPONENT,
		APPLYENTITY,
	}
)

func actionsString() string {
	var l []string
	for _, v := range actions {
		l = append(l, v.String())
	}
	return strings.Join(l, ",")
}
