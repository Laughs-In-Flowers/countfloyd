package feature

import "fmt"

type featureError struct {
	err  string
	vals []interface{}
}

func (f *featureError) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(f.err, f.vals...))
}

func (f *featureError) Out(vals ...interface{}) *featureError {
	f.vals = vals
	return f
}

func Frror(err string) *featureError {
	return &featureError{err: err}
}
