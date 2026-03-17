package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrOrderExistsOwn = errors.New("order already exists for this user")

var ErrOrderExistsOther = errors.New("order already exists for another user")

type OrderRepository interface {
	Add(ctx context.Context, userID int64, orderID string) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
	GetOrdersPendingAccrual(ctx context.Context) ([]*model.Order, error)
	UpdateOrderAccrual(ctx context.Context, tx *conn.Tx, order *model.Order) error
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
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != pgerrcode.UniqueViolation {
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
	var list []*model.Order
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, status, accrual, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*model.Order{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o model.Order
		var accrual sql.NullFloat64
		o.UserID = userID
		err = rows.Scan(&o.ID, &o.OrderID, &o.Status, &accrual, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if accrual.Valid {
			o.Accrual = new(accrual.Float64)
		}
		list = append(list, new(o))
	}
	return list, rows.Err()
}

func (r *orderRepo) GetOrdersPendingAccrual(ctx context.Context) ([]*model.Order, error) {
	var lastID int64
	const chunkSize = 1
	var list []*model.Order
	for {
		rows, err := r.db.QueryContext(ctx,
			`SELECT id, order_id, user_id, status, created_at, updated_at FROM orders WHERE status IN ('NEW', 'PROCESSING') AND id > $1 ORDER BY id ASC LIMIT $2`,
			lastID, chunkSize)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return list, nil
			}
			return nil, err
		}
		var chunkCount int
		for rows.Next() {
			var o model.Order
			var accrual sql.NullFloat64
			err = rows.Scan(&o.ID, &o.OrderID, &o.UserID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
			if err != nil {
				_ = rows.Close()
				return nil, err
			}
			if accrual.Valid {
				o.Accrual = new(accrual.Float64)
			}
			list = append(list, &o)
			lastID = o.ID
			chunkCount++
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
		_ = rows.Close()
		if chunkCount < chunkSize {
			break
		}
	}
	return list, nil
}

func (r *orderRepo) UpdateOrderAccrual(ctx context.Context, tx *conn.Tx, order *model.Order) error {
	_, err := tx.ExecContext(ctx, "UPDATE orders SET status = $1,accrual = $2 WHERE id = $3;", order.Status, order.Accrual, order.ID)
	if err != nil {
		return fmt.Errorf("update order accrual error: %w", err)
	}
	return nil
}
