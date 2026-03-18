package api

import (
	"context"
	"errors"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/repository/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestHandler_UserIDFromRequest(t *testing.T) {
	D, _ := newMockConnDB(t)

	ctrl := gomock.NewController(t)
	mockUserRepo := mock.NewMockUserRepository(ctrl)

	repos := repository.Repositories{User: mockUserRepo}
	handler := NewHandler(D, repos, zap.NewNop())
	t.Parallel()
	t.Run("empty_token_after_trim", func(t *testing.T) {
		got, err := handler.UserIDFromRequest(context.Background(), "Bearer ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Fatalf("userID = %d, want 0", got)
		}
	})

	t.Run("valid_token_calls_repo", func(t *testing.T) {
		mockUserRepo.EXPECT().
			UserIDFromSession(gomock.Any(), "token123").
			Return(int64(42), nil)

		got, err := handler.UserIDFromRequest(context.Background(), "Bearer token123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Fatalf("userID = %d, want 42", got)
		}
	})

	t.Run("repo_error_is_returned", func(t *testing.T) {
		mockUserRepo.EXPECT().
			UserIDFromSession(gomock.Any(), "bad").
			Return(int64(0), errors.New("repo error"))

		_, err := handler.UserIDFromRequest(context.Background(), "Bearer bad")
		if err == nil {
			t.Fatalf("expected error, got nil")
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
