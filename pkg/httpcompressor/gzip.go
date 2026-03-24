// Package httpcompressor contains reusable HTTP gzip wrappers.
package httpcompressor

import (
	"compress/gzip"
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
)

// NewGzipReader wraps body with gzip reader.
func NewGzipReader(r io.ReadCloser) (*CompressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return newCompressReader(r, gr)
}

// NewGzipWriter wraps response writer with gzip compressor.
func NewGzipWriter(w http.ResponseWriter) *CompressWriter {
	gw := gzip.NewWriter(w)
	return newCompressWriter(w, gw, constant.GzipEncoding)
}
