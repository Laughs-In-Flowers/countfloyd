package constructor

import "fmt"

type cError struct {
	err  string
	vals []interface{}
}

func (c *cError) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(c.err, c.vals...))
}

func (c *cError) Out(vals ...interface{}) *cError {
	c.vals = vals
	return c
}

func Crror(err string) *cError {
	return &cError{err: err}
}
