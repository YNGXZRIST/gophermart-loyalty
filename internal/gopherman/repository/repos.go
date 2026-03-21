package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
)

type repoBase struct{ db *conn.DB }

func (b *repoBase) q(ctx context.Context) trmanager.Querier {
	return trmanager.Resolve(ctx, b.db)
}

type Repositories struct {
	User       UserRepository
	Order      OrderRepository
	Withdrawal WithdrawalRepository
}
