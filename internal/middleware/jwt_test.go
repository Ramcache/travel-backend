package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/Ramcache/travel-backend/internal/helpers"
    "github.com/Ramcache/travel-backend/internal/middleware"
)

func TestJWTAuth_SetsContext(t *testing.T) {
    const secret = "mw-secret"
    token, err := helpers.GenerateJWT(secret, 7, "Tester", 3, time.Hour)
    if err != nil {
        t.Fatalf("GenerateJWT error: %v", err)
    }

    called := false
    var gotUID, gotRole any

    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
        gotUID = r.Context().Value(middleware.UserIDKey)
        gotRole = r.Context().Value("role_id") // middleware stores role under string key
        w.WriteHeader(http.StatusOK)
    })

    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/x", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    middleware.JWTAuth(secret)(next).ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200 from next, got %d", rr.Code)
    }
    if !called {
        t.Fatal("next handler was not called")
    }
    if gotUID.(int) != 7 || gotRole.(int) != 3 {
        t.Fatalf("context values mismatch: uid=%v role=%v", gotUID, gotRole)
    }
}

func TestJWTAuth_MissingOrInvalidToken(t *testing.T) {
    const secret = "mw-secret"

    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // Missing
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/x", nil)
    middleware.JWTAuth(secret)(next).ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 for missing token, got %d", rr.Code)
    }

    // Invalid
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/x", nil)
    req.Header.Set("Authorization", "Bearer invalid.token")
    middleware.JWTAuth(secret)(next).ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 for invalid token, got %d", rr.Code)
    }
}
