package helpers_test

import (
    "testing"
    "time"

    "github.com/Ramcache/travel-backend/internal/helpers"
)

func TestGenerateAndParseJWT(t *testing.T) {
    const secret = "test-secret"
    token, err := helpers.GenerateJWT(secret, 42, "John Doe", 2, time.Hour)
    if err != nil {
        t.Fatalf("GenerateJWT error: %v", err)
    }
    if token == "" {
        t.Fatal("expected non-empty token")
    }

    claims, err := helpers.ParseJWT(secret, token)
    if err != nil {
        t.Fatalf("ParseJWT error: %v", err)
    }

    // numeric claims are decoded as float64 in MapClaims
    if got := int(claims["user_id"].(float64)); got != 42 {
        t.Fatalf("user_id claim mismatch: got %d, want 42", got)
    }
    if got := claims["full_name"].(string); got != "John Doe" {
        t.Fatalf("full_name claim mismatch: got %q", got)
    }
    if got := int(claims["role_id"].(float64)); got != 2 {
        t.Fatalf("role_id claim mismatch: got %d, want 2", got)
    }
    if _, ok := claims["exp"]; !ok {
        t.Fatal("expected exp claim to be present")
    }
}
