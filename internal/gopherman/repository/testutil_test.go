package repository

import (
	"database/sql"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockConnDB(t *testing.T) (*conn.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return &conn.DB{DB: db}, mock
}

func requireTxDone(t *testing.T, tx *sql.Tx) {
	t.Helper()
	_ = tx.Rollback()
}
