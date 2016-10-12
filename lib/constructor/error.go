package constructor

import "fmt"

type featureError struct {
	err  string
	vals []interface{}
}

func (m *featureError) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(m.err, m.vals...))
}

func (m *featureError) Out(vals ...interface{}) *featureError {
	m.vals = vals
	return m
}

func Crror(err string) *featureError {
	return &featureError{err: err}
}
