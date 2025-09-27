package helpers_test

import (
	"context"
	"testing"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

func TestSetAndGetUserID(t *testing.T) {
	ctx := context.Background()
	if got := helpers.GetUserID(ctx); got != 0 {
		t.Fatalf("expected zero value when user id not set, got %d", got)
	}

	ctx = helpers.SetUserID(ctx, 42)
	if got := helpers.GetUserID(ctx); got != 42 {
		t.Fatalf("expected user id 42, got %d", got)
	}
}
