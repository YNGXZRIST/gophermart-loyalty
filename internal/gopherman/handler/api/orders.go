package api

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/service"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextkey.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	res := h.ser.GetOrders(r.Context(), service.GetOrdersInput{UserID: userID})
	if res.Err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".GetOrders", res.Err)
		h.Lgr.Info("get orders error", zap.String("error", lerr.Error()))
		w.WriteHeader(res.Code)
		return
	}
	data, err := service.OrdersJSON(res.Orders)
	if err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".GetOrders.Marshal", err)
		h.Lgr.Info("marshal error", zap.String("error", lerr.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
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
	res := h.ser.AddOrder(r.Context(), service.AddOrderInput{
		UserID:  userID,
		OrderID: string(body),
	})
	if res.Err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".AddOrder", res.Err)
		h.Lgr.Info("add order error", zap.String("error", lerr.Error()))
	}
	w.WriteHeader(res.Code)
}
