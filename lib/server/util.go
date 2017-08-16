package server

import (
	"os"

	"github.com/Laughs-In-Flowers/data"
)

func fileData(path string) (*os.File, []byte, error) {
	fl, err := data.Open(path)
	if err != nil {
		return nil, nil, err
	}
	var n int64
	if fi, err := fl.Stat(); err == nil {
		if size := fi.Size(); size < 1e9 {
			n = size
		}
	}
	b := make([]byte, n)
	_, err = fl.Read(b)
	if err != nil {
		return nil, nil, err
	}
	return fl, b, nil
}
