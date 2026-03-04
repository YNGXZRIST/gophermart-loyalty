package httpcompressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

const (
	ContentEncodingHeader = "Content-Encoding"
	AcceptEncodingHeader  = "Accept-Encoding"
	GzipEncoding          = "gzip"
	ContentTypeHeader     = "Content-Type"
	ApplicationJSON       = "application/json"
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
	return newCompressWriter(w, gw, GzipEncoding)
}
