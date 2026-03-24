// Package service contains application business logic.
package service

import (
	"context"

	"gophermart-loyalty/internal/gopherman/db/conn"
	trm "gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
)

// Response is a generic service-layer operation result.
type Response struct {
	Code int
	Err  error
}

// Service aggregates business use-cases and repository dependencies.
type Service struct {
	Rep              repository.Repositories
	accrualWriter    *repository.AccrualWriter
	withdrawalWriter *repository.WithdrawalWriter
}

// NewService constructs a Service and initializes transactional writers.
func NewService(db *conn.DB, repos repository.Repositories) *Service {
	manager := trm.NewManager(db)
	return &Service{
		Rep:              repos,
		accrualWriter:    repository.NewAccrualWriter(manager, repos),
		withdrawalWriter: repository.NewWithdrawalWriter(manager, repos),
	}
}

// ApplyAccrualResult applies accrual status and balance changes atomically.
func (s *Service) ApplyAccrualResult(ctx context.Context, order *model.Order) error {
	return s.accrualWriter.ApplyResult(ctx, order)
}
