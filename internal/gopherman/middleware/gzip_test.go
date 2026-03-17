package middleware

import (
	"bytes"
	"compress/gzip"
	"gophermart-loyalty/internal/gopherman/constant"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipCompressor(t *testing.T) {
	tests := []struct {
		name               string
		acceptEncoding     string
		contentEncoding    string
		reqBody            []byte
		wantStatus         int
		wantNextCalled     bool
		wantRespEncoding   string
		wantRespBodyDecode bool
	}{
		{
			name:               "no encoding headers passes through",
			acceptEncoding:     "",
			contentEncoding:    "",
			reqBody:            nil,
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantRespEncoding:   "",
			wantRespBodyDecode: false,
		},
		{
			name:               "Accept-Encoding gzip compresses response",
			acceptEncoding:     constant.GzipEncoding,
			contentEncoding:    "",
			reqBody:            nil,
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantRespEncoding:   constant.GzipEncoding,
			wantRespBodyDecode: true,
		},
		{
			name:               "Content-Encoding gzip decompresses request body",
			acceptEncoding:     "",
			contentEncoding:    constant.GzipEncoding,
			reqBody:            gzipEncode([]byte("hello")),
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantRespEncoding:   "",
			wantRespBodyDecode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			var nextBody []byte
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if r.Body != nil {
					nextBody, _ = io.ReadAll(r.Body)
				}
				w.WriteHeader(tt.wantStatus)
				_, _ = w.Write([]byte("ok"))
			})

			handler := GzipCompressor(next)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.reqBody))
			if tt.acceptEncoding != "" {
				req.Header.Set(constant.AcceptEncodingHeader, tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set(constant.ContentEncodingHeader, tt.contentEncoding)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if !nextCalled {
				t.Error("next handler was not called")
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantRespEncoding != "" {
				if got := rec.Header().Get(constant.ContentEncodingHeader); got != tt.wantRespEncoding {
					t.Errorf("Content-Encoding = %q, want %q", got, tt.wantRespEncoding)
				}
			}
			if tt.wantRespBodyDecode {
				gr, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("gzip.NewReader: %v", err)
				}
				defer gr.Close()
				decoded, err := io.ReadAll(gr)
				if err != nil {
					t.Fatalf("ReadAll decoded: %v", err)
				}
				if string(decoded) != "ok" {
					t.Errorf("decoded body = %q, want %q", decoded, "ok")
				}
			}
			if tt.contentEncoding == constant.GzipEncoding && len(nextBody) > 0 && string(nextBody) != "hello" {
				t.Errorf("next received body = %q, want %q", nextBody, "hello")
			}
		})
	}
}

func TestGzipMiddleware_invalidGzipBodyReturns500(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called")
	})

	handler := GzipCompressor(next)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not gzip")))
	req.Header.Set(constant.ContentEncodingHeader, constant.GzipEncoding)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func gzipEncode(p []byte) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write(p)
	_ = gw.Close()
	return buf.Bytes()
}
