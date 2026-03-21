package service

import (
	"gophermart-loyalty/internal/gopherman/db/conn"
	trm "gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/repository"
)

type Response struct {
	Code int
	Err  error
}
type Service struct {
	Rep       repository.Repositories
	TrManager *trm.Manager
}

func NewService(db *conn.DB, repos repository.Repositories) *Service {
	manager := trm.NewManager(db)
	return &Service{TrManager: manager, Rep: repos}
}
