package services

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

var (
	ErrTripNotFound = errors.New("trip not found")
	ErrInvalidTrip  = errors.New("invalid trip data")
)

type TripService struct {
	repo *repository.TripRepository
	log  *zap.SugaredLogger
}

func NewTripService(repo *repository.TripRepository, log *zap.SugaredLogger) *TripService {
	return &TripService{repo: repo, log: log}
}

// List — список туров
func (s *TripService) List(ctx context.Context, city, ttype, season string) ([]models.Trip, error) {
	s.log.Debugw("list_trips", "city", city, "type", ttype, "season", season)
	return s.repo.List(ctx, city, ttype, season)
}

// Get — получить тур по ID
func (s *TripService) Get(ctx context.Context, id int) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}
	return trip, nil
}

// Create — создать новый тур
func (s *TripService) Create(ctx context.Context, req models.CreateTripRequest) (*models.Trip, error) {
	if req.Title == "" || req.DepartureCity == "" || req.TripType == "" {
		return nil, helpers.ErrInvalidInput("Название тура, город вылета и тип тура обязательны")
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, helpers.ErrInvalidInput("Некорректная дата начала")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, helpers.ErrInvalidInput("Некорректная дата окончания")
	}

	var deadline time.Time
	if req.BookingDeadline != "" {
		deadline, err = time.Parse(time.RFC3339, req.BookingDeadline)
		if err != nil {
			return nil, helpers.ErrInvalidInput("Некорректная дата окончания бронирования")
		}
	}

	trip := &models.Trip{
		Title:           req.Title,
		Description:     req.Description,
		PhotoURL:        req.PhotoURL,
		DepartureCity:   req.DepartureCity,
		TripType:        req.TripType,
		Season:          req.Season,
		Price:           req.Price,
		Currency:        req.Currency,
		StartDate:       start,
		EndDate:         end,
		BookingDeadline: deadline,
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		s.log.Errorw("trip_create_failed", "title", req.Title, "err", err)
		return nil, err
	}

	s.log.Infow("trip_created", "id", trip.ID, "title", trip.Title)
	return trip, nil
}

// Update — обновить тур
func (s *TripService) Update(ctx context.Context, id int, req models.UpdateTripRequest) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, ErrTripNotFound
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
		parsed, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, helpers.ErrInvalidInput("Некорректная дата начала")
		}
		trip.StartDate = parsed
	}
	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, helpers.ErrInvalidInput("Некорректная дата окончания")
		}
		trip.EndDate = parsed
	}
	if req.BookingDeadline != nil {
		parsed, err := time.Parse(time.RFC3339, *req.BookingDeadline)
		if err != nil {
			return nil, helpers.ErrInvalidInput("Некорректная дата окончания бронирования")
		}
		trip.BookingDeadline = parsed
	}

	if err := s.repo.Update(ctx, trip); err != nil {
		s.log.Errorw("trip_update_failed", "id", id, "err", err)
		return nil, err
	}

	s.log.Infow("trip_updated", "id", trip.ID)
	return trip, nil
}

// Delete — удалить тур
func (s *TripService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Errorw("trip_delete_failed", "id", id, "err", err)
		return err
	}

	s.log.Infow("trip_deleted", "id", id)
	return nil
}
