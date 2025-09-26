package helpers_test

import (
    "errors"
    "testing"

    "github.com/go-playground/validator/v10"
    "github.com/jackc/pgx/v5/pgconn"

    "github.com/Ramcache/travel-backend/internal/helpers"
)

// A sample DTO to validate
type reg struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=6"`
}

func TestFieldsFromValidationErr(t *testing.T) {
    v := validator.New()

    // Missing required fields -> 'required' messages
    err := v.Struct(reg{})
    fields := helpers.FieldsFromValidationErr(err)
    if fields["Email"] == "" || fields["Password"] == "" {
        t.Fatalf("expected validation messages, got: %#v", fields)
    }

    // Wrong email -> 'email' message (set required satisfied)
    err = v.Struct(reg{Email: "bad@", Password: "123456"})
    fields = helpers.FieldsFromValidationErr(err)
    if got := fields["Email"]; got == "" {
        t.Fatalf("expected email error, got none: %#v", fields)
    }

    // Too short password -> 'min' message
    err = v.Struct(reg{Email: "john@example.com", Password: "123"})
    fields = helpers.FieldsFromValidationErr(err)
    if got := fields["Password"]; got == "" {
        t.Fatalf("expected min length error, got none: %#v", fields)
    }
}

func TestMapPgErr(t *testing.T) {
    status, code, msg := helpers.MapPgErr(&pgconn.PgError{Code: "23505"})
    if status != 409 || code != "conflict" || msg == "" {
        t.Fatalf("unique violation mapping mismatch: %d, %s, %q", status, code, msg)
    }

    status, code, msg = helpers.MapPgErr(&pgconn.PgError{Code: "23503"})
    if status != 400 || code != "bad_request" || msg == "" {
        t.Fatalf("fk violation mapping mismatch: %d, %s, %q", status, code, msg)
    }

    // Unknown error -> internal_error
    status, code, _ = helpers.MapPgErr(errors.New("boom"))
    if status != 500 || code != "internal_error" {
        t.Fatalf("default mapping mismatch: %d, %s", status, code)
    }
}
