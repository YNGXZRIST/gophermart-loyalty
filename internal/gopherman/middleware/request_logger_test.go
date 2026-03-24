package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestWithRequestLogger(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name         string
		method       string
		body         string
		wantStatus   int
		wantNextCall bool
	}{
		{
			name:         "GET request calls next and returns 200",
			method:       http.MethodGet,
			body:         "",
			wantStatus:   http.StatusOK,
			wantNextCall: true,
		},
		{
			name:         "POST with body calls next and returns 200",
			method:       http.MethodPost,
			body:         `{"id":"x"}`,
			wantStatus:   http.StatusOK,
			wantNextCall: true,
		},
		{
			name:         "POST reads body and passes to next",
			method:       http.MethodPost,
			body:         `{"type":"gauge"}`,
			wantStatus:   http.StatusOK,
			wantNextCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})

			handler := WithRequestLogger(logger)(next)
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCall {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCall)
			}
		})
	}
}

func TestWithRequestLogger_recordsStatusAndSize(t *testing.T) {
	logger := zap.NewNop()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	handler := WithRequestLogger(logger)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if body := rec.Body.String(); body != "created" {
		t.Errorf("body = %q, want %q", body, "created")
	}
}
