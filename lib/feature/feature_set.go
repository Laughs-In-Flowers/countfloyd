package feature

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

func (fs *FeatureSet) Keys() []string {
	var ret []string
	for _, v := range fs.list {
		ret = append(ret, v.Tag)
	}
	return ret
}

func (fs *FeatureSet) List() []RawFeature {
	return fs.list
}

func (fs *FeatureSet) ListString() string {
	b, err := json.Marshal(&fs.list)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (fs *FeatureSet) String() string {
	b, err := fs.Bytes()
	if err != nil {
		return err.Error()
	}
	return string(b)
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

func DecodeFeatureSetValue(s string) (*FeatureSet, error) {
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(d)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	fs := new(bytes.Buffer)
	io.Copy(fs, r)
	var rf []RawFeature
	err = json.Unmarshal(fs.Bytes(), &rf)
	if err != nil {
		return nil, err
	}
	return &FeatureSet{"", "", rf}, nil
}
