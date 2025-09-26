package helpers_test

import (
    "testing"

    "github.com/Ramcache/travel-backend/internal/helpers"
)

func TestHashAndCheckPassword(t *testing.T) {
    hash, err := helpers.HashPassword("s3cr3t")
    if err != nil {
        t.Fatalf("HashPassword returned error: %v", err)
    }
    if hash == "" {
        t.Fatal("expected non-empty hash")
    }
    if ok := helpers.CheckPassword(hash, "s3cr3t"); !ok {
        t.Error("expected password to match the hash")
    }
    if ok := helpers.CheckPassword(hash, "wrong"); ok {
        t.Error("expected wrong password to NOT match the hash")
    }
}
