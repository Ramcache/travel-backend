package middleware_test

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ramcache/travel-backend/internal/middleware"
)

func TestRoleAuth_AllowsRequiredRole(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RoleAuth(2)(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), helpers.RoleIDKey, 2))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for role 2, got %d", rr.Code)
	}
}

func TestRoleAuth_ForbiddenForOtherRoles(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.RoleAuth(2)(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), helpers.RoleIDKey, 1))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for role 1, got %d", rr.Code)
	}
}
