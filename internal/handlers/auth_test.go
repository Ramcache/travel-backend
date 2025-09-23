package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// ---- Mock ----
type MockAuthService struct{ mock.Mock }

func (m *MockAuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAuthService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}
func (m *MockAuthService) GetByID(context.Context, int) (*models.User, error) { return nil, nil }
func (m *MockAuthService) UpdateProfile(context.Context, int, models.UpdateProfileRequest) (*models.User, error) {
	return nil, nil
}

// ---- Tests ----

func TestAuthHandler_Register_OK(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := handlers.NewAuthHandler(mockSvc, zaptest.NewLogger(t).Sugar())

	user := &models.User{ID: 1, Email: "t@e.com"}
	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("models.RegisterRequest")).
		Return(user, nil)

	body := []byte(`{"email":"t@e.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_Login_OK(t *testing.T) {
	mockSvc := new(MockAuthService)
	h := handlers.NewAuthHandler(mockSvc, zaptest.NewLogger(t).Sugar())

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("models.LoginRequest")).
		Return("token123", nil)

	body := []byte(`{"email":"a@b.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
