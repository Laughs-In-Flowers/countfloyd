package server

import "fmt"

type serverError struct {
	err  string
	vals []interface{}
}

func (s *serverError) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(s.err, s.vals...))
}

func (s *serverError) Out(vals ...interface{}) *serverError {
	s.vals = vals
	return s
}

func Srror(err string) *serverError {
	return &serverError{err: err}
}
