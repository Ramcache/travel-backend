package services_test

import (
	"context"
	"errors"
	"github.com/Ramcache/travel-backend/internal/helpers"
	"testing"
	"time"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// ---- Mock ----
type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}
func (m *MockUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepo) Create(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *MockUserRepo) Update(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *MockUserRepo) UpdatePassword(ctx context.Context, id int, password string) error {
	return m.Called(ctx, id, password).Error(0)
}
func (m *MockUserRepo) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

// ---- Tests ----

func TestAuthService_Register_InvalidInput(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	u, err := svc.Register(context.Background(), models.RegisterRequest{})
	assert.Nil(t, u)
	assert.Error(t, err)
}

func TestAuthService_Register_EmailTaken(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	repo.On("GetByEmail", mock.Anything, "test@mail.com").
		Return(&models.User{ID: 1}, nil)

	u, err := svc.Register(context.Background(), models.RegisterRequest{
		Email: "test@mail.com", Password: "123456",
	})
	assert.Nil(t, u)
	assert.ErrorIs(t, err, services.ErrEmailTaken)
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	repo.On("GetByEmail", mock.Anything, "new@mail.com").
		Return((*models.User)(nil), errors.New("not found"))
	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(nil)

	u, err := svc.Register(context.Background(), models.RegisterRequest{
		Email: "new@mail.com", Password: "123456", FullName: "Test",
	})
	assert.NoError(t, err)
	assert.Equal(t, "new@mail.com", u.Email)
	repo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	repo.On("GetByEmail", mock.Anything, "a@b.com").
		Return((*models.User)(nil), errors.New("not found"))

	token, err := svc.Login(context.Background(), models.LoginRequest{
		Email: "a@b.com", Password: "123",
	})
	assert.Empty(t, token)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	// bcrypt hash от "rightpass"
	hash, _ := helpers.HashPassword("rightpass")
	repo.On("GetByEmail", mock.Anything, "user@mail.com").
		Return(&models.User{ID: 1, Email: "user@mail.com", Password: hash}, nil)

	token, err := svc.Login(context.Background(), models.LoginRequest{
		Email: "user@mail.com", Password: "wrong",
	})
	assert.Empty(t, token)
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	hash, _ := helpers.HashPassword("123456")
	repo.On("GetByEmail", mock.Anything, "ok@mail.com").
		Return(&models.User{ID: 1, Email: "ok@mail.com", Password: hash}, nil)

	token, err := svc.Login(context.Background(), models.LoginRequest{
		Email: "ok@mail.com", Password: "123456",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_UpdateProfile_Success(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	old := &models.User{ID: 1, FullName: "Old"}
	repo.On("GetByID", mock.Anything, 1).Return(old, nil)
	repo.On("Update", mock.Anything, old).Return(nil)

	newName := "New"
	updated, err := svc.UpdateProfile(context.Background(), 1,
		models.UpdateProfileRequest{FullName: &newName})
	assert.NoError(t, err)
	assert.Equal(t, "New", updated.FullName)
}

func TestAuthService_GetByID_NotFound(t *testing.T) {
	repo := new(MockUserRepo)
	svc := services.NewAuthService(repo, "secret", 24*time.Hour, zaptest.NewLogger(t).Sugar())

	repo.On("GetByID", mock.Anything, 99).Return((*models.User)(nil), errors.New("not found"))

	u, err := svc.GetByID(context.Background(), 99)
	assert.Nil(t, u)
	assert.Error(t, err)
}
