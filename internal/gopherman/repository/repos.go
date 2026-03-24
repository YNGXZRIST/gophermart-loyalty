// Package repository provides PostgreSQL-backed data access implementations.
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

// Repositories aggregates concrete repository implementations.
type Repositories struct {
	User       *UserRepo
	Order      *OrderRepo
	Withdrawal *WithdrawalRepo
}
