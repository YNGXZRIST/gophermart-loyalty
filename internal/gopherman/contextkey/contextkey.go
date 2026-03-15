package contextkey

import "context"

type key int

const userIDKey key = 0

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (userID int64, ok bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
