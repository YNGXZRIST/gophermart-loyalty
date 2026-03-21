package conn

import (
	"context"
	"database/sql"
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/db"
	"gophermart-loyalty/internal/gopherman/db/retryable"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	DBLabel  = "DB"
	PGXLabel = DBLabel + ".PGX"
)

type DB struct {
	*sql.DB
	*db.Config
}

func NewConn(cfg *db.Config) (*DB, error) {
	if cfg == nil || cfg.DNS == "" {
		return nil, labelerrors.NewLabelError(PGXLabel+".NewConn.DSN", fmt.Errorf("database DSN is not set"))
	}
	dsn := cfg.DNS
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, labelerrors.NewLabelError(PGXLabel+".NewConn.Open", fmt.Errorf("error connecting to database: %w", err))
	}
	return &DB{DB: conn, Config: cfg}, nil
}
func (D *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return retryable.RunWithRetry(ctx, func() (sql.Result, error) {
		return D.DB.ExecContext(ctx, query, args...)
	})
}

func (D *DB) Exec(query string, args ...any) (sql.Result, error) {
	return retryable.RunWithRetry(context.Background(), func() (sql.Result, error) {
		return D.DB.Exec(query, args...)
	})
}

func (D *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return retryable.RunWithRetry(context.Background(), func() (*sql.Rows, error) {
		return D.DB.Query(query, args...)
	})
}

func (D *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return retryable.RunWithRetry(ctx, func() (*sql.Rows, error) {
		return D.DB.QueryContext(ctx, query, args...)
	})
}
