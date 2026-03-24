package repository

import (
	"context"
	"fmt"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
)

// WithdrawalWriter applies withdrawal side effects atomically.
type WithdrawalWriter struct {
	mgr *trmanager.Manager
	rep Repositories
}

// NewWithdrawalWriter constructs transactional writer for withdrawals.
func NewWithdrawalWriter(mgr *trmanager.Manager, rep Repositories) *WithdrawalWriter {
	return &WithdrawalWriter{mgr: mgr, rep: rep}
}

// MakeWithdrawal creates withdrawal and updates user totals in one transaction.
func (w *WithdrawalWriter) MakeWithdrawal(ctx context.Context, wd *model.Withdrawal) error {
	return w.mgr.WithinTx(ctx, nil, func(ctx context.Context) error {
		if err := w.rep.Withdrawal.Add(ctx, wd); err != nil {
			return fmt.Errorf("add withdrawal: %w", err)
		}
		if err := w.rep.User.IncrementWithdrawn(ctx, wd); err != nil {
			return fmt.Errorf("increment withdrawn: %w", err)
		}
		return nil
	})
}
