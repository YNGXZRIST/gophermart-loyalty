// Package contextkey provides strongly-typed context helpers.
package contextkey

import (
	"context"
)

type key int

const userIDKey key = 0

// WithUserID returns a context carrying authenticated user ID.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts authenticated user ID from context.
func UserIDFromContext(ctx context.Context) (userID int64, ok bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
