package server

import (
	"encoding/json"

	"github.com/Laughs-In-Flowers/countfloyd/lib/feature"
	"github.com/Laughs-In-Flowers/data"
)

type Response struct {
	Error string
	Data  *data.Vector
}

// remove giant unmarshaling pita with a string across the boundary
func rErrFmt(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func NewResponse(b []byte) *Response {
	r := &Response{}
	err := json.Unmarshal(b, r)
	if err != nil {
		r.Error = rErrFmt(err)
	}
	return r
}

func (r *Response) ToByte() []byte {
	rb, _ := json.Marshal(r)
	return rb
}

func EmptyResponse() *Response {
	return &Response{"", data.New("")}
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

func ErrorResponse(e error) *Response {
	return &Response{
		rErrFmt(e), nil,
	}
}

var NullResponse []byte
