package conn

import (
	"context"
	"database/sql"
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/db"
	"gophermart-loyalty/internal/gopherman/db/pgerrors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const maxRetries = 3

type Tx struct {
	*sql.Tx
}

type DB struct {
	*sql.DB
	*db.Config
}

func NewConn(cfg *db.Config) (*DB, error) {
	if cfg == nil || cfg.DNS == "" {
		return nil, fmt.Errorf("database DSN is not set")
	}
	dsn := cfg.DNS
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	return &DB{DB: conn, Config: cfg}, nil
}

func runWithRetry[T any](ctx context.Context, op func() (T, error)) (T, error) {
	var zero T
	var lastRes T
	var lastErr error
	sleepSeconds := 1
	classifier := pgerrors.NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxRetries; attempt++ {
		lastRes, lastErr = op()
		if lastErr == nil {
			return lastRes, nil
		}
		if classifier.Classify(lastErr) == pgerrors.NonRetriable {
			return zero, pgerrors.NewPgError(lastErr)
		}
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(time.Duration(sleepSeconds) * time.Second):
			sleepSeconds += 2
		}
	}
	if lastErr != nil {
		lastErr = pgerrors.NewPgError(lastErr)
	}
	return lastRes, lastErr
}

func (D *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return runWithRetry(ctx, func() (sql.Result, error) {
		return D.DB.ExecContext(ctx, query, args...)
	})
}

func (D *DB) Exec(query string, args ...any) (sql.Result, error) {
	return runWithRetry(context.Background(), func() (sql.Result, error) {
		return D.DB.Exec(query, args...)
	})
}

func (D *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(context.Background(), func() (*sql.Rows, error) {
		return D.DB.Query(query, args...)
	})
}

func (D *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(ctx, func() (*sql.Rows, error) {
		return D.DB.QueryContext(ctx, query, args...)
	})
}
func (D *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	sqlTx, err := D.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: sqlTx}, nil
}
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return runWithRetry(ctx, func() (sql.Result, error) {
		return tx.Tx.ExecContext(ctx, query, args...)
	})
}
func (tx *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return runWithRetry(context.Background(), func() (sql.Result, error) {
		return tx.Tx.Exec(query, args...)
	})
}

func (tx *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(context.Background(), func() (*sql.Rows, error) {
		return tx.Tx.Query(query, args...)
	})
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return runWithRetry(ctx, func() (*sql.Rows, error) {
		return tx.Tx.QueryContext(ctx, query, args...)
	})
}
