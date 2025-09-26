package helpers_test

import (
    "testing"
    "time"

    "github.com/Ramcache/travel-backend/internal/helpers"
)

func TestParseFlexibleDate_SupportedFormats(t *testing.T) {
    // ISO date
    d, err := helpers.ParseFlexibleDate("2025-09-26")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if d.Year() != 2025 || d.Month() != time.September || d.Day() != 26 {
        t.Fatalf("parsed date mismatch: got %v", d)
    }

    // RFC3339
    d, err = helpers.ParseFlexibleDate("2025-09-26T10:15:30Z")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if d.UTC().Hour() != 10 || d.UTC().Minute() != 15 || d.UTC().Second() != 30 {
        t.Fatalf("parsed datetime mismatch: got %v", d)
    }
}

func TestParseFlexibleDate_Unsupported(t *testing.T) {
    if _, err := helpers.ParseFlexibleDate("26/09/2025"); err == nil {
        t.Fatal("expected error for unsupported date format")
    }
}
