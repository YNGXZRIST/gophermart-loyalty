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

func TestNewOrderRepository(t *testing.T) {
	db := &conn.DB{}
	got := NewOrderRepository(db)
	if got == nil {
		t.Fatalf("NewUserRepository returned nil")
	}
	if _, ok := got.(*orderRepo); !ok {
		t.Fatalf("NewUserRepository() type = %T, want *userRepo", got)
	}
}

func Test_orderRepo_Add(t *testing.T) {
	ctx := context.Background()
	userID := int64(5)
	orderID := "order-1"

	db, mock := newMockConnDB(t)
	r := &orderRepo{repoBase: repoBase{db: db}}
	mock.MatchExpectationsInOrder(false)

	mock.ExpectBegin()

	mock.ExpectQuery(OrderGetOwnerQuery).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
	mock.ExpectExec(OrderAddOrderQuery).
		WithArgs(userID, orderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	if err := r.Add(ctx, userID, orderID); err != nil {
		t.Fatalf("Add() error = %v, want nil", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_orderRepo_GetByUserID(t *testing.T) {
	ctx := context.Background()
	userID := int64(10)
	db, mock := newMockConnDB(t)
	r := &orderRepo{repoBase: repoBase{db: db}}
	mock.MatchExpectationsInOrder(false)

	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now().Add(-time.Minute)
	accrual := 1.23

	mock.ExpectQuery(OrderGetByUidQuery).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "order_id", "status", "accrual", "created_at", "updated_at"}).
				AddRow(int64(1), "123", "NEW", accrual, createdAt, updatedAt),
		)

	got, err := r.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}
	if len(got) != 1 || got[0].OrderID != "123" || got[0].UserID != userID || got[0].Status != "NEW" || got[0].Accrual == nil || *got[0].Accrual != accrual {
		t.Fatalf("GetByUserID() got %+v, want order 123 user %d accrual %v", got[0], userID, accrual)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_orderRepo_GetOrdersPendingAccrual(t *testing.T) {
	ctx := context.Background()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	sqlDB := &conn.DB{DB: db}
	r := &orderRepo{repoBase: repoBase{db: sqlDB}}

	createdAt := time.Now().Add(-time.Minute)
	updatedAt := time.Now().Add(-time.Second)

	rows1 := sqlmock.NewRows([]string{"id", "order_id", "user_id", "status", "created_at", "updated_at"}).
		AddRow(int64(1), "order-1", int64(10), "NEW", createdAt, updatedAt)
	rows2 := sqlmock.NewRows([]string{"id", "order_id", "user_id", "status", "created_at", "updated_at"})

	mock.ExpectQuery(OrderGetPendingOrdersWithLimitQuery).
		WithArgs(int64(0), 1).
		WillReturnRows(rows1)
	mock.ExpectQuery(OrderGetPendingOrdersWithLimitQuery).
		WithArgs(int64(1), 1).
		WillReturnRows(rows2)

	got, err := r.GetOrdersPendingAccrual(ctx)
	if err != nil {
		t.Fatalf("GetOrdersPendingAccrual() error = %v, want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("GetOrdersPendingAccrual() len = %d, want 1", len(got))
	}
	if got[0].ID != 1 || got[0].OrderID != "order-1" || got[0].UserID != 10 || got[0].Status != "NEW" {
		t.Fatalf("GetOrdersPendingAccrual() got = %+v, want order-1 NEW user 10", got[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func Test_orderRepo_UpdateOrderAccrual(t *testing.T) {
	ctx := context.Background()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	sqlDB := &conn.DB{DB: db}
	r := &orderRepo{repoBase: repoBase{db: sqlDB}}

	var accrual = 12.34
	order := &model.Order{ID: 1, Status: "PROCESSING", Accrual: &accrual}

	mock.ExpectBegin()
	txSQL, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin: %v", err)
	}

	mock.ExpectExec(OrderUpdatePendingOrderQuery).
		WithArgs(order.Status, accrual, order.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	txCtx := trmanager.WithTx(ctx, &conn.Tx{Tx: txSQL})
	if err := r.UpdateOrderAccrual(txCtx, order); err != nil {
		t.Fatalf("UpdateOrderAccrual() error = %v, want nil", err)
	}

	_ = txSQL.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
