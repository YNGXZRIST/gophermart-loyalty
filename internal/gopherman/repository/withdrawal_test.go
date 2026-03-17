package repository

import (
	"context"
	"testing"

	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository/mock"

	"github.com/golang/mock/gomock"
)

func TestWithdrawalRepository_GetByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWithdrawal := mock.NewMockWithdrawalRepository(ctrl)
	repos := Repositories{Withdrawal: mockWithdrawal}
	ctx := context.Background()
	userID := int64(7)

	withdrawals := []*model.Withdrawal{{UserID: userID, OrderID: "w1", Sum: 100}}
	mockWithdrawal.EXPECT().
		GetByUserID(gomock.Any(), userID).
		Return(withdrawals, nil)

	got, err := repos.Withdrawal.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(got) != 1 || got[0].OrderID != "w1" || got[0].Sum != 100 {
		t.Errorf("GetByUserID: got %v, want one withdrawal w1 sum=100", got)
	}
}
