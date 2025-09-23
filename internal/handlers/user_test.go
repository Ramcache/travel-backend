package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// ---- Mock ----
type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) Create(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *MockUserRepository) Update(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id int, password string) error {
	return m.Called(ctx, id, password).Error(0)
}
func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func TestUserHandler_List(t *testing.T) {
	mockRepo := new(MockUserRepository)
	h := handlers.NewUserHandler(mockRepo, zaptest.NewLogger(t).Sugar())

	mockRepo.On("GetAll", mock.Anything).
		Return([]models.User{{ID: 1, Email: "a@b.com"}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	h := handlers.NewUserHandler(mockRepo, zaptest.NewLogger(t).Sugar())

	mockRepo.On("GetByID", mock.Anything, 5).
		Return((*models.User)(nil), repository.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/5", nil)
	req = withChiURLParam(req, "id", "5")

	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUserHandler_Create_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	h := handlers.NewUserHandler(mockRepo, zaptest.NewLogger(t).Sugar())

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	body := []byte(`{"email":"new@mail.com","password":"123456","full_name":"Test","role_id":1}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_Update_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	h := handlers.NewUserHandler(mockRepo, zaptest.NewLogger(t).Sugar())

	user := &models.User{ID: 1, FullName: "Old"}
	mockRepo.On("GetByID", mock.Anything, 1).Return(user, nil)
	mockRepo.On("Update", mock.Anything, user).Return(nil)

	body := []byte(`{"full_name":"New"}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParam(req, "id", "1")

	w := httptest.NewRecorder()
	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_Delete_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	h := handlers.NewUserHandler(mockRepo, zaptest.NewLogger(t).Sugar())

	mockRepo.On("Delete", mock.Anything, 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	req = withChiURLParam(req, "id", "1")

	w := httptest.NewRecorder()
	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
