package middleware

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentTypeJSON(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		contentType    string
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "application/json passes and sets response header",
			contentType:    constant.ApplicationJSON,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "empty Content-Type returns 415",
			contentType:    "",
			wantStatus:     http.StatusUnsupportedMediaType,
			wantNextCalled: false,
		},
		{
			name:           "text/plain returns 415",
			contentType:    "text/plain",
			wantStatus:     http.StatusUnsupportedMediaType,
			wantNextCalled: false,
		},
		{
			name:           "application/xml returns 415",
			contentType:    "application/xml",
			wantStatus:     http.StatusUnsupportedMediaType,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.contentType != "" {
				req.Header.Set(constant.ContentTypeHeader, tt.contentType)
			}
			rec := httptest.NewRecorder()

			handler := ContentTypeJSON(next)
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantNextCalled && rec.Header().Get(constant.ContentTypeHeader) != constant.ApplicationJSON {
				t.Errorf("Content-Type = %q, want %q", rec.Header().Get(constant.ContentTypeHeader), constant.ApplicationJSON)
			}
		})
	}
}
