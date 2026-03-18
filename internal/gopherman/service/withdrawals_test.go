package service

import (
	"context"
	"errors"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/repository/mock"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
)

func testMockConn(t *testing.T, setup func(sqlmock.Sqlmock)) *conn.DB {
	t.Helper()
	sqlDB, m, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	if setup != nil {
		setup(m)
	}
	t.Cleanup(func() {
		if err := m.ExpectationsWereMet(); err != nil {
			t.Errorf("sqlmock: %v", err)
		}
		_ = sqlDB.Close()
	})
	return &conn.DB{DB: sqlDB}
}

func TestService_GetWithdrawals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uid := int64(3)
	list := []*model.Withdrawal{{OrderID: "79927398713", Sum: 50}}

	tests := []struct {
		name     string
		setup    func(w *mock.MockWithdrawalRepository)
		wantCode int
		wantLen  int
		wantErr  bool
	}{
		{
			name: "error",
			setup: func(w *mock.MockWithdrawalRepository) {
				w.EXPECT().GetByUserID(ctx, uid).Return(nil, errors.New("fail"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "ok",
			setup: func(w *mock.MockWithdrawalRepository) {
				w.EXPECT().GetByUserID(ctx, uid).Return(list, nil)
			},
			wantCode: http.StatusOK,
			wantLen:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := mock.NewMockUserRepository(ctrl)
			o := mock.NewMockOrderRepository(ctrl)
			w := mock.NewMockWithdrawalRepository(ctrl)
			tt.setup(w)
			s := NewService(nil, repository.Repositories{User: u, Order: o, Withdrawal: w})
			out := s.GetWithdrawals(ctx, GetWithdrawalsInput{UserID: uid})
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d", out.Code)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("want err")
			}
			if !tt.wantErr && len(out.Withdrawals) != tt.wantLen {
				t.Errorf("len = %d", len(out.Withdrawals))
			}
		})
	}
}

func TestService_AddWithdrawal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uid := int64(1)
	validOrder := "79927398713"

	tests := []struct {
		name     string
		mockSQL  func(sqlmock.Sqlmock)
		in       WithdrawalInput
		setup    func(u *mock.MockUserRepository, w *mock.MockWithdrawalRepository)
		wantCode int
		wantErr  bool
		check    func(t *testing.T, out WithdrawOutput)
	}{
		{
			name:     "empty order id",
			in:       WithdrawalInput{UserID: uid, OrderID: "  ", Amount: 10},
			wantCode: http.StatusBadRequest,
			wantErr:  true,
		},
		{
			name:     "invalid luhn",
			in:       WithdrawalInput{UserID: uid, OrderID: "111", Amount: 10},
			wantCode: http.StatusUnprocessableEntity,
			wantErr:  true,
		},
		{
			name: "get user error",
			in:   WithdrawalInput{UserID: uid, OrderID: validOrder, Amount: 5},
			setup: func(u *mock.MockUserRepository, _ *mock.MockWithdrawalRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(nil, errors.New("db"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "insufficient balance",
			in:   WithdrawalInput{UserID: uid, OrderID: validOrder, Amount: 999},
			setup: func(u *mock.MockUserRepository, _ *mock.MockWithdrawalRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(&model.User{ID: uid, Balance: 10}, nil)
			},
			wantCode: http.StatusPaymentRequired,
			wantErr:  true,
		},
		{
			name: "BeginTx fails",
			mockSQL: func(m sqlmock.Sqlmock) {
				m.ExpectBegin().WillReturnError(errors.New("db closed"))
			},
			in: WithdrawalInput{UserID: uid, OrderID: validOrder, Amount: 1},
			setup: func(u *mock.MockUserRepository, _ *mock.MockWithdrawalRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(&model.User{ID: uid, Balance: 100}, nil)
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "Withdrawal.Add fails",
			mockSQL: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectRollback()
			},
			in: WithdrawalInput{UserID: uid, OrderID: validOrder, Amount: 25},
			setup: func(u *mock.MockUserRepository, w *mock.MockWithdrawalRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(&model.User{ID: uid, Balance: 100}, nil)
				w.EXPECT().Add(ctx, gomock.Any(), gomock.Any()).Return(errors.New("insert fail"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "IncrementWithdrawn fails",
			mockSQL: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectRollback()
			},
			in: WithdrawalInput{UserID: uid, OrderID: validOrder, Amount: 25},
			setup: func(u *mock.MockUserRepository, w *mock.MockWithdrawalRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(&model.User{ID: uid, Balance: 100}, nil)
				w.EXPECT().Add(ctx, gomock.Any(), gomock.Any()).Return(nil)
				u.EXPECT().IncrementWithdrawn(ctx, gomock.Any(), gomock.Any()).Return(errors.New("upd fail"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			mockSQL: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectCommit()
			},
			in: WithdrawalInput{UserID: uid, OrderID: validOrder, Amount: 30},
			setup: func(u *mock.MockUserRepository, w *mock.MockWithdrawalRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(&model.User{ID: uid, Balance: 100}, nil)
				w.EXPECT().Add(ctx, gomock.Any(), gomock.AssignableToTypeOf(&model.Withdrawal{})).DoAndReturn(
					func(_ context.Context, _ *conn.Tx, wd *model.Withdrawal) error {
						if wd.OrderID != validOrder || wd.Sum != 30 || wd.UserID != uid {
							t.Errorf("withdrawal fields: %+v", wd)
						}
						return nil
					})
				u.EXPECT().IncrementWithdrawn(ctx, gomock.Any(), gomock.AssignableToTypeOf(&model.Withdrawal{})).Return(nil)
			},
			wantCode: http.StatusOK,
			check: func(t *testing.T, out WithdrawOutput) {
				if out.Withdraw == nil || out.Withdraw.Sum != 30 {
					t.Errorf("Withdraw = %+v", out.Withdraw)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			u := mock.NewMockUserRepository(ctrl)
			o := mock.NewMockOrderRepository(ctrl)
			w := mock.NewMockWithdrawalRepository(ctrl)
			if tt.setup != nil {
				tt.setup(u, w)
			}
			var db *conn.DB
			switch {
			case tt.mockSQL != nil:
				db = testMockConn(t, tt.mockSQL)
			case tt.name == "empty order id" || tt.name == "invalid luhn" ||
				tt.name == "get user error" || tt.name == "insufficient balance":
				db = &conn.DB{}
			default:
				t.Fatalf("case %q: set mockSQL", tt.name)
			}
			s := NewService(db, repository.Repositories{User: u, Order: o, Withdrawal: w})
			out := s.AddWithdrawal(ctx, tt.in)
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", out.Code, tt.wantCode)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("want err")
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}
}
