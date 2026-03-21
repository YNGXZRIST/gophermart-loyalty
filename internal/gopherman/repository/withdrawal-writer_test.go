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

func TestWithdrawalWriter_MakeWithdrawal_success(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	userRepo := NewUserRepository(sqlDB)
	wdRepo := NewWithdrawalRepository(sqlDB)
	repos := Repositories{User: userRepo, Withdrawal: wdRepo}
	w := NewWithdrawalWriter(trmanager.NewManager(sqlDB), repos)

	userID := int64(1)
	sum := 25.0
	initialBalance := 100.0
	initialWithdrawn := 20.0
	expectedBalance := initialBalance - sum
	expectedWithdrawn := initialWithdrawn + sum
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)
	wd := &model.Withdrawal{UserID: userID, OrderID: "w1", Sum: sum}

	mock.ExpectBegin()
	mock.ExpectExec(WithdrawalAddQuery).
		WithArgs(userID, wd.OrderID, sum).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
		}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", initialBalance, initialWithdrawn))
	mock.ExpectExec(UserIncrementWithdrawnQuery).
		WithArgs(expectedBalance, expectedWithdrawn, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := w.MakeWithdrawal(ctx, wd); err != nil {
		t.Fatalf("MakeWithdrawal() = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestWithdrawalWriter_MakeWithdrawal_add_fails(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	userRepo := NewUserRepository(sqlDB)
	wdRepo := NewWithdrawalRepository(sqlDB)
	repos := Repositories{User: userRepo, Withdrawal: wdRepo}
	w := NewWithdrawalWriter(trmanager.NewManager(sqlDB), repos)

	wd := &model.Withdrawal{UserID: 1, OrderID: "w1", Sum: 10}
	dbErr := errors.New("insert failed")

	mock.ExpectBegin()
	mock.ExpectExec(WithdrawalAddQuery).
		WithArgs(wd.UserID, wd.OrderID, wd.Sum).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	if err := w.MakeWithdrawal(ctx, wd); err == nil {
		t.Fatal("MakeWithdrawal() error = nil, want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestWithdrawalWriter_MakeWithdrawal_increment_fails(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	userRepo := NewUserRepository(sqlDB)
	wdRepo := NewWithdrawalRepository(sqlDB)
	repos := Repositories{User: userRepo, Withdrawal: wdRepo}
	w := NewWithdrawalWriter(trmanager.NewManager(sqlDB), repos)

	userID := int64(1)
	sum := 15.0
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)
	wd := &model.Withdrawal{UserID: userID, OrderID: "w2", Sum: sum}
	dbErr := errors.New("update user failed")

	mock.ExpectBegin()
	mock.ExpectExec(WithdrawalAddQuery).
		WithArgs(userID, wd.OrderID, sum).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(UserGetByIDQuery).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "login", "pass", "created_at", "updated_at", "last_login_ip", "balance", "withdrawn",
		}).AddRow(userID, "test", "pass", createdAt, updatedAt, "old-ip", 100.0, 10.0))
	mock.ExpectExec(UserIncrementWithdrawnQuery).
		WithArgs(100.0-sum, 10.0+sum, userID).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	if err := w.MakeWithdrawal(ctx, wd); err == nil {
		t.Fatal("MakeWithdrawal() error = nil, want error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
}

func TestNewWithdrawalWriter(t *testing.T) {
	db := &conn.DB{}
	mgr := trmanager.NewManager(db)
	repos := Repositories{}
	w := NewWithdrawalWriter(mgr, repos)
	if w == nil || w.mgr != mgr || w.rep != repos {
		t.Fatalf("NewWithdrawalWriter() = %+v", w)
	}
}
