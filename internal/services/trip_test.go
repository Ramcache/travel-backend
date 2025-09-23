package services_test

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// ==== мок репозитория ====
type MockTripRepo struct{ mock.Mock }

func (m *MockTripRepo) List(ctx context.Context, city, ttype, season string) ([]models.Trip, error) {
	args := m.Called(ctx, city, ttype, season)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripRepo) GetByID(ctx context.Context, id int) (*models.Trip, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Trip), args.Error(1)
}
func (m *MockTripRepo) Create(ctx context.Context, t *models.Trip) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}
func (m *MockTripRepo) Update(ctx context.Context, t *models.Trip) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}
func (m *MockTripRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockTripRepo) GetMain(ctx context.Context) (*models.Trip, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.Trip), args.Error(1)
}
func (m *MockTripRepo) ResetMain(ctx context.Context, excludeID *int) error {
	args := m.Called(ctx, excludeID)
	return args.Error(0)
}
func (m *MockTripRepo) Popular(ctx context.Context, limit int) ([]models.Trip, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripRepo) IncrementViews(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockTripRepo) IncrementBuys(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ==== тесты ====

func TestTripService_Get_OK(t *testing.T) {
	mockRepo := new(MockTripRepo)
	service := services.NewTripService(mockRepo, nil)

	exp := &models.Trip{ID: 1, Title: "Egypt"}
	mockRepo.On("GetByID", mock.Anything, 1).Return(exp, nil)

	trip, err := service.Get(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, "Egypt", trip.Title)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Create_InvalidDate(t *testing.T) {
	mockRepo := new(MockTripRepo)
	service := services.NewTripService(mockRepo, nil)

	req := models.CreateTripRequest{
		Title: "Bad", DepartureCity: "Moscow", TripType: "пляжный",
		StartDate: "bad", EndDate: "bad",
	}

	trip, err := service.Create(context.Background(), req)
	assert.Nil(t, trip)
	assert.Error(t, err)
}
