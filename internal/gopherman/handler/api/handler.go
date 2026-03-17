package api

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	r "gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/service"
	"strings"

	"go.uber.org/zap"
)

type Handler struct {
	ser *service.Service
	Lgr *zap.Logger
}

func NewHandler(conn *conn.DB, repos r.Repositories, lgr *zap.Logger) *Handler {
	newService := service.NewService(conn, repos)
	return &Handler{
		ser: newService,
		Lgr: lgr,
	}
}

func (h *Handler) UserIDFromRequest(ctx context.Context, token string) (int64, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return 0, nil
	}
	return h.ser.Rep.User.UserIDFromSession(ctx, token)
}
