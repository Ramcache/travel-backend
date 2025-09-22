package helpers

import (
	"errors"
	"time"
)

func ParseFlexibleDate(s string) (time.Time, error) {
	// пробуем YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// пробуем RFC3339
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("unsupported date format")
}
