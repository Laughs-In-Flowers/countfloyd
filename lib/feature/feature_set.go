package feature

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type FeatureSet struct {
	key   string
	value string
	list  []RawFeature
}

func (fs *FeatureSet) Key() string {
	if fs.key == "" {
		v, _ := fs.Bytes()
		fs.key = fmt.Sprintf("%x", md5.Sum(v))
	}
	return fs.key
}

func (fs *FeatureSet) Bytes() ([]byte, error) {
	return json.Marshal(fs.list)
}

func (fs *FeatureSet) Compress() *bytes.Buffer {
	b := new(bytes.Buffer)
	w := zlib.NewWriter(b)
	by, _ := fs.Bytes()
	w.Write(by)
	w.Close()
	return b
}

func (fs *FeatureSet) Value() string {
	return fs.base64Encode()
}

func (fs *FeatureSet) base64Encode() string {
	if fs.value == "" {
		b := fs.Compress()
		fs.value = base64.StdEncoding.EncodeToString(b.Bytes())
	}
	return fs.value
}
