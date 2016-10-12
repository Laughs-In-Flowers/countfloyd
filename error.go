package main

import "fmt"

type cfcError struct {
	err  string
	vals []interface{}
}

func (c *cfcError) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(c.err, c.vals...))
}

func (c *cfcError) Out(vals ...interface{}) *cfcError {
	c.vals = vals
	return c
}

func Crror(err string) *cfcError {
	return &cfcError{err: err}
}
