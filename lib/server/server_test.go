package server

import (
	"bytes"
	"testing"

	"github.com/Laughs-In-Flowers/data"
)

var (
	testRequests []*Request = []*Request{
		NewRequest(ByteService("none"), ByteAction("unknown"), nil),
		NewRequest(ByteService("system"), ByteAction("ping"), nil),
		NewRequest(ByteService("data"), ByteAction("apply"), container("TEST")),
	}
	testResponses []*Response = []*Response{
		EmptyResponse(),
	}
)

func container(tag string) *data.Container {
	d := data.New(tag)
	return d
}

func TestRequest(t *testing.T) {
	for _, r := range testRequests {
		rb := r.ToByte()
		sp := NewSpace(rb)
		nr := parse(sp)
		if bytes.Compare(r.Service, nr.Service) != 0 {
			t.Errorf("request and parsed request services are not the same %v %v", r.Service, nr.Service)
		}
		if bytes.Compare(r.Action, nr.Action) != 0 {
			t.Errorf("request and parsed request actions are not the same %v %v", r.Action, nr.Action)
		}
		if bytes.Compare(r.Service, SYSTEM) == 0 {
			if r.Data != nil && nr.Data != nil {
				t.Error("request and parsed request data should be nil, but are not")
			}
		}
		if bytes.Compare(r.Service, DATA) == 0 {
			d, nd := r.Data, nr.Data
			if d.ToString("container.id") != nd.ToString("container.id") {
				t.Error("request data and parsed request data are unequal")
			}
		}
	}
}

func TestResponse(t *testing.T) {
	for _, r := range testResponses {
		rb := r.ToByte()
		nr := NewResponse(rb)
		if r.Error != nr.Error {
			t.Errorf("response and new response error field are not the same %v %v", r.Error, nr.Error)
		}
	}
}

func TestServer(t *testing.T) {}
