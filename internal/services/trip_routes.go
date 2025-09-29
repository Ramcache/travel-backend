package services

import (
	"context"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type TripRouteService struct {
	repo repository.TripRouteRepository
}

func NewTripRouteService(repo repository.TripRouteRepository) *TripRouteService {
	return &TripRouteService{repo: repo}
}

func (s *TripRouteService) Create(ctx context.Context, tripID int, req models.TripRouteRequest) (*models.TripRoute, error) {
	return s.repo.Create(ctx, tripID, req)
}

func (s *TripRouteService) Update(ctx context.Context, id int, req models.TripRouteRequest) (*models.TripRoute, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *TripRouteService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *TripRouteService) ListByTrip(ctx context.Context, tripID int) ([]models.TripRoute, error) {
	return s.repo.ListByTrip(ctx, tripID)
}
