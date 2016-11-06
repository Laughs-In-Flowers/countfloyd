package constructor

import "fmt"

type crror struct {
	err  string
	vals []interface{}
}

func (c *crror) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(c.err, c.vals...))
}

func (c *crror) Out(vals ...interface{}) *crror {
	c.vals = vals
	return c
}

func Crror(err string) *crror {
	return &crror{err: err}
}
