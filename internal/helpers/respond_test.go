package helpers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

func TestJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	helpers.JSON(rr, http.StatusCreated, map[string]int{"x": 1})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status code mismatch: %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content type mismatch: %q", ct)
	}
	var env helpers.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("json decode error: %v", err)
	}
	if !env.Success {
		t.Fatal("expected success=true")
	}
	m, ok := env.Data.(map[string]interface{})
	if !ok || int(m["x"].(float64)) != 1 {
		t.Fatalf("unexpected data: %#v", env.Data)
	}
}

func TestError(t *testing.T) {
	rr := httptest.NewRecorder()
	helpers.Error(rr, http.StatusForbidden, "nope")

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status code mismatch: %d", rr.Code)
	}
	var env helpers.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("json decode error: %v", err)
	}
	if env.Success {
		t.Fatal("expected success=false")
	}
	ed, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected data shape: %#v", env.Data)
	}
	if ed["code"] != "forbidden" || ed["message"] == "" {
		t.Fatalf("unexpected error payload: %#v", ed)
	}
}

func TestValidationError(t *testing.T) {
	rr := httptest.NewRecorder()
	helpers.ValidationError(rr, map[string]string{"Email": "Некорректный email"})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status code mismatch: %d", rr.Code)
	}
	var env helpers.Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("json decode error: %v", err)
	}
	if env.Success {
		t.Fatal("expected success=false")
	}
	ed := env.Data.(map[string]interface{})
	if ed["code"] != "validation_failed" {
		t.Fatalf("unexpected code: %v", ed["code"])
	}
}
