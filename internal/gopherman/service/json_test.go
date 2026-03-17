package service

import (
	"encoding/json"
	"testing"
	"time"

	"gophermart-loyalty/internal/gopherman/model"
)

func TestOrdersJSON(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 2, 15, 0, 0, 0, time.UTC)
	accr := 10.5
	tests := []struct {
		name    string
		orders  []*model.Order
		want    string
		wantErr bool
	}{
		{
			name:   "nil slice is empty json array",
			orders: nil,
			want:   "[]",
		},
		{
			name:   "empty slice",
			orders: []*model.Order{},
			want:   "[]",
		},
		{
			name: "single order",
			orders: []*model.Order{
				{OrderID: "79927398713", Status: "PROCESSED", Accrual: &accr, CreatedAt: now},
			},
			want: `[{"number":"79927398713","status":"PROCESSED","accrual":10.5,"uploaded_at":"2024-01-02T15:00:00Z"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := OrdersJSON(tt.orders)
			if (err != nil) != tt.wantErr {
				t.Fatalf("OrdersJSON() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if string(got) != tt.want {
				t.Errorf("OrdersJSON() = %s, want %s", got, tt.want)
			}
			if !json.Valid(got) {
				t.Error("OrdersJSON() invalid JSON")
			}
		})
	}
}

func TestWithdrawalsJSON(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		withdrawals []*model.Withdrawal
		want        string
		wantErr     bool
	}{
		{
			name:        "nil is empty array",
			withdrawals: nil,
			want:        "[]",
		},
		{
			name:        "empty slice",
			withdrawals: []*model.Withdrawal{},
			want:        "[]",
		},
		{
			name: "one withdrawal",
			withdrawals: []*model.Withdrawal{
				{OrderID: "79927398713", Sum: 100.25, UpdatedAt: now},
			},
			want: `[{"order":"79927398713","sum":100.25,"processed_at":"2024-03-01T12:00:00Z"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := WithdrawalsJSON(tt.withdrawals)
			if (err != nil) != tt.wantErr {
				t.Fatalf("WithdrawalsJSON() err = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("WithdrawalsJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}
