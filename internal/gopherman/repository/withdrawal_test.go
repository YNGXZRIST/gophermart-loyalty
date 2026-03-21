package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewWithdrawalRepository(t *testing.T) {
	db := &conn.DB{}
	got := NewWithdrawalRepository(db)
	if got == nil {
		t.Fatalf("NewWithdrawalRepository returned nil")
	}
	if got.db != db {
		t.Fatalf("NewWithdrawalRepository().db = %v, want %v", got.db, db)
	}
}

func TestWithdrawalRepo_Add(t *testing.T) {
	ctx := context.Background()

	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	m.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	r := &WithdrawalRepo{repoBase: repoBase{db: sqlDB}}

	userID := int64(7)
	orderID := "w1"
	sum := 100.0

	m.ExpectBegin()
	txSQL, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin: %v", err)
	}

	m.ExpectExec(withdrawalAddQuery).
		WithArgs(userID, orderID, sum).
		WillReturnResult(sqlmock.NewResult(1, 1))

	m.ExpectRollback()

	txCtx := trmanager.WithTx(ctx, &conn.Tx{Tx: txSQL})
	if err := r.Add(txCtx, &model.Withdrawal{UserID: userID, OrderID: orderID, Sum: sum}); err != nil {
		t.Fatalf("Add() error = %v, want nil", err)
	}

	_ = txSQL.Rollback()

	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestWithdrawalRepo_GetByUserID(t *testing.T) {
	ctx := context.Background()
	userID := int64(7)

	db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	m.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	r := &WithdrawalRepo{repoBase: repoBase{db: sqlDB}}

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	m.ExpectQuery(withdrawalGetByUserIDQuery).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{"order_id", "sum", "created_at", "updated_at"}).
				AddRow("w1", 100.0, createdAt, updatedAt),
		)

	got, err := r.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(got) != 1 || got[0].OrderID != "w1" || got[0].Sum != 100 {
		t.Errorf("GetByUserID: got %v, want one withdrawal w1 sum=100", got)
	}

	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
