package services_test

import (
	"context"
	"testing"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockTripRepo struct{ mock.Mock }

func (m *MockTripRepo) List(ctx context.Context, c, t, s string) ([]models.Trip, error) {
	args := m.Called(ctx, c, t, s)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripRepo) GetByID(ctx context.Context, id int) (*models.Trip, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.Trip), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTripRepo) Create(ctx context.Context, t *models.Trip) error {
	return m.Called(ctx, t).Error(0)
}
func (m *MockTripRepo) Update(ctx context.Context, t *models.Trip) error {
	return m.Called(ctx, t).Error(0)
}
func (m *MockTripRepo) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTripRepo) GetMain(ctx context.Context) (*models.Trip, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*models.Trip), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTripRepo) ResetMain(ctx context.Context, excludeID *int) error {
	return m.Called(ctx, excludeID).Error(0)
}
func (m *MockTripRepo) Popular(ctx context.Context, limit int) ([]models.Trip, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.Trip), args.Error(1)
}
func (m *MockTripRepo) IncrementViews(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTripRepo) IncrementBuys(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func TestTripService_Create_InvalidInput(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	req := models.CreateTripRequest{
		Title: "", DepartureCity: "", TripType: "",
	}
	trip, err := svc.Create(context.Background(), req)
	assert.Nil(t, trip)
	assert.Error(t, err)
}

func TestTripService_Create_InvalidDate(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	req := models.CreateTripRequest{
		Title: "Trip", DepartureCity: "Moscow", TripType: "пляжный",
		StartDate: "bad", EndDate: "2025-07-01",
	}
	trip, err := svc.Create(context.Background(), req)
	assert.Nil(t, trip)
	assert.Error(t, err)
}

func TestTripService_Create_Success(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	req := models.CreateTripRequest{
		Title: "Trip", DepartureCity: "Moscow", TripType: "пляжный",
		StartDate: "2025-07-01", EndDate: "2025-07-05",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Trip")).Return(nil)

	trip, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "Trip", trip.Title)
	mockRepo.AssertExpectations(t)
}

func TestTripService_Get_NotFound(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	mockRepo.On("GetByID", mock.Anything, 99).Return((*models.Trip)(nil), repository.ErrNotFound)

	trip, err := svc.Get(context.Background(), 99)
	assert.Nil(t, trip)
	assert.Error(t, err)
}

func TestTripService_Update_InvalidDate(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	old := &models.Trip{ID: 1, Title: "Old"}
	mockRepo.On("GetByID", mock.Anything, 1).Return(old, nil)

	req := models.UpdateTripRequest{StartDate: ptr("bad-date")}
	trip, err := svc.Update(context.Background(), 1, req)
	assert.Nil(t, trip)
	assert.Error(t, err)
}

func TestTripService_Delete_Success(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	mockRepo.On("Delete", mock.Anything, 1).Return(nil)
	assert.NoError(t, svc.Delete(context.Background(), 1))
	mockRepo.AssertExpectations(t)
}

func TestTripService_IncrementViews_Buys(t *testing.T) {
	mockRepo := new(MockTripRepo)
	svc := services.NewTripService(
		mockRepo,
		nil,
		nil,
		nil,
		"test-frontend",
		zaptest.NewLogger(t).Sugar(),
	)

	mockRepo.On("IncrementViews", mock.Anything, 1).Return(nil)
	mockRepo.On("IncrementBuys", mock.Anything, 1).Return(nil)

	assert.NoError(t, svc.IncrementViews(context.Background(), 1))
	assert.NoError(t, svc.IncrementBuys(context.Background(), 1))
}

func ptr(s string) *string { return &s }
