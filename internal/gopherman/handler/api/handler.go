package api

import (
	"gophermart-loyalty/internal/gopherman/db/conn"

	"go.uber.org/zap"
)

type Handler struct {
	conn *conn.DB
	lgr  *zap.Logger
}

func NewHandler(conn *conn.DB, lgr *zap.Logger) *Handler {
	return &Handler{
		conn: conn,
		lgr:  lgr,
	}
}
