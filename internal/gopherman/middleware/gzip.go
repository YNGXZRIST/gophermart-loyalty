package middleware

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/pkg/httpcompressor"
	"net/http"
	"strings"
)

func GzipCompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get(constant.AcceptEncodingHeader)
		supportsGzip := strings.Contains(acceptEncoding, constant.GzipEncoding)
		if supportsGzip {
			cw := httpcompressor.NewGzipWriter(w)
			ow = cw
			defer cw.Close()
		}
		contentEncoding := r.Header.Get(constant.ContentEncodingHeader)
		sendsGzip := strings.Contains(contentEncoding, constant.GzipEncoding)
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
