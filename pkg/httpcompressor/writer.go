package httpcompressor

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
)

// Compressor defines writer interface for compressed output.
type Compressor interface {
	io.WriteCloser
	Reset(w io.Writer)
}

// CompressWriter writes compressed HTTP responses.
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

// Header returns response headers of underlying writer.
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses bytes and writes them to response.
func (c *CompressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.w.Header().Set(constant.ContentEncodingHeader, c.encoding)
		c.wroteHeader = true
	}
	return c.compressor.Write(p)
}

// WriteHeader writes HTTP status code and compression headers.
func (c *CompressWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.w.Header().Set(constant.ContentEncodingHeader, c.encoding)
		c.wroteHeader = true
		c.w.WriteHeader(statusCode)
	}
}

// Close closes underlying compressor.
func (c *CompressWriter) Close() error {
	return c.compressor.Close()
}
