package handlers_test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockTripService struct{ mock.Mock }

func (m *MockTripService) List(ctx context.Context, city, ttype, season string) ([]models.Trip, error) {
	args := m.Called(ctx, city, ttype, season)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripService) Get(ctx context.Context, id int) (*models.Trip, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.Trip), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTripService) Create(ctx context.Context, req models.CreateTripRequest) (*models.Trip, error) {
	args := m.Called(ctx, req)
	if v := args.Get(0); v != nil {
		return v.(*models.Trip), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTripService) Update(ctx context.Context, id int, req models.UpdateTripRequest) (*models.Trip, error) {
	args := m.Called(ctx, id, req)
	if v := args.Get(0); v != nil {
		return v.(*models.Trip), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTripService) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTripService) GetMain(ctx context.Context) (*models.Trip, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*models.Trip), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTripService) Popular(ctx context.Context, limit int) ([]models.Trip, error) { // <--- добавлено
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripService) IncrementViews(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTripService) IncrementBuys(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

type tripsResponse struct {
	Success bool          `json:"success"`
	Data    []models.Trip `json:"data"`
}

func TestTripHandler_List(t *testing.T) {
	mockSvc := new(MockTripService)
	log := zaptest.NewLogger(t).Sugar()

	h := handlers.NewTripHandler(mockSvc, log)

	// мок ответа
	mockSvc.On("List", mock.Anything, "", "", "").Return([]models.Trip{
		{ID: 1, Title: "Egypt"},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var respBody tripsResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}
	if !respBody.Success {
		t.Fatalf("expected success true, got false")
	}
	if len(respBody.Data) != 1 || respBody.Data[0].Title != "Egypt" {
		t.Fatalf("unexpected response %+v", respBody.Data)
	}

}

func TestTripHandler_Popular(t *testing.T) {
	mockSvc := new(MockTripService)
	log := zaptest.NewLogger(t).Sugar()
	h := handlers.NewTripHandler(mockSvc, log)

	mockSvc.
		On("Popular", mock.Anything, 5).
		Return([]models.Trip{{ID: 1, Title: "Egypt"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips/popular?limit=5", nil)
	w := httptest.NewRecorder()

	h.Popular(w, req)

	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)

	var respBody tripsResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&respBody))

	require.True(t, respBody.Success)
	require.Len(t, respBody.Data, 1)
	require.Equal(t, "Egypt", respBody.Data[0].Title)

	mockSvc.AssertExpectations(t)
}
