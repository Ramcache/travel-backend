package middleware

import (
	"fmt"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"net/http"
)

func RoleAuth(requiredRoles ...int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Context().Value(helpers.RoleIDKey)
			if raw == nil {
				helpers.Error(w, http.StatusForbidden, "no role")
				return
			}

			roleID, ok := raw.(int)
			if !ok {
				helpers.Error(w, http.StatusForbidden, "invalid role type")
				return
			}

			allowed := false
			for _, rr := range requiredRoles {
				if roleID == rr {
					allowed = true
					break
				}
			}
			if !allowed {
				helpers.Error(w, http.StatusForbidden,
					fmt.Sprintf("forbidden: need role %v", requiredRoles))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
