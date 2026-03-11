package api

import (
	"gophermart-loyalty/internal/gopherman/repository"

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
