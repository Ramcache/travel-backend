package helpers_test

import (
	"encoding/json"
	"testing"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

func TestPaginatedResponseJSON(t *testing.T) {
	resp := helpers.PaginatedResponse[int]{
		Total: 2,
		Items: []int{1, 2},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	expected := `{"total":2,"items":[1,2]}`
	if string(data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data))
	}
}
