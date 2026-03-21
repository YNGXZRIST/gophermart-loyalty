package repository

import (
	"context"
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
	"gophermart-loyalty/internal/gopherman/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	OrderGetOwnerQuery                  = `SELECT user_id FROM orders WHERE order_id = $1`
	OrderAddOrderQuery                  = `INSERT INTO orders (user_id, order_id, status) VALUES ($1, $2, 'NEW')`
	OrderGetByUidQuery                  = `SELECT id, order_id, status, accrual, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	OrderGetPendingOrdersWithLimitQuery = `SELECT id, order_id, user_id, status, created_at, updated_at FROM orders WHERE status IN ('NEW', 'PROCESSING') AND id > $1 ORDER BY id ASC LIMIT $2`
	OrderUpdatePendingOrderQuery        = "UPDATE orders SET status = $1,accrual = $2 WHERE id = $3;"
)

var ErrOrderExistsOwn = errors.New("order already exists for this user")

var ErrOrderExistsOther = errors.New("order already exists for another user")

type OrderRepository interface {
	Add(ctx context.Context, userID int64, orderID string) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
	GetOrdersPendingAccrual(ctx context.Context) ([]*model.Order, error)
	UpdateOrderAccrual(ctx context.Context, order *model.Order) error
}

type orderRepo struct {
	repoBase
}

func NewOrderRepository(db *conn.DB) OrderRepository {
	return &orderRepo{repoBase: repoBase{db: db}}
}

func (r *orderRepo) Add(ctx context.Context, userID int64, orderID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".Order.Add.BeginTx", err)
	}
	defer tx.Rollback()
	var ownerID sql.NullInt64
	err = r.db.QueryRowContext(ctx, OrderGetOwnerQuery, orderID).Scan(&ownerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return labelerrors.NewLabelError(constant.LabelRepository+".Order.Add.CheckOwner", err)
	}
	if ownerID.Valid {
		if ownerID.Int64 == userID {
			return labelerrors.NewLabelError(constant.LabelRepository+".Order.Add.ExistsOwn", ErrOrderExistsOwn)
		}
		return labelerrors.NewLabelError(constant.LabelRepository+".Order.Add.ExistsOther", ErrOrderExistsOther)
	}
	_, err = r.repoBase.q(ctx).ExecContext(ctx,
		OrderAddOrderQuery,
		userID, orderID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != pgerrcode.UniqueViolation {
			return labelerrors.NewLabelError(constant.LabelRepository+".Order.Add.Insert", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".Order.Add.Commit", err)
	}
	return nil
}

func (r *orderRepo) GetByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	var list []*model.Order
	rows, err := r.repoBase.q(ctx).QueryContext(ctx,
		OrderGetByUidQuery,
		userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*model.Order{}, nil
		}
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".Order.GetByUserID.Query", err)
	}
	defer rows.Close()
	for rows.Next() {
		var o model.Order
		var accrual sql.NullFloat64
		o.UserID = userID
		err = rows.Scan(&o.ID, &o.OrderID, &o.Status, &accrual, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, labelerrors.NewLabelError(constant.LabelRepository+".Order.GetByUserID.Scan", err)
		}
		if accrual.Valid {
			o.Accrual = new(accrual.Float64)
		}
		list = append(list, new(o))
	}
	if err := rows.Err(); err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelRepository+".Order.GetByUserID.Rows", err)
	}
	return list, nil
}

func (r *orderRepo) GetOrdersPendingAccrual(ctx context.Context) ([]*model.Order, error) {
	var lastID int64
	const chunkSize = 1
	var list []*model.Order
	for {
		rows, err := r.repoBase.q(ctx).QueryContext(ctx,
			OrderGetPendingOrdersWithLimitQuery,
			lastID, chunkSize)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return list, nil
			}
			return nil, labelerrors.NewLabelError(constant.LabelRepository+".Order.GetOrdersPendingAccrual.Query", err)
		}
		var chunkCount int
		for rows.Next() {
			var o model.Order
			var accrual sql.NullFloat64
			err = rows.Scan(&o.ID, &o.OrderID, &o.UserID, &o.Status, &o.CreatedAt, &o.UpdatedAt)
			if err != nil {
				_ = rows.Close()
				return nil, labelerrors.NewLabelError(constant.LabelRepository+".Order.GetOrdersPendingAccrual.Scan", err)
			}
			if accrual.Valid {
				o.Accrual = new(accrual.Float64)
			}
			list = append(list, &o)
			lastID = o.ID
			chunkCount++
		}
		if err = rows.Err(); err != nil {
			return nil, labelerrors.NewLabelError(constant.LabelRepository+".Order.GetOrdersPendingAccrual.Rows", err)
		}
		_ = rows.Close()
		if chunkCount < chunkSize {
			break
		}
	}
	return list, nil
}
func (r *orderRepo) UpdateOrderAccrual(ctx context.Context, order *model.Order) error {

	_, err := r.repoBase.q(ctx).ExecContext(ctx, OrderUpdatePendingOrderQuery, order.Status, order.Accrual, order.ID)
	if err != nil {
		return labelerrors.NewLabelError(constant.LabelRepository+".Order.UpdateOrderAccrual.Exec", err)
	}
	return nil
}
