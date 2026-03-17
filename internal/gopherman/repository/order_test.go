package repository

import (
	"context"
	"testing"

	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository/mock"

	"github.com/golang/mock/gomock"
)

func TestOrderRepository_GetByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mock.NewMockOrderRepository(ctrl)
	repos := Repositories{Order: mockOrder}
	ctx := context.Background()
	userID := int64(10)

	orders := []*model.Order{{UserID: userID, OrderID: "123", Status: "NEW"}}
	mockOrder.EXPECT().
		GetByUserID(gomock.Any(), userID).
		Return(orders, nil)

	got, err := repos.Order.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(got) != 1 || got[0].OrderID != "123" {
		t.Errorf("GetByUserID: got %v, want one order 123", got)
	}
}

func TestOrderRepository_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mock.NewMockOrderRepository(ctrl)
	repos := Repositories{Order: mockOrder}
	ctx := context.Background()

	mockOrder.EXPECT().
		Add(gomock.Any(), int64(5), "order-1").
		Return(nil)

	if err := repos.Order.Add(ctx, 5, "order-1"); err != nil {
		t.Fatalf("Add: %v", err)
	}
}
