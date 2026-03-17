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

func TestService_GetBalance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	uid := int64(9)

	tests := []struct {
		name     string
		setup    func(u *mock.MockUserRepository)
		wantCode int
		want     model.BalanceResponse
		wantErr  bool
	}{
		{
			name: "GetByID fails",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(nil, errors.New("not found"))
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name: "success",
			setup: func(u *mock.MockUserRepository) {
				u.EXPECT().GetByID(ctx, uid).Return(&model.User{
					ID:        uid,
					Balance:   500.5,
					Withdrawn: 100,
				}, nil)
			},
			wantCode: http.StatusOK,
			want: model.BalanceResponse{
				Current:   500.5,
				Withdrawn: 100,
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
			tt.setup(u)
			s := NewService(nil, repository.Repositories{User: u, Order: o, Withdrawal: w})
			out := s.GetBalance(ctx, BalanceInput{UserID: uid})
			if out.Code != tt.wantCode {
				t.Errorf("Code = %d", out.Code)
			}
			if tt.wantErr && out.Err == nil {
				t.Error("want err")
			}
			if !tt.wantErr {
				if out.Balance.Current != tt.want.Current || out.Balance.Withdrawn != tt.want.Withdrawn {
					t.Errorf("Balance = %+v, want %+v", out.Balance, tt.want)
				}
			}
		})
	}
}
