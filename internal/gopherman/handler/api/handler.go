package api

import (
	"gophermart-loyalty/internal/gopherman/db/conn"
)

type Handler struct {
	conn *conn.DB
}

func NewHandler(conn *conn.DB) *Handler {
	return &Handler{
		conn: conn,
	}
}
