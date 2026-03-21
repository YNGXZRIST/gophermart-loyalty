package service

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/repository"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestService_GetWithdrawals(t *testing.T) {
	ctx := context.Background()
	uid := int64(3)

	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	D := &conn.DB{DB: db}

	m.ExpectQuery(repository.WithdrawalGetByUserIDQuery).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"order_id", "sum", "created_at", "updated_at"}))

	s := NewService(D, repository.Repositories{
		User:       repository.NewUserRepository(D),
		Order:      repository.NewOrderRepository(D),
		Withdrawal: repository.NewWithdrawalRepository(D),
	})
	out := s.GetWithdrawals(ctx, GetWithdrawalsInput{UserID: uid})
	if out.Code != http.StatusOK {
		t.Fatalf("Code = %d, want %d", out.Code, http.StatusOK)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestService_AddWithdrawal_Validation(t *testing.T) {
	s := NewService(nil, repository.Repositories{})
	out := s.AddWithdrawal(context.Background(), WithdrawalInput{UserID: 1, OrderID: "   ", Amount: 10})
	if out.Code != http.StatusBadRequest {
		t.Fatalf("Code = %d, want %d", out.Code, http.StatusBadRequest)
	}
}
