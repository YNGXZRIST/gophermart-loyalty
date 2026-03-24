package middleware

import (
	"gophermart-loyalty/pkg/httpcompressor"
	"net/http"
	"strings"
)

const (
	// ContentEncodingHeader is request content-encoding header key.
	ContentEncodingHeader = "Content-Encoding"
	// AcceptEncodingHeader is request accept-encoding header key.
	AcceptEncodingHeader = "Accept-Encoding"
	// GzipEncoding is gzip content encoding token.
	GzipEncoding = "gzip"
)

// GzipCompressor enables gzip request/response compression when supported.
func GzipCompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get(AcceptEncodingHeader)
		supportsGzip := strings.Contains(acceptEncoding, GzipEncoding)
		if supportsGzip {
			cw := httpcompressor.NewGzipWriter(w)
			ow = cw
			defer cw.Close()
		}
		contentEncoding := r.Header.Get(ContentEncodingHeader)
		sendsGzip := strings.Contains(contentEncoding, GzipEncoding)
		if sendsGzip {
			cr, err := httpcompressor.NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
