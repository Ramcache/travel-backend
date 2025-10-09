package helpers

import (
	"errors"
	"fmt"
	"time"
)

func ParseFlexibleDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("unsupported date format")
}

func ParseDateAny(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	// 1) строгий YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// 2) RFC3339 (2026-01-16T00:00:00Z)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// 3) часто присылают без Z
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date: %q", s)
}
