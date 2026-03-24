package accrual

import (
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/middleware"
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"
)

// MockServer emulates external accrual API for local/testing scenarios.
type MockServer struct {
	Server    *http.Server
	Addr      string
	mu        sync.Mutex
	ordersMap map[string]bool
}

// AvailableTypes maps pseudo-random index to mocked accrual status values.
var AvailableTypes = map[int]string{
	0: Invalid,
	1: Processing,
	2: Processed,
}

// NewMockServer starts an HTTP mock accrual server bound to cfg.AccrualAddress.
func NewMockServer(cfg *server.Config) *MockServer {
	u, err := url.Parse(cfg.AccrualAddress)
	if err != nil {
		panic("accrual mocker: invalid AccrualAddress: " + err.Error())
	}
	addr := u.Host
	r := chi.NewRouter()
	srv := &http.Server{Addr: addr, Handler: r}
	m := &MockServer{Server: srv, Addr: addr, ordersMap: make(map[string]bool)}
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestLimit(rate.NewLimiter(1, 3), 60))
		r.Get("/api/orders/{orderID}", m.handleMockOrder)
	})
	go func() {
		_ = srv.ListenAndServe()
	}()
	return m
}
func (c *MockServer) handleMockOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	orderID := chi.URLParam(r, "orderID")
	if orderID == "" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	resp := Response{
		Order: orderID,
	}
	if _, ok := c.ordersMap[orderID]; !ok {
		c.mu.Lock()
		c.ordersMap[orderID] = true
		resp.Status = Registered
		c.mu.Unlock()
	} else {
		n := rand.Intn(100) + 1
		status := AvailableTypes[n%3]
		resp.Status = status
		if resp.Status == Processed {
			resp.Accrual = new(float64(n))
		}
	}
	w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
	_ = json.NewEncoder(w).Encode(resp)
}
