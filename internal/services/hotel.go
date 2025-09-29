package services

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type HotelService struct {
	repo repository.HotelRepositoryI
}

func NewHotelService(repo repository.HotelRepositoryI) *HotelService {
	return &HotelService{repo: repo}
}

func (s *HotelService) Create(ctx context.Context, h *models.Hotel) error {
	return s.repo.Create(ctx, h)
}

func (s *HotelService) Get(ctx context.Context, id int) (*models.Hotel, error) {
	return s.repo.Get(ctx, id)
}

func (s *HotelService) List(ctx context.Context) ([]models.Hotel, error) {
	return s.repo.List(ctx)
}

func (s *HotelService) Update(ctx context.Context, h *models.Hotel) error {
	return s.repo.Update(ctx, h)
}

func (s *HotelService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *HotelService) Attach(ctx context.Context, th *models.TripHotel) error {
	return s.repo.Attach(ctx, th)
}

func (s *HotelService) ListByTrip(ctx context.Context, tripID int) ([]models.Hotel, error) {
	return s.repo.ListByTrip(ctx, tripID)
}
