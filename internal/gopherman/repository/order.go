package repository

import (
	"context"
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrOrderExistsOwn = errors.New("order already exists for this user")

var ErrOrderExistsOther = errors.New("order already exists for another user")

type OrderRepository interface {
	Add(ctx context.Context, userID int64, orderID string) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
}

type orderRepo struct {
	db *conn.DB
}

func NewOrderRepository(db *conn.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Add(ctx context.Context, userID int64, orderID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (user_id, order_id, status) VALUES ($1, $2, 'NEW')`,
		userID, orderID)
	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != pgerrcode.UniqueViolation {
			return err
		}
	}
	var ownerID int64
	err = r.db.QueryRowContext(ctx, `SELECT user_id FROM orders WHERE order_id = $1`, orderID).Scan(&ownerID)
	if err != nil {
		return err
	}
	if ownerID == userID {
		return ErrOrderExistsOwn
	}
	return ErrOrderExistsOther
}

func (r *orderRepo) GetByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT order_id, status, accrual, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Order
	for rows.Next() {
		var o model.Order
		var accrual sql.NullFloat64
		o.UserID = userID
		err = rows.Scan(&o.OrderID, &o.Status, &accrual, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if accrual.Valid {
			v := accrual.Float64
			o.Accrual = &v
		}
		ord := o
		list = append(list, &ord)
	}
	return list, rows.Err()
}
