package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/repository/mock"

	"github.com/golang/mock/gomock"
)

func TestService_GetOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uid := int64(100)
	orders := []*model.Order{{OrderID: "79927398713", Status: "NEW"}}

	tests := []struct {
		name     string
		setup    func(o *mock.MockOrderRepository)
		wantCode int
		wantLen  int
		wantErr  bool
	}{
		{
			name: "repository error",
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().GetByUserID(ctx, uid).Return(nil, errors.New("query failed"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success empty",
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().GetByUserID(ctx, uid).Return([]*model.Order{}, nil)
			},
			wantCode: http.StatusOK,
			wantLen:  0,
		},
		{
			name: "success with orders",
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().GetByUserID(ctx, uid).Return(orders, nil)
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
			tt.setup(o)
			s := NewService(nil, repository.Repositories{User: u, Order: o, Withdrawal: w})
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
		setup    func(o *mock.MockOrderRepository)
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
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().Add(ctx, uid, validOrder).Return(repository.ErrOrderExistsOwn)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "other user order conflict",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().Add(ctx, uid, validOrder).Return(repository.ErrOrderExistsOther)
			},
			wantCode: http.StatusConflict,
			wantErr:  true,
		},
		{
			name: "generic add error",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().Add(ctx, uid, validOrder).Return(errors.New("db"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "accepted new order",
			in:   AddOrderInput{UserID: uid, OrderID: validOrder},
			setup: func(o *mock.MockOrderRepository) {
				o.EXPECT().Add(ctx, uid, validOrder).Return(nil)
			},
			wantCode: http.StatusAccepted,
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
				tt.setup(o)
			}
			s := NewService(nil, repository.Repositories{User: u, Order: o, Withdrawal: w})
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
		})
	}
}
