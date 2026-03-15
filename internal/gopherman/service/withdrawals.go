package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gophermart-loyalty/internal/gopherman/model"
	"net/http"
)

type GetWithdrawalsInput struct {
	UserID int64
}
type GetWithdrawalsOutput struct {
	Response
	Withdrawals []*model.Withdrawal
}

func (s *Service) GetWithdrawals(ctx context.Context, in GetWithdrawalsInput) GetWithdrawalsOutput {
	withdrawals, err := s.Rep.Withdrawal.GetByUserID(ctx, in.UserID)
	if err != nil {
		return GetWithdrawalsOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("get withdrawals: %w", err),
			},
		}
	}
	return GetWithdrawalsOutput{
		Response: Response{
			Code: http.StatusOK,
		},
		Withdrawals: withdrawals,
	}
}
func WithdrawalsJSON(withdrawals []*model.Withdrawal) ([]byte, error) {
	if withdrawals == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(withdrawals)
}
