package repository

import (
	"context"
	"errors"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAccrualWriter_ApplyResult_success_with_balance(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	orderRepo := NewOrderRepository(sqlDB)
	userRepo := NewUserRepository(sqlDB)
	repos := Repositories{Order: orderRepo, User: userRepo}
	w := NewAccrualWriter(trmanager.NewManager(sqlDB), repos)

	accrual := 12.5
	order := &model.Order{ID: 1, UserID: 7, Status: "PROCESSING", Accrual: &accrual}
	userID := int64(7)
	initialBalance := 100.0
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)

	mock.ExpectBegin()
	mock.ExpectExec(OrderUpdatePendingOrderQuery).
		WithArgs(order.Status, accrual, order.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
		}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", initialBalance, 20.0))
	mock.ExpectExec(UserIncrementBalanceQuery).
		WithArgs(initialBalance+accrual, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := w.ApplyResult(ctx, order); err != nil {
		t.Fatalf("ApplyResult() = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestAccrualWriter_ApplyResult_success_nil_accrual_skips_balance(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	orderRepo := NewOrderRepository(sqlDB)
	userRepo := NewUserRepository(sqlDB)
	repos := Repositories{Order: orderRepo, User: userRepo}
	w := NewAccrualWriter(trmanager.NewManager(sqlDB), repos)

	order := &model.Order{ID: 2, UserID: 3, Status: "REGISTERED", Accrual: nil}

	mock.ExpectBegin()
	mock.ExpectExec(OrderUpdatePendingOrderQuery).
		WithArgs(order.Status, nil, order.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := w.ApplyResult(ctx, order); err != nil {
		t.Fatalf("ApplyResult() = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestAccrualWriter_ApplyResult_update_fails_rollbacks(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	orderRepo := NewOrderRepository(sqlDB)
	userRepo := NewUserRepository(sqlDB)
	repos := Repositories{Order: orderRepo, User: userRepo}
	w := NewAccrualWriter(trmanager.NewManager(sqlDB), repos)

	accrual := 1.0
	order := &model.Order{ID: 1, UserID: 7, Status: "PROCESSING", Accrual: &accrual}
	dbErr := errors.New("update failed")

	mock.ExpectBegin()
	mock.ExpectExec(OrderUpdatePendingOrderQuery).
		WithArgs(order.Status, accrual, order.ID).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	if err := w.ApplyResult(ctx, order); err == nil {
		t.Fatal("ApplyResult() error = nil, want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestAccrualWriter_ApplyResult_balance_fails_rollbacks(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	orderRepo := NewOrderRepository(sqlDB)
	userRepo := NewUserRepository(sqlDB)
	repos := Repositories{Order: orderRepo, User: userRepo}
	w := NewAccrualWriter(trmanager.NewManager(sqlDB), repos)

	accrual := 5.0
	order := &model.Order{ID: 1, UserID: 7, Status: "PROCESSING", Accrual: &accrual}
	userID := int64(7)
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)
	dbErr := errors.New("balance exec failed")

	mock.ExpectBegin()
	mock.ExpectExec(OrderUpdatePendingOrderQuery).
		WithArgs(order.Status, accrual, order.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
		}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", 100.0, 0.0))
	mock.ExpectExec(UserIncrementBalanceQuery).
		WithArgs(100.0+accrual, userID).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	if err := w.ApplyResult(ctx, order); err == nil {
		t.Fatal("ApplyResult() error = nil, want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestNewAccrualWriter(t *testing.T) {
	db := &conn.DB{}
	mgr := trmanager.NewManager(db)
	repos := Repositories{}
	w := NewAccrualWriter(mgr, repos)
	if w == nil || w.mgr != mgr || w.rep != repos {
		t.Fatalf("NewAccrualWriter() = %+v", w)
	}
}
