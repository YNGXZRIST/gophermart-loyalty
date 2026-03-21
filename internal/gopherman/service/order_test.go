package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestService_GetOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uid := int64(100)
	orders := []*model.Order{{OrderID: "79927398713", Status: "NEW"}}

	tests := []struct {
		name     string
		setup    func(m sqlmock.Sqlmock)
		wantCode int
		wantLen  int
		wantErr  bool
	}{
		{
			name: "repository error",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(repository.OrderGetByUIDQuery).WithArgs(uid).WillReturnError(errors.New("query failed"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success empty",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(repository.OrderGetByUIDQuery).
					WithArgs(uid).
					WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "status", "accrual", "created_at", "updated_at"}))
			},
			wantCode: http.StatusOK,
			wantLen:  0,
		},
		{
			name: "success with orders",
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(repository.OrderGetByUIDQuery).
					WithArgs(uid).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "order_id", "status", "accrual", "created_at", "updated_at"}).
							AddRow(int64(1), orders[0].OrderID, orders[0].Status, 0.0, time.Now(), time.Now()),
					)
			},
			wantCode: http.StatusOK,
			wantLen:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()
			D := &conn.DB{DB: db}
			u := repository.NewUserRepository(D)
			o := repository.NewOrderRepository(D)
			w := repository.NewWithdrawalRepository(D)
			tt.setup(m)
			s := NewService(D, repository.Repositories{User: u, Order: o, Withdrawal: w})
			out := s.GetOrders(ctx, GetOrdersInput{UserID: uid})
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d", out.Code)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("want error")
			}
			if !tt.wantErr && len(out.Orders) != tt.wantLen {
				t.Errorf("len(Orders) = %d, want %d", len(out.Orders), tt.wantLen)
			}
			if err := m.ExpectationsWereMet(); err != nil {
				t.Fatalf("sqlmock expectations not met: %v", err)
			}
		})
	}
}

func TestService_AddOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uid := int64(55)
	validOrder := "79927398713"

	tests := []struct {
		name     string
		in       AddOrderInput
		setup    func(m sqlmock.Sqlmock)
		wantCode int
		wantErr  bool
	}{
		{
			name:     "empty after trim",
			in:       AddOrderInput{UserID: uid, OrderID: "   "},
			wantCode: http.StatusBadRequest,
			wantErr:  true,
		},
		{
			name:     "invalid luhn",
			in:       AddOrderInput{UserID: uid, OrderID: "12345"},
			wantCode: http.StatusUnprocessableEntity,
			wantErr:  true,
		},
		{
			name: "already own order returns 200",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(repository.OrderGetOwnerQuery).
					WithArgs(validOrder).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(uid))
				m.ExpectRollback()
			},
			wantCode: http.StatusOK,
		},
		{
			name: "other user order conflict",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(repository.OrderGetOwnerQuery).
					WithArgs(validOrder).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(uid + 1))
				m.ExpectRollback()
			},
			wantCode: http.StatusConflict,
			wantErr:  true,
		},
		{
			name: "generic add error",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(repository.OrderGetOwnerQuery).
					WithArgs(validOrder).
					WillReturnError(sqlmock.ErrCancelled)
				m.ExpectRollback()
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "accepted new order",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(repository.OrderGetOwnerQuery).
					WithArgs(validOrder).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
				m.ExpectExec(repository.OrderAddOrderQuery).
					WithArgs(uid, validOrder).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			},
			wantCode: http.StatusAccepted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()
			D := &conn.DB{DB: db}
			u := repository.NewUserRepository(D)
			o := repository.NewOrderRepository(D)
			w := repository.NewWithdrawalRepository(D)
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewService(D, repository.Repositories{User: u, Order: o, Withdrawal: w})
			resp := s.AddOrder(ctx, tt.in)
			if resp.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", resp.Code, tt.wantCode)
			}
			if tt.wantErr && resp.Err == nil {
				t.Error("want err")
			}
			if !tt.wantErr && resp.Err != nil {
				t.Errorf("unexpected err: %v", resp.Err)
			}
			if err := m.ExpectationsWereMet(); err != nil {
				t.Fatalf("sqlmock expectations not met: %v", err)
			}
		})
	}
}
