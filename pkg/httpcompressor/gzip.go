package httpcompressor

import (
	"compress/gzip"
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
)

func NewGzipReader(r io.ReadCloser) (*CompressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return newCompressReader(r, gr)
}

func NewGzipWriter(w http.ResponseWriter) *CompressWriter {
	gw := gzip.NewWriter(w)
	return newCompressWriter(w, gw, constant.GzipEncoding)
}
