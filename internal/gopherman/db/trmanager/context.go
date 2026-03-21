package trmanager

import (
	"context"
	"gophermart-loyalty/internal/gopherman/db/conn"
)

type txKey struct {
}

func WithTx(ctx context.Context, tx *conn.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
func TxFromContext(ctx context.Context) (*conn.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*conn.Tx)
	return tx, ok
}
