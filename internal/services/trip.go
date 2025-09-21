package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type TripService struct {
	repo *repository.TripRepository
	log  *zap.SugaredLogger
}

func NewTripService(repo *repository.TripRepository, log *zap.SugaredLogger) *TripService {
	return &TripService{repo: repo, log: log}
}

func (s *TripService) List(ctx context.Context, city, ttype, season string) ([]models.Trip, error) {
	s.log.Debugw("list_trips", "city", city, "type", ttype, "season", season)
	return s.repo.List(ctx, city, ttype, season)
}

func (s *TripService) Get(ctx context.Context, id int) (*models.Trip, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TripService) Create(ctx context.Context, req models.CreateTripRequest) (*models.Trip, error) {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)

	trip := &models.Trip{
		Title:         req.Title,
		Description:   req.Description,
		PhotoURL:      req.PhotoURL,
		DepartureCity: req.DepartureCity,
		TripType:      req.TripType,
		Season:        req.Season,
		Price:         req.Price,
		Currency:      req.Currency,
		StartDate:     start,
		EndDate:       end,
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, err
	}
	s.log.Infow("trip_created", "id", trip.ID, "title", trip.Title)
	return trip, nil
}

func (s *TripService) Update(ctx context.Context, id int, req models.UpdateTripRequest) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		trip.Title = *req.Title
	}
	if req.Description != nil {
		trip.Description = *req.Description
	}
	if req.PhotoURL != nil {
		trip.PhotoURL = *req.PhotoURL
	}
	if req.DepartureCity != nil {
		trip.DepartureCity = *req.DepartureCity
	}
	if req.TripType != nil {
		trip.TripType = *req.TripType
	}
	if req.Season != nil {
		trip.Season = *req.Season
	}
	if req.Price != nil {
		trip.Price = *req.Price
	}
	if req.Currency != nil {
		trip.Currency = *req.Currency
	}
	if req.StartDate != nil {
		trip.StartDate, _ = time.Parse("2006-01-02", *req.StartDate)
	}
	if req.EndDate != nil {
		trip.EndDate, _ = time.Parse("2006-01-02", *req.EndDate)
	}

	if err := s.repo.Update(ctx, trip); err != nil {
		return nil, err
	}
	s.log.Infow("trip_updated", "id", trip.ID)
	return trip, nil
}

func (s *TripService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.log.Infow("trip_deleted", "id", id)
	return nil
}
