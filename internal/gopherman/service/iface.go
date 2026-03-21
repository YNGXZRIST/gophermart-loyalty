package service

import (
	"context"

	"gophermart-loyalty/internal/gopherman/db/conn"
	trm "gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
)

type Response struct {
	Code int
	Err  error
}
type Service struct {
	Rep              repository.Repositories
	accrualWriter    *repository.AccrualWriter
	withdrawalWriter *repository.WithdrawalWriter
}

func NewService(db *conn.DB, repos repository.Repositories) *Service {
	manager := trm.NewManager(db)
	return &Service{
		Rep:              repos,
		accrualWriter:    repository.NewAccrualWriter(manager, repos),
		withdrawalWriter: repository.NewWithdrawalWriter(manager, repos),
	}
}

func (s *Service) ApplyAccrualResult(ctx context.Context, order *model.Order) error {
	return s.accrualWriter.ApplyResult(ctx, order)
}
