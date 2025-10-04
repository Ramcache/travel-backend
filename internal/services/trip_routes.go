package services

import (
	"context"
	"time"

	"github.com/Ramcache/travel-backend/internal/helpers"
	"github.com/Ramcache/travel-backend/internal/models"
	"github.com/Ramcache/travel-backend/internal/repository"
)

type TripRouteService struct {
	repo repository.TripRouteRepository
}

func NewTripRouteService(repo repository.TripRouteRepository) *TripRouteService {
	return &TripRouteService{repo: repo}
}

func (s *TripRouteService) Update(ctx context.Context, id int, req models.TripRouteRequest) (*models.TripRoute, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *TripRouteService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// Старый ответ (совместимость)
func (s *TripRouteService) GetRouteResponse(ctx context.Context, tripID int) (*models.TripRouteResponse, error) {
	routes, err := s.repo.ListByTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}

	segments := make([]models.TripRouteSegment, 0, len(routes))
	var total time.Duration
	for _, rt := range routes {
		segments = append(segments, models.TripRouteSegment{
			City:      rt.City,
			Transport: rt.Transport,
			Duration:  rt.Duration,
			StopTime:  rt.StopTime,
		})
		total += helpers.ParseDurationText(rt.Duration)
		total += helpers.ParseDurationText(rt.StopTime)
	}
	return &models.TripRouteResponse{
		Route:         segments,
		TotalDuration: helpers.FormatDuration(total),
	}, nil
}

// Новый UI-ответ для плашки
func (s *TripRouteService) GetUIRoute(ctx context.Context, tripID int) (*models.TripRouteUIResponse, error) {
	routes, err := s.repo.ListByTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if len(routes) == 0 {
		return &models.TripRouteUIResponse{
			Items:                []models.TripRouteUIItem{},
			TotalDurationText:    "0 минут",
			TotalDurationMinutes: 0,
		}, nil
	}

	items := make([]models.TripRouteUIItem, 0, len(routes)*2-1)
	items = append(items, models.TripRouteUIItem{
		Kind:         "city",
		City:         routes[0].City,
		StopTimeText: routes[0].StopTime,
	})

	var total time.Duration
	// Переход от i-1 города к i-му: транспорт/длительность/стоп из i-го
	for i := 1; i < len(routes); i++ {
		legDur := helpers.ParseDurationText(routes[i].Duration)
		stopDur := helpers.ParseDurationText(routes[i].StopTime)
		total += legDur + stopDur

		items = append(items, models.TripRouteUIItem{
			Kind:         "leg",
			Transport:    routes[i].Transport,
			DurationText: routes[i].Duration,
		})
		items = append(items, models.TripRouteUIItem{
			Kind:         "city",
			City:         routes[i].City,
			StopTimeText: routes[i].StopTime,
		})
	}

	resp := &models.TripRouteUIResponse{
		From:                 routes[0].City,
		To:                   routes[len(routes)-1].City,
		Items:                items,
		TotalDurationText:    helpers.FormatDuration(total),
		TotalDurationMinutes: int(total.Minutes()),
	}
	return resp, nil
}
func (s *TripRouteService) CreateBatch(ctx context.Context, tripID int, reqs []models.TripRouteRequest) ([]models.TripRoute, error) {
	out := make([]models.TripRoute, 0, len(reqs))

	for i, req := range reqs {
		if req.Position == 0 {
			req.Position = i + 1
		}

		rt := &models.TripRoute{
			TripID:    tripID,
			City:      req.City,
			Transport: req.Transport,
			Duration:  req.Duration,
			StopTime:  req.StopTime,
			Position:  req.Position,
		}

		if err := s.repo.Create(ctx, rt); err != nil {
			return nil, err
		}

		out = append(out, *rt)
	}

	return out, nil
}

func (s *TripRouteService) GetCitiesResponse(ctx context.Context, tripID int) (*models.TripRouteCitiesResponse, error) {
	routes, err := s.repo.ListByTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	resp := models.ConvertRoutesToCities(routes)
	return &resp, nil
}
