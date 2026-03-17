package api

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/service"
	"gophermart-loyalty/internal/gopherman/validator"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	ctx := r.Context()
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	res := h.ser.Register(ctx, service.RegisterInput{
		Req: req,
		IP:  ip,
	})
	if res.Err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".Register", res.Err)
		h.Lgr.Info("register error", zap.String("error", lerr.Error()))
		w.WriteHeader(res.Code)
		return
	}
	w.Header().Set("Authorization", "Bearer "+res.Token)
	w.WriteHeader(res.Code)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	ctx := r.Context()
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	res := h.ser.Login(ctx, service.LoginInput{Req: req, IP: ip})
	if res.Err != nil {
		lerr := labelerrors.NewLabelError(constant.LabelApiHandler+".Login", res.Err)
		h.Lgr.Info("login error", zap.String("error", lerr.Error()))
		w.WriteHeader(res.Code)
		return
	}
	w.Header().Set("Authorization", "Bearer "+res.Token)
	w.WriteHeader(res.Code)
}
