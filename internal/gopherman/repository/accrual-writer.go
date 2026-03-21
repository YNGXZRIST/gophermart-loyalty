package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/trmanager"
	"gophermart-loyalty/internal/gopherman/model"
)

type AccrualWriter struct {
	mgr *trmanager.Manager
	rep Repositories
}

func NewAccrualWriter(mgr *trmanager.Manager, rep Repositories) *AccrualWriter {
	return &AccrualWriter{mgr: mgr, rep: rep}
}

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
