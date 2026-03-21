package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
)

// AccrualWriter applies accrual updates within a single transaction.
type AccrualWriter struct {
	mgr *trmanager.Manager
	rep Repositories
}

// NewAccrualWriter constructs transactional writer for accrual updates.
func NewAccrualWriter(mgr *trmanager.Manager, rep Repositories) *AccrualWriter {
	return &AccrualWriter{mgr: mgr, rep: rep}
}

// ApplyResult updates order state and increments user balance when needed.
func (w *AccrualWriter) ApplyResult(ctx context.Context, order *model.Order) error {
	return w.mgr.WithinTx(ctx, nil, func(ctx context.Context) error {
		if err := w.rep.Order.UpdateOrderAccrual(ctx, order); err != nil {
			return err
		}
		if order.Accrual != nil {
			return w.rep.User.IncrementBalance(ctx, order.UserID, *order.Accrual)
		}
		return nil
	})
}
