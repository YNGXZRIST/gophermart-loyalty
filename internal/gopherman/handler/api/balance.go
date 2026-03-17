package api

import (
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/service"
	"net/http"

	"go.uber.org/zap"
)

type WithdrawalRequest struct {
	OrderID string  `json:"order"`
	Amount  float64 `json:"sum"`
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextkey.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	res := h.ser.GetBalance(r.Context(), service.BalanceInput{UserID: userID})
	if res.Err != nil {
		h.lgr.Info("get balance error", zap.String("error", res.Err.Error()))
		w.WriteHeader(res.Code)
		return
	}
	w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
	bytes, err := json.Marshal(res.Balance)
	if err != nil {
		h.lgr.Info("marshal balance error", zap.String("error", res.Err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(res.Code)
	w.Write(bytes)
}
func (h *Handler) MakeWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextkey.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var req WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	res := h.ser.AddWithdrawal(r.Context(), service.WithdrawalInput{
		UserID:  userID,
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if res.Err != nil {
		h.lgr.Info("add withdrawal error", zap.String("error", res.Err.Error()))
	}
	w.WriteHeader(res.Code)
}
