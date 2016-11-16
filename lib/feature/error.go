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

var (
	ExistsError       = Frror("A %s named %s already exists.").Out
	DoesNotExistError = Frror("A %s named %s does not exist.").Out
	NotFoundError     = Frror("%s named %s not found, exiting").Out
)
