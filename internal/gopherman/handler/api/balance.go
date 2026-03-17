package api

import (
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
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
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".GetBalance", res.Err)
		h.Lgr.Info("get balance error", zap.String("error", lerr.Error()))
		w.WriteHeader(res.Code)
		return
	}
	w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
	bytes, err := json.Marshal(res.Balance)
	if err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".GetBalance.Marshal", err)
		h.Lgr.Info("marshal balance error", zap.String("error", lerr.Error()))
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
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".MakeWithdraw", res.Err)
		h.Lgr.Info("add withdrawal error", zap.String("error", lerr.Error()))
	}
	w.WriteHeader(res.Code)
}
