package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRequestLimit_allowsWithinBurst(t *testing.T) {
	t.Parallel()
	lim := rate.NewLimiter(rate.Inf, 1)
	mw := RequestLimit(lim, 60)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i, rr.Code)
		}
	}
}

func TestRequestLimit_secondRequestRateLimited(t *testing.T) {
	t.Parallel()
	lim := rate.NewLimiter(rate.Every(time.Hour), 1)
	retryAfter := 42
	mw := RequestLimit(lim, retryAfter)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr1 := httptest.NewRecorder()
	mw.ServeHTTP(rr1, req.Clone(req.Context()))
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request: status %d, want 200", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	mw.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: status %d, want 429", rr2.Code)
	}
	if got := rr2.Header().Get("Retry-After"); got != strconv.Itoa(retryAfter) {
		t.Fatalf("Retry-After = %q, want %q", got, strconv.Itoa(retryAfter))
	}
}

func TestRequestLimit_firstRequestBlockedWhenNoTokens(t *testing.T) {
	t.Parallel()
	lim := rate.NewLimiter(0, 0)
	mw := RequestLimit(lim, 10)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("next must not be called")
	}))

	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429", rr.Code)
	}
	if rr.Header().Get("Retry-After") != "10" {
		t.Fatalf("Retry-After = %q", rr.Header().Get("Retry-After"))
	}
}
