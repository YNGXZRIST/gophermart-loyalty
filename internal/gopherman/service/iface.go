package service

import (
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/repository"
)

type Response struct {
	Code int
	Err  error
}
type Service struct {
	Rep repository.Repositories
	db  *conn.DB
}

func NewService(db *conn.DB, repos repository.Repositories) *Service {
	return &Service{db: db, Rep: repos}
}
