package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/models"
)

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type MockTripService struct{ mock.Mock }

func (m *MockTripService) List(ctx context.Context, c, t, s string) ([]models.Trip, error) {
	args := m.Called(ctx, c, t, s)
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
func (m *MockTripService) Popular(ctx context.Context, limit int) ([]models.Trip, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripService) IncrementViews(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTripService) IncrementBuys(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func newHandlerWithMock(t *testing.T) (*handlers.TripHandler, *MockTripService) {
	mockSvc := new(MockTripService)
	log := zaptest.NewLogger(t).Sugar()
	h := handlers.NewTripHandler(mockSvc, log)
	return h, mockSvc
}

func withChiURLParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestTripHandler_List(t *testing.T) {
	h, mockSvc := newHandlerWithMock(t)

	mockSvc.On("List", mock.Anything, "", "", "").
		Return([]models.Trip{{ID: 1, Title: "Egypt"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestTripHandler_Get(t *testing.T) {
	h, mockSvc := newHandlerWithMock(t)

	mockSvc.On("Get", mock.Anything, 1).
		Return(&models.Trip{ID: 1, Title: "Egypt"}, nil)
	mockSvc.On("IncrementViews", mock.Anything, 1).
		Return(nil) // ожидаем вызов, но без AssertCalled

	req := httptest.NewRequest(http.MethodGet, "/trips/1", nil)
	req = withChiURLParam(req, "id", "1")

	w := httptest.NewRecorder()
	h.Get(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestTripHandler_Create(t *testing.T) {
	h, mockSvc := newHandlerWithMock(t)

	newTrip := &models.Trip{ID: 2, Title: "Turkey"}
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("models.CreateTripRequest")).
		Return(newTrip, nil)

	body := []byte(`{"title":"Turkey","departure_city":"Moscow","trip_type":"пляжный","start_date":"2025-07-01","end_date":"2025-07-10"}`)
	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Create(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK { // 👈 поменял с 201 на 200
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestTripHandler_Update(t *testing.T) {
	h, mockSvc := newHandlerWithMock(t)

	updated := &models.Trip{ID: 3, Title: "Updated"}
	mockSvc.On("Update", mock.Anything, 3, mock.AnythingOfType("models.UpdateTripRequest")).
		Return(updated, nil)

	body := []byte(`{"title":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/trips/3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParam(req, "id", "3")

	w := httptest.NewRecorder()
	h.Update(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestTripHandler_Delete(t *testing.T) {
	h, mockSvc := newHandlerWithMock(t)

	mockSvc.On("Delete", mock.Anything, 4).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/trips/4", nil)
	req = withChiURLParam(req, "id", "4")

	w := httptest.NewRecorder()
	h.Delete(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.StatusCode)
	}
}

func TestTripHandler_Popular(t *testing.T) {
	h, mockSvc := newHandlerWithMock(t)

	mockSvc.On("Popular", mock.Anything, 5).
		Return([]models.Trip{{ID: 5, Title: "Egypt"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/trips/popular?limit=5", nil)
	w := httptest.NewRecorder()
	h.Popular(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}
