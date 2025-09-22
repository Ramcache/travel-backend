package helpers

import "context"

type ctxKey string

const userIDKey ctxKey = "user_id"

func SetUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func GetUserID(ctx context.Context) int {
	v := ctx.Value(userIDKey)
	if id, ok := v.(int); ok {
		return id
	}
	return 0
}
