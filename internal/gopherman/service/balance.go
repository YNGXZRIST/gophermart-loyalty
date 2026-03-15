package service

import (
	"context"
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/pkg/luhn"
	"net/http"
	"strings"
)

type BalanceInput struct {
	UserID int64
}
type BalanceOutput struct {
	Response
	Balance model.BalanceResponse
}
type WithdrawalInput struct {
	UserID  int64
	OrderID string
	Amount  float64
}
type WithdrawOutput struct {
	Response
	Withdraw *model.Withdrawal
}

func (s *Service) GetBalance(ctx context.Context, in BalanceInput) BalanceOutput {
	user, err := s.Rep.User.GetByID(ctx, in.UserID)
	if err != nil {
		return BalanceOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  errors.New("unnable to get user"),
			},
		}
	}
	return BalanceOutput{
		Response: Response{
			Code: http.StatusOK,
		},
		Balance: model.BalanceResponse{
			Current:   user.Balance,
			Withdrawn: user.Withdrawn,
		},
	}
}

func (s *Service) AddWithdrawal(ctx context.Context, in WithdrawalInput) WithdrawOutput {
	orderID := strings.TrimSpace(in.OrderID)
	if orderID == "" {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusBadRequest,
				Err:  errors.New("order id is required"),
			},
		}
	}
	if !luhn.Validate(orderID) {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusUnprocessableEntity,
				Err:  errors.New("invalid order number"),
			},
		}
	}
	user, err := s.Rep.User.GetByID(ctx, in.UserID)
	if err != nil {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("get user error: %w", err),
			},
		}
	}
	if user.Balance < in.Amount {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusPaymentRequired,
				Err:  errors.New("insufficient balance"),
			},
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("begin tx: %w", err),
			},
		}
	}
	defer tx.Rollback()
	withdrawal := &model.Withdrawal{
		UserID:  in.UserID,
		OrderID: orderID,
		Sum:     in.Amount,
	}
	err = s.Rep.Withdrawal.Add(ctx, tx, withdrawal)
	if err != nil {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("add withdrawal: %w", err),
			},
		}
	}
	err = s.Rep.User.IncrementWithdrawn(ctx, tx, withdrawal)
	if err != nil {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("increment withdrawn: %w", err),
			},
		}
	}
	err = tx.Commit()
	if err != nil {
		return WithdrawOutput{
			Response: Response{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("commit transaction error: %w", err),
			},
		}
	}
	return WithdrawOutput{
		Response: Response{Code: http.StatusOK},
		Withdraw: withdrawal,
	}
}
