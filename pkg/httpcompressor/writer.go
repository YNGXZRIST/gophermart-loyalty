package httpcompressor

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
)

type Compressor interface {
	io.WriteCloser
	Reset(w io.Writer)
}
type CompressWriter struct {
	w           http.ResponseWriter
	compressor  Compressor
	encoding    string
	wroteHeader bool
}

func newCompressWriter(w http.ResponseWriter, compressor Compressor, encoding string) *CompressWriter {
	compressor.Reset(w)
	return &CompressWriter{
		w:           w,
		compressor:  compressor,
		encoding:    encoding,
		wroteHeader: false,
	}
}
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.w.Header().Set(constant.ContentEncodingHeader, c.encoding)
		c.wroteHeader = true
	}
	return c.compressor.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.w.Header().Set(constant.ContentEncodingHeader, c.encoding)
		c.wroteHeader = true
		c.w.WriteHeader(statusCode)
	}
}
func (c *CompressWriter) Close() error {
	return c.compressor.Close()
}
