package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"
)

type WithdrawalRepository interface {
	Add(ctx context.Context, tx *conn.Tx, withdrawal *model.Withdrawal) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
}
type WithdrawalRepo struct {
	db *conn.DB
}

const (
	withdrawalAddQuery         = `INSERT INTO withdrawals (user_id, order_id, sum) VALUES ($1, $2, $3)`
	withdrawalGetByUserIDQuery = `SELECT order_id, sum, created_at, updated_at FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC`
)

func NewWithdrawalRepository(db *conn.DB) *WithdrawalRepo {
	return &WithdrawalRepo{db: db}
}

func (r *WithdrawalRepo) Add(ctx context.Context, tx *conn.Tx, w *model.Withdrawal) error {

	_, err := tx.ExecContext(ctx,
		withdrawalAddQuery,
		w.UserID, w.OrderID, w.Sum)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".Withdrawal.Add.Exec", err)
	}
	return nil
}
func (r *WithdrawalRepo) GetByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx,
		withdrawalGetByUserIDQuery,
		userID)
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".Withdrawal.GetByUserID.Query", err)
	}
	defer rows.Close()

	var list []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		w.UserID = userID
		err = rows.Scan(&w.OrderID, &w.Sum, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, labelerrors.NewLabelError(constant.LabelRepository+".Withdrawal.GetByUserID.Scan", err)
		}
		list = append(list, new(w))
	}
	if err := rows.Err(); err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".Withdrawal.GetByUserID.Rows", err)
	}
	return list, nil
}
