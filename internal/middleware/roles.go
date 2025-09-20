package middleware

import (
	"net/http"
	"strconv"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

func RoleAuth(requiredRoles ...int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleIDRaw := r.Context().Value("role_id")
			if roleIDRaw == nil {
				helpers.Error(w, http.StatusForbidden, "no role")
				return
			}
			roleID, ok := roleIDRaw.(int)
			if !ok {
				helpers.Error(w, http.StatusForbidden, "invalid role")
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
					"forbidden: need role "+strconv.Itoa(requiredRoles[0]))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
