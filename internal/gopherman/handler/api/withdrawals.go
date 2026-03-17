package api

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/service"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextkey.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	res := h.ser.GetWithdrawals(r.Context(), service.GetWithdrawalsInput{UserID: userID})
	if res.Err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".GetWithdrawals", res.Err)
		h.Lgr.Info("get orders error", zap.String("error", lerr.Error()))
		w.WriteHeader(res.Code)
		return
	}
	w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
	w.WriteHeader(http.StatusOK)
	bytes, err := service.WithdrawalsJSON(res.Withdrawals)
	if err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".GetWithdrawals.Marshal", err)
		h.Lgr.Info("get orders error", zap.String("error", lerr.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
