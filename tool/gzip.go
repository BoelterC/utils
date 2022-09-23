package tool

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Decompress(data []byte) (string, error) {
	r := bytes.NewReader(data)
	gz, e1 := gzip.NewReader(r)
	if e1 != nil {
		return "", e1
	}
	output, e2 := ioutil.ReadAll(gz)
	if e2 != nil {
		return "", e2
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return string(output), nil
}
