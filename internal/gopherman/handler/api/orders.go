package api

import (
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/service"
	"io"
	"net/http"
)

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextkey.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	res := service.GetOrders(r.Context(), h.orderRepo, service.GetOrdersInput{UserID: userID})
	if res.Err != nil {
		w.WriteHeader(res.Code)
		return
	}
	data, err := service.OrdersJSON(res.Orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.Code)
	w.Write(data)
}

func (h *Handler) AddOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, ok := contextkey.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	res := service.AddOrder(r.Context(), h.orderRepo, service.AddOrderInput{
		UserID:  userID,
		OrderID: string(body),
	})
	if res.Err != nil {
		w.WriteHeader(res.Code)
		return
	}
	w.WriteHeader(res.Code)
}
