package api

import (
	"context"
	"gophermart-loyalty/internal/gopherman/repository"
	"strings"

	"go.uber.org/zap"
)

type Handler struct {
	userRepo repository.UserRepository
	lgr      *zap.Logger
}

func NewHandler(userRepo repository.UserRepository, lgr *zap.Logger) *Handler {
	return &Handler{
		userRepo: userRepo,
		lgr:      lgr,
	}
}

func (h *Handler) ValidateSession(ctx context.Context, token string) (bool, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return false, nil
	}
	return h.userRepo.IsValidSession(ctx, token)
}
