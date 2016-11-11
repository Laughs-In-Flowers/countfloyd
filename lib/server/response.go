package server

import (
	"encoding/json"
	"strings"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

type Response struct {
	Error error
	Data  *data.Vector
}

func NewResponse(b []byte) *Response {
	r := &Response{}
	err := json.Unmarshal(b, r)
	if err != nil {
		r.Error = err
	}
	return r
}

func (r *Response) ToByte() []byte {
	rb, _ := json.Marshal(&r)
	return rb
}

func EmptyResponse() *Response {
	return &Response{nil, data.New("")}
}

func taggedFromConstructor(t ...feature.Constructor) []string {
	var ret []string
	for _, v := range t {
		ret = append(ret, v.Tag())
	}
	return ret
}

func taggedFromRawFeature(t ...feature.RawFeature) []string {
	var ret []string
	for _, v := range t {
		ret = append(ret, v.Tag)
	}
	return ret
}

func StatusResponse(s *Server) *Response {
	d := data.New("")

	d.Set(data.NewStringItem("socket", s.SocketPath))
	d.Set(data.NewStringItem("services", servicesString()))
	d.Set(data.NewStringItem("actions", actionsString()))

	lc := s.ListConstructors()
	cs := taggedFromConstructor(lc...)
	d.Set(data.NewStringItem("constructors", strings.Join(cs, ",")))

	lf := s.List("")
	fs := taggedFromRawFeature(lf...)
	d.Set(data.NewStringItem("features", strings.Join(fs, ",")))

	return &Response{
		nil, d,
	}
}

func ErrorResponse(e error) *Response {
	return &Response{
		e, nil,
	}
}

var NullResponse []byte
