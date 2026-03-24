package middleware

import (
	"bytes"
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
		Body   string
	}
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// Write records response size/body and forwards bytes to the underlying writer.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.Body = string(b)
	return size, err
}

// WriteHeader records status code and forwards it to the underlying writer.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithRequestLogger logs request and response metadata.
func WithRequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sugar := logger.Sugar()
			if r.Method == http.MethodPost {
				var buf bytes.Buffer
				tee := io.TeeReader(r.Body, &buf)
				body, err := io.ReadAll(tee)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				sugar.Infoln(
					"uri", r.RequestURI,
					"method", r.Method,
					"request body", string(body),
				)
				r.Body = io.NopCloser(&buf)
			}
			start := time.Now()
			responseData := &responseData{
				status: 0,
				size:   0,
				Body:   "",
			}
			acceptEncoding := r.Header.Get(constant.AcceptEncodingHeader)
			supportsGzip := strings.Contains(acceptEncoding, constant.GzipEncoding)
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)
			duration := time.Since(start)
			sugar.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
				"supportsGzip", supportsGzip,
				"body", responseData.Body,
			)

		})
	}
}
