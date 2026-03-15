package api

import (
	"context"
	r "gophermart-loyalty/internal/gopherman/repository"
	"strings"

	"go.uber.org/zap"
)

type Handler struct {
	userRepo  r.UserRepository
	orderRepo r.OrderRepository
	lgr       *zap.Logger
}

func NewHandler(userRepo r.UserRepository, orderRepo r.OrderRepository, lgr *zap.Logger) *Handler {
	return &Handler{
		userRepo:  userRepo,
		orderRepo: orderRepo,
		lgr:       lgr,
	}
}

func (h *Handler) UserIDFromRequest(ctx context.Context, token string) (int64, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return 0, nil
	}
	return h.userRepo.UserIDFromSession(ctx, token)
}
