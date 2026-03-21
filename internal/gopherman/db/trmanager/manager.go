package trmanager

import (
	"context"
	"database/sql"
	"gophermart-loyalty/internal/gopherman/db/conn"
)

type Manager struct {
	db *conn.DB
}

func NewManager(db *conn.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) WithinTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}
	return tx.Commit()
}
func (m *Manager) WithoutTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := fn(ctx); err != nil {
		return err
	}
	return nil
}
