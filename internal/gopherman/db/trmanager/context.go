package trmanager

import (
	"context"
)

type txKey struct {
}

func WithTx(ctx context.Context, tx *Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
func TxFromContext(ctx context.Context) (*Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*Tx)
	return tx, ok
}
