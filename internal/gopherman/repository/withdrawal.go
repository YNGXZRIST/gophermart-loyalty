package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
)

// WithdrawalRepo provides persistence operations for withdrawals.
type WithdrawalRepo struct {
	repoBase
}

const (
	// WithdrawalAddQuery inserts a withdrawal record.
	WithdrawalAddQuery = `INSERT INTO withdrawals (user_id, order_id, sum) VALUES ($1, $2, $3)`
	// WithdrawalGetByUserIDQuery lists withdrawals for a user.
	WithdrawalGetByUserIDQuery = `SELECT order_id, sum, created_at, updated_at FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC`
)

// NewWithdrawalRepository creates a withdrawal repository bound to DB.
func NewWithdrawalRepository(db *conn.DB) *WithdrawalRepo {
	return &WithdrawalRepo{repoBase: repoBase{db: db}}
}

// Add persists a new withdrawal entry.
func (r *WithdrawalRepo) Add(ctx context.Context, w *model.Withdrawal) error {
	_, err := r.repoBase.q(ctx).ExecContext(ctx,
		WithdrawalAddQuery,
		w.UserID, w.OrderID, w.Sum)
	if err != nil {
		return labelerrors.NewLabelError(labelRepository+".Withdrawal.Add.Exec", err)
	}
	return nil
}

// GetByUserID returns all withdrawals for the specified user.
func (r *WithdrawalRepo) GetByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	rows, err := r.repoBase.q(ctx).QueryContext(ctx,
		WithdrawalGetByUserIDQuery,
		userID)
	if err != nil {
		return nil, labelerrors.NewLabelError(labelRepository+".Withdrawal.GetByUserID.Query", err)
	}
	defer rows.Close()

	var list []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		w.UserID = userID
		err = rows.Scan(&w.OrderID, &w.Sum, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, labelerrors.NewLabelError(labelRepository+".Withdrawal.GetByUserID.Scan", err)
		}
		list = append(list, new(w))
	}
	if err := rows.Err(); err != nil {
		return nil, labelerrors.NewLabelError(labelRepository+".Withdrawal.GetByUserID.Rows", err)
	}
	return list, nil
}
