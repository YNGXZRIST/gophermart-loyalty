package api

import (
	"fmt"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/validator"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	ctx := r.Context()
	err := validator.DecodeAndValidate(r, &req)
	isValid := req.Validate()
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if isValid != nil {
		h.lgr.Info("not valid request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.userRepo.GetByLogin(ctx, req.Login)
	if err != nil {
		h.lgr.Info("get user by login failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}
	user, err = h.userRepo.Register(ctx, req.Login, req.Pass, ip)
	if err != nil {
		h.lgr.Info("register user failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	session, err := h.userRepo.CreateSession(ctx, user.ID, ip)
	if err != nil {
		h.lgr.Info("create session failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+session)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	ctx := r.Context()
	err := validator.DecodeAndValidate(r, &req)
	isValid := req.Validate()
	if isValid != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	u, err := h.userRepo.GetByLogin(ctx, req.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if u == nil {
		fmt.Println("user not found")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err = password.Compare(u.Pass, req.Pass); err != nil {
		fmt.Println("password compare failed")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	session, err := h.userRepo.CreateSession(ctx, u.ID, ip)
	if err != nil {
		h.lgr.Info("create session failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("session:", session)
	w.Header().Set("Authorization", "Bearer "+session)
	w.WriteHeader(http.StatusOK)

}
