package api

import (
	"context"
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/repository"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestHandler_UserIDFromRequest(t *testing.T) {
	t.Parallel()
	t.Run("empty_token_after_trim", func(t *testing.T) {
		D, _ := newMockConnDB(t)
		handler := NewHandler(D, repository.Repositories{User: repository.NewUserRepository(D)}, zap.NewNop())
		got, err := handler.UserIDFromRequest(context.Background(), "Bearer ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Fatalf("userID = %d, want 0", got)
		}
	})

	t.Run("valid_token_calls_repo", func(t *testing.T) {
		db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		D := &conn.DB{DB: db}
		handler := NewHandler(D, repository.Repositories{User: repository.NewUserRepository(D)}, zap.NewNop())
		m.ExpectQuery(repository.UserUserIDFromSessionQuery).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "ip", "created_at"}).AddRow(int64(42), time.Now().Add(time.Hour), "ip", time.Now()))

		got, err := handler.UserIDFromRequest(context.Background(), "Bearer token123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Fatalf("userID = %d, want 42", got)
		}
		if err := m.ExpectationsWereMet(); err != nil {
			t.Fatalf("sqlmock expectations not met: %v", err)
		}
	})

	t.Run("repo_error_is_returned", func(t *testing.T) {
		db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		D := &conn.DB{DB: db}
		handler := NewHandler(D, repository.Repositories{User: repository.NewUserRepository(D)}, zap.NewNop())
		m.ExpectQuery(repository.UserUserIDFromSessionQuery).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(errors.New("repo error"))

		_, err = handler.UserIDFromRequest(context.Background(), "Bearer bad")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("sql_no_rows_is_returned", func(t *testing.T) {
		db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()
		D := &conn.DB{DB: db}
		handler := NewHandler(D, repository.Repositories{User: repository.NewUserRepository(D)}, zap.NewNop())
		m.ExpectQuery(repository.UserUserIDFromSessionQuery).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)
		_, err = handler.UserIDFromRequest(context.Background(), "Bearer bad")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestNewHandler(t *testing.T) {
	type args struct {
		conn  *conn.DB
		repos repository.Repositories
		lgr   *zap.Logger
	}

	D, _ := newMockConnDB(t)
	repos := repository.Repositories{}
	lgr := zap.NewNop()

	h := NewHandler(D, repos, lgr)
	if h == nil {
		t.Fatalf("NewHandler() = nil")
	}
	if h.ser == nil {
		t.Fatalf("NewHandler().ser = nil")
	}
	if h.Lgr != lgr {
		t.Fatalf("NewHandler().Lgr != provided logger")
	}
}
