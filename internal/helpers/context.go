package helpers

import "context"

type ctxKey string

const UserIDKey ctxKey = "user_id"
const RoleIDKey ctxKey = "role_id"

func SetUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

func GetUserID(ctx context.Context) int {
	if v := ctx.Value(UserIDKey); v != nil {
		if id, ok := v.(int); ok {
			return id
		}
	}
	return 0
}
