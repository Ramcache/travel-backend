package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			claims, err := helpers.ParseJWT(secret, tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			rawUID, ok := claims["user_id"]
			if !ok {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			uidFloat, ok := rawUID.(float64)
			if !ok {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			rawRole := claims["role_id"]
			roleFloat, ok := rawRole.(float64)
			if !ok {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), helpers.UserIDKey, int(uidFloat))
			ctx = context.WithValue(ctx, helpers.RoleIDKey, int(roleFloat))

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
