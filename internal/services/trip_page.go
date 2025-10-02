package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/models"
)

// TripPageService агрегирует данные из уже существующих сервисов проекта
type TripPageService struct {
	trips    *TripService
	hotels   *HotelService
	reviews  *ReviewService
	news     *NewsService
	routes   *TripRouteService
	currency *CurrencyService
	log      *zap.SugaredLogger
}

func NewTripPageService(
	trips *TripService,
	hotels *HotelService,
	reviews *ReviewService,
	news *NewsService,
	routes *TripRouteService,
	currency *CurrencyService,
	log *zap.SugaredLogger,
) *TripPageService {
	return &TripPageService{
		trips:    trips,
		hotels:   hotels,
		reviews:  reviews,
		news:     news,
		routes:   routes,
		currency: currency,
		log:      log,
	}
}

// Get собирает полный набор данных для страницы тура
func (s *TripPageService) Get(ctx context.Context, id int) (*models.TripPageResponse, error) {
	trip, err := s.trips.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Routes
	routes, err := s.routes.GetRouteResponse(ctx, id)
	if err != nil {
		s.log.Errorw("trip_page_routes_failed", "trip_id", id, "err", err)
		routes = nil
	}

	// Hotels
	hotels, err := s.hotels.ListByTrip(ctx, id)
	if err != nil {
		s.log.Errorw("trip_page_hotels_failed", "trip_id", id, "err", err)
		hotels = nil
	}
	// Options
	options, err := s.trips.repo.GetOptions(ctx, id)
	if err != nil {
		s.log.Errorw("trip_page_options_failed", "trip_id", id, "err", err)
		options = nil
	}

	// Reviews (берём последние 3 для первого экрана)
	reviewItems, total, err := s.reviews.ListByTrip(ctx, id, 3, 0)
	if err != nil {
		s.log.Errorw("trip_page_reviews_failed", "trip_id", id, "err", err)
		reviewItems, total = nil, 0
	}

	// Popular trips (например, 6 шт.)
	popular, err := s.trips.Popular(ctx, 6)
	if err != nil {
		s.log.Errorw("trip_page_popular_failed", "trip_id", id, "err", err)
		popular = nil
	}

	newsItems, _, err := s.news.PublicList(ctx, 6, 0)
	if err != nil {
		s.log.Errorw("trip_page_news_failed", "trip_id", id, "err", err)
		newsItems = nil
	}

	// Currency (кешируется в CurrencyService)
	rates, err := s.currency.GetRates(ctx)
	if err != nil {
		s.log.Errorw("trip_page_currency_failed", "trip_id", id, "err", err)
	}

	// Countdown
	var cd *models.Countdown
	if trip.BookingDeadline != nil {
		now := time.Now()
		diff := trip.BookingDeadline.Sub(now)
		if diff > 0 {
			cd = &models.Countdown{
				Days:    int(diff.Hours()) / 24,
				Hours:   int(diff.Hours()) % 24,
				Minutes: int(diff.Minutes()) % 60,
				Seconds: int(diff.Seconds()) % 60,
			}
		} else {
			cd = &models.Countdown{Days: 0, Hours: 0, Minutes: 0, Seconds: 0}
		}
	}

	resp := &models.TripPageResponse{
		Trip:         *trip,
		Countdown:    cd,
		DurationDays: models.CalcDurationDays(trip.StartDate, trip.EndDate),
		Routes:       routes,
		Hotels:       models.ToHotelResponses(hotels),
		Options:      options,
		Reviews: models.TripPageReviews{
			Total: total,
			Items: reviewItems,
		},
		PopularTrips: popular,
		News:         newsItems,
		CurrencyRates: models.CurrencyRatesPayload{
			USD: rates.USD,
			SAR: rates.SAR,
		},
	}
	return resp, nil
}

func (s *TripPageService) ListAll(ctx context.Context) ([]models.TripPageResponse, error) {
	trips, err := s.trips.List(ctx, "", "", "")
	if err != nil {
		return nil, err
	}

	var results []models.TripPageResponse
	for _, t := range trips {
		data, err := s.Get(ctx, t.ID) // t.ID теперь доступен
		if err != nil {
			return nil, err
		}
		results = append(results, *data)
	}
	return results, nil
}

func (s *TripPageService) GetWithRelations(ctx context.Context, id int) (*models.TripWithRelations, error) {
	trip, err := s.trips.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	hotels, err := s.hotels.ListByTrip(ctx, id)
	if err != nil {
		s.log.Errorw("trip_relations_hotels_failed", "trip_id", id, "err", err)
		hotels = nil
	}

	routes, err := s.routes.GetCitiesResponse(ctx, id)
	if err != nil {
		s.log.Errorw("trip_relations_routes_failed", "trip_id", id, "err", err)
		routes = &models.TripRouteCitiesResponse{Cities: map[string]models.TripRouteCity{}}
	}

	return &models.TripWithRelations{
		Trip:   *trip,
		Hotels: models.ToHotelResponses(hotels),
		Routes: routes,
	}, nil
}

func (s *TripPageService) ListWithRelations(ctx context.Context) ([]models.TripWithRelations, error) {
	trips, err := s.trips.List(ctx, "", "", "")
	if err != nil {
		return nil, err
	}

	var results []models.TripWithRelations
	for _, t := range trips {
		// Отели
		hotels, err := s.hotels.ListByTrip(ctx, t.ID)
		if err != nil {
			s.log.Errorw("trip_relations_hotels_failed", "trip_id", t.ID, "err", err)
			hotels = nil
		}

		// Маршрут
		routes, err := s.routes.GetCitiesResponse(ctx, t.ID)
		if err != nil {
			s.log.Errorw("trip_relations_routes_failed", "trip_id", t.ID, "err", err)
			routes = &models.TripRouteCitiesResponse{Cities: map[string]models.TripRouteCity{}}
		}

		results = append(results, models.TripWithRelations{
			Trip:   t,
			Hotels: models.ToHotelResponses(hotels),
			Routes: routes,
		})
	}

	return results, nil
}
