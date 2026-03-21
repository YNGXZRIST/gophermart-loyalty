package trmanager

import (
	"context"
	"database/sql"
	"gophermart-loyalty/internal/gopherman/db/conn"
)

type Querier interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func Resolve(ctx context.Context, db *conn.DB) Querier {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return db
}
