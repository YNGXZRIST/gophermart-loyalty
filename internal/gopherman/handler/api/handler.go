// Package api contains HTTP handlers for gophermart endpoints.
package api

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	r "gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/service"
	"strings"

	"go.uber.org/zap"
)

// Handler groups HTTP handlers and service dependencies.
type Handler struct {
	ser *service.Service
	Lgr *zap.Logger
}

// NewHandler creates API handler set with service dependencies.
func NewHandler(conn *conn.DB, repos r.Repositories, lgr *zap.Logger) *Handler {
	newService := service.NewService(conn, repos)
	return &Handler{
		ser: newService,
		Lgr: lgr,
	}
}

// UserIDFromRequest extracts user id from bearer token.
func (h *Handler) UserIDFromRequest(ctx context.Context, token string) (int64, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return 0, nil
	}
	return h.ser.Rep.User.UserIDFromSession(ctx, token)
}
