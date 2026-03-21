package httpretryable

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRetryableClient(t *testing.T) {
	c := NewRetryableClient()
	if c == nil {
		t.Fatal("NewRetryableClient() returned nil")
	}
	if c.client == nil {
		t.Fatal("NewRetryableClient() client must not be nil")
	}
	if c.RetryMax != 0 {
		t.Fatalf("RetryMax = %d, want 0", c.RetryMax)
	}
}

func TestRetryableClient_Do(t *testing.T) {
	t.Run("returns_success_without_retry", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewRetryableClient()
		c.RetryMax = 1
		req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
		resp, err := c.Do(context.Background(), req)
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("retries_on_500_then_success", func(t *testing.T) {
		var count atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			n := count.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("retry"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewRetryableClient()
		c.RetryMax = 3
		req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
		resp, err := c.Do(context.Background(), req)
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}
		defer resp.Body.Close()
		if got := count.Load(); got != 2 {
			t.Fatalf("request attempts = %d, want 2", got)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("retries_on_429_then_success", func(t *testing.T) {
		var count atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			n := count.Add(1)
			if n == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("rate-limited"))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewRetryableClient()
		c.RetryMax = 3
		req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
		resp, err := c.Do(context.Background(), req)
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}
		defer resp.Body.Close()
		if got := count.Load(); got != 2 {
			t.Fatalf("request attempts = %d, want 2", got)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("returns_error_when_retries_exceeded", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		c := NewRetryableClient()
		c.RetryMax = 2
		req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
		resp, err := c.Do(context.Background(), req)
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}

		if err == nil {
			t.Fatal("Do() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "max retries exceeded") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("honors_context_when_waiting_rate_limit", func(t *testing.T) {
		c := NewRetryableClient()
		c.RetryMax = 1
		c.bumpRateBefore(time.Now().Add(200 * time.Millisecond))

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		resp, err := c.Do(ctx, req)
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
		}
		if err == nil {
			t.Fatal("Do() error = nil, want context timeout error")
		}
		if !strings.Contains(err.Error(), "timed out waiting for rate before") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRetryableClient_bumpRateBefore(t *testing.T) {
	t.Run("keeps_maximum_deadline", func(t *testing.T) {
		c := NewRetryableClient()
		first := time.Now().Add(3 * time.Second)
		second := time.Now().Add(1 * time.Second)
		third := time.Now().Add(5 * time.Second)

		c.bumpRateBefore(first)
		c.bumpRateBefore(second)
		got := time.Unix(0, c.rateBeforeUnixNano.Load())
		if got.Before(first) {
			t.Fatalf("rateBefore decreased: got %v, want >= %v", got, first)
		}

		c.bumpRateBefore(third)
		got = time.Unix(0, c.rateBeforeUnixNano.Load())
		if got.Before(third) {
			t.Fatalf("rateBefore not updated: got %v, want >= %v", got, third)
		}
	})

	t.Run("is_safe_under_concurrency", func(t *testing.T) {
		c := NewRetryableClient()
		base := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			i := i
			go func() {
				defer wg.Done()
				c.bumpRateBefore(base.Add(time.Duration(i) * time.Millisecond))
			}()
		}
		wg.Wait()

		got := time.Unix(0, c.rateBeforeUnixNano.Load())
		want := base.Add(49 * time.Millisecond)
		if got.Before(want) {
			t.Fatalf("rateBefore = %v, want >= %v", got, want)
		}
	})
}

func TestRetryableClient_processRetryTimeout(t *testing.T) {
	t.Run("uses_retry_after_seconds", func(t *testing.T) {
		c := NewRetryableClient()
		start := time.Now()

		resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}
		resp.Header.Set("Retry-After", "1")
		if err := c.processRetryTimeout(resp); err != nil {
			t.Fatalf("processRetryTimeout() error = %v", err)
		}

		got := time.Unix(0, c.rateBeforeUnixNano.Load())
		if got.Before(start.Add(900 * time.Millisecond)) {
			t.Fatalf("rateBefore too early: got %v", got)
		}
	})

	t.Run("uses_default_timeout_when_header_missing", func(t *testing.T) {
		c := NewRetryableClient()
		start := time.Now()

		resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}
		if err := c.processRetryTimeout(resp); err != nil {
			t.Fatalf("processRetryTimeout() error = %v", err)
		}

		got := time.Unix(0, c.rateBeforeUnixNano.Load())
		if got.Before(start.Add(defaultTimeout - 200*time.Millisecond)) {
			t.Fatalf("rateBefore too early: got %v", got)
		}
	})

	t.Run("returns_error_for_invalid_retry_after", func(t *testing.T) {
		c := NewRetryableClient()
		resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}
		resp.Header.Set("Retry-After", "invalid")
		if err := c.processRetryTimeout(resp); err == nil {
			t.Fatal("processRetryTimeout() error = nil, want error")
		}
	})
}
