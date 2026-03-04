package middleware

import (
	"gophermart-loyalty/pkg/httpcompressor"
	"net/http"
	"strings"
)

func GzipCompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get(httpcompressor.AcceptEncodingHeader)
		supportsGzip := strings.Contains(acceptEncoding, httpcompressor.GzipEncoding)
		if supportsGzip {
			cw := httpcompressor.NewGzipWriter(w)
			ow = cw
			defer cw.Close()
		}
		contentEncoding := r.Header.Get(httpcompressor.ContentEncodingHeader)
		sendsGzip := strings.Contains(contentEncoding, httpcompressor.GzipEncoding)
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
