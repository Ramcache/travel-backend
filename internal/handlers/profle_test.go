package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// ---- Mock ----
type MockProfileService struct{ mock.Mock }

func (m *MockProfileService) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockProfileService) UpdateProfile(ctx context.Context, id int, req models.UpdateProfileRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// заглушки для интерфейса AuthService
func (m *MockProfileService) Register(context.Context, models.RegisterRequest) (*models.User, error) {
	return nil, nil
}
func (m *MockProfileService) Login(context.Context, models.LoginRequest) (string, error) {
	return "", nil
}

func TestProfileHandler_Get_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	h := handlers.NewProfileHandler(mockSvc, zaptest.NewLogger(t).Sugar())

	mockSvc.On("GetByID", mock.Anything, 1).
		Return(&models.User{ID: 1, Email: "a@b.com"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req = req.WithContext(helpers.SetUserID(context.Background(), 1))

	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProfileHandler_Get_NotFound(t *testing.T) {
	mockSvc := new(MockProfileService)
	h := handlers.NewProfileHandler(mockSvc, zaptest.NewLogger(t).Sugar())

	mockSvc.On("GetByID", mock.Anything, 2).
		Return((*models.User)(nil), services.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req = req.WithContext(helpers.SetUserID(context.Background(), 2))

	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileHandler_Update_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	h := handlers.NewProfileHandler(mockSvc, zaptest.NewLogger(t).Sugar())

	updated := &models.User{ID: 1, FullName: "New"}
	mockSvc.On("UpdateProfile", mock.Anything, 1, mock.AnythingOfType("models.UpdateProfileRequest")).
		Return(updated, nil)

	body := []byte(`{"full_name":"New"}`)
	req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(helpers.SetUserID(context.Background(), 1))

	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
