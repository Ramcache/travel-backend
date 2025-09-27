package helpers_test

import (
	"errors"
	"testing"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

func TestErrInvalidInput(t *testing.T) {
	err := helpers.ErrInvalidInput("bad input")
	if err.Error() != "bad input" {
		t.Fatalf("expected error message to be preserved, got %q", err.Error())
	}

	if !helpers.IsInvalidInput(err) {
		t.Fatal("expected error to be recognized as InvalidInputError")
	}

	if helpers.IsInvalidInput(errors.New("other")) {
		t.Fatal("expected IsInvalidInput to return false for other error types")
	}
}
