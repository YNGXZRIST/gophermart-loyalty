package accrual

import (
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/constant"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestMockServer_handleMockOrder(t *testing.T) {
	c := &MockServer{
		Server:    &http.Server{Addr: ":8080"},
		Addr:      ":8080",
		mu:        sync.Mutex{},
		ordersMap: make(map[string]bool),
	}

	router := chi.NewRouter()
	router.Get("/api/orders/{orderID}", c.handleMockOrder)

	t.Run("first", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/orders/123", nil)
		router.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Fatalf("status code = %d, want %d", got, want)
		}
		if ct := rr.Header().Get(constant.ContentTypeHeader); ct != constant.ApplicationJSON {
			t.Fatalf("content-type = %q, want %q", ct, constant.ApplicationJSON)
		}

		var resp Response
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Order != "123" {
			t.Fatalf("resp.Order = %q, want %q", resp.Order, "123")
		}
		if resp.Status != constant.Registered {
			t.Fatalf("resp.Status = %q, want %q", resp.Status, constant.Registered)
		}
		if resp.Accrual != nil {
			t.Fatalf("resp.Accrual must be nil for registered status")
		}

		c.mu.Lock()
		_, ok := c.ordersMap["123"]
		c.mu.Unlock()
		if !ok {
			t.Fatalf("ordersMap must contain orderID=123 after first request")
		}
	})

	t.Run("second", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/orders/123", nil)
		router.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusOK; got != want {
			t.Fatalf("second status code = %d, want %d", got, want)
		}

		var resp Response
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal second response: %v", err)
		}

		if resp.Status == constant.Registered {
			t.Fatalf("second resp.Status must not be %q", constant.Registered)
		}
		if resp.Status == constant.Processed && resp.Accrual == nil {
			t.Fatalf("processed status must have non-nil accrual")
		}

		c.mu.Lock()
		_, ok := c.ordersMap["123"]
		c.mu.Unlock()
		if !ok {
			t.Fatalf("ordersMap must still contain orderID=123 after second request")
		}
	})

	t.Run("post req", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/orders/123", nil)
		c.handleMockOrder(rr, req)
		if got, want := rr.Code, http.StatusMethodNotAllowed; got != want {
			t.Fatalf("non-get status code = %d, want %d", got, want)
		}
	})

	t.Run("empty order", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c.handleMockOrder(rr, req)
		if got, want := rr.Code, http.StatusNotFound; got != want {
			t.Fatalf("missing orderID status code = %d, want %d", got, want)
		}
	})
}
