package repository

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"
)

type WithdrawalRepository interface {
	Add(ctx context.Context, tx *conn.Tx, withdrawal *model.Withdrawal) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error)
}
type WithdrawalRepo struct {
	db *conn.DB
}

func NewWithdrawalRepository(db *conn.DB) *WithdrawalRepo {
	return &WithdrawalRepo{db: db}
}

func (r *WithdrawalRepo) Add(ctx context.Context, tx *conn.Tx, w *model.Withdrawal) error {

	_, err := tx.ExecContext(ctx,
		`INSERT INTO withdrawals (user_id, order_id, sum) VALUES ($1, $2, $3)`,
		w.UserID, w.OrderID, w.Sum)
	if err != nil {
		return err
	}
	return err
}
func (r *WithdrawalRepo) GetByUserID(ctx context.Context, userID int64) ([]*model.Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT order_id, sum, created_at, updated_at FROM withdrawals WHERE user_id = $1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		w.UserID = userID
		err = rows.Scan(&w.OrderID, &w.Sum, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, err
		}
		wth := w
		list = append(list, &wth)
	}
	return list, rows.Err()
}
