package server

import (
	"encoding/json"
	"strings"

	"github.com/Laughs-In-Flowers/data"
)

type Response struct {
	Error error
	Data  *data.Container
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

func StatusResponse(s *Server) *Response {
	d := data.New("")

	d.Set(data.NewStringItem("socket", s.SocketPath))
	d.Set(data.NewStringItem("services", servicesString()))
	d.Set(data.NewStringItem("actions", actionsString()))
	d.Set(data.NewStringItem("features", strings.Join(s.ListKeys(""), ",")))
	d.Set(data.NewStringItem("constructors", strings.Join(s.ListConstructorTags(), ",")))

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
