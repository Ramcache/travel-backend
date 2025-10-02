package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	repo          repository.TripRepositoryI
	orderRepo     *repository.OrderRepo
	tripHotelRepo repository.HotelRepositoryI
	routeRepo     repository.TripRouteRepository
	telegram      *helpers.TelegramClient
	frontendURL   string
	log           *zap.SugaredLogger
}

func NewTripService(repo repository.TripRepositoryI, orderRepo *repository.OrderRepo, tripHotelRepo repository.HotelRepositoryI, routeRepo repository.TripRouteRepository, telegram *helpers.TelegramClient, frontendURL string, log *zap.SugaredLogger) *TripService {
	return &TripService{
		repo:          repo,
		orderRepo:     orderRepo,
		tripHotelRepo: tripHotelRepo,
		routeRepo:     routeRepo,
		telegram:      telegram,
		frontendURL:   frontendURL,
		log:           log,
	}
}

type TripServiceI interface {
	List(ctx context.Context, f models.TripFilter) ([]models.Trip, error)
	Get(ctx context.Context, id int) (*models.Trip, error)
	Create(ctx context.Context, req models.CreateTripRequest) (*models.Trip, error)
	Update(ctx context.Context, id int, req models.UpdateTripRequest) (*models.Trip, error)
	Delete(ctx context.Context, id int) error
	GetMain(ctx context.Context) (*models.Trip, error)
	Popular(ctx context.Context, limit int) ([]models.Trip, error)
	IncrementViews(ctx context.Context, id int) error
	IncrementBuys(ctx context.Context, id int) error
	Buy(ctx context.Context, id int, req models.BuyRequest) error
	BuyWithoutTrip(ctx context.Context, req models.BuyRequest) error
	CreateHotel(ctx context.Context, hotel *models.Hotel) error
	CreateRoute(ctx context.Context, tripID int, req models.TripRouteRequest) (*models.TripRoute, error)
}

// List — список туров
func (s *TripService) List(ctx context.Context, f models.TripFilter) ([]models.Trip, error) {
	return s.repo.List(ctx, f)
}

// Get — получить тур по ID
func (s *TripService) Get(ctx context.Context, id int) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}
	return trip, nil
}

func (s *TripService) Create(ctx context.Context, req models.CreateTripRequest) (*models.Trip, error) {
	t := &models.Trip{
		Title:         req.Title,
		Description:   req.Description,
		PhotoURL:      req.PhotoURL,
		DepartureCity: req.DepartureCity,
		TripType:      req.TripType,
		Season:        req.Season,
		Price:         req.Price,
		Currency:      req.Currency,
		Main:          req.Main,
		Active:        req.Active, // 🔹 вот этого не хватало
	}

	// parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}
	var bookingDeadline *time.Time
	if req.BookingDeadline != "" {
		bd, err := time.Parse("2006-01-02", req.BookingDeadline)
		if err != nil {
			return nil, fmt.Errorf("invalid booking_deadline: %w", err)
		}
		bookingDeadline = &bd
	}

	t.StartDate = startDate
	t.EndDate = endDate
	t.BookingDeadline = bookingDeadline

	// создаём тур
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}

	// если передали отели — привяжем
	if len(req.Hotels) > 0 {
		for _, h := range req.Hotels {
			if h.HotelID > 0 {
				th := &models.TripHotel{
					TripID:  t.ID,
					HotelID: h.HotelID,
					Nights:  h.Nights,
				}
				if err := s.tripHotelRepo.Attach(ctx, th); err != nil {
					return nil, err
				}
			}
		}
	}

	return t, nil
}

func (s *TripService) Update(ctx context.Context, id int, req models.UpdateTripRequest) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// обновляем только переданные поля
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
	if req.Main != nil {
		trip.Main = *req.Main
	}
	if req.Active != nil { // 🔹 добавлено
		trip.Active = *req.Active
	}
	if req.StartDate != nil {
		if d, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			trip.StartDate = d
		} else {
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
	}
	if req.EndDate != nil {
		if d, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			trip.EndDate = d
		} else {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
	}
	if req.BookingDeadline != nil {
		if *req.BookingDeadline == "" {
			trip.BookingDeadline = nil
		} else if d, err := time.Parse("2006-01-02", *req.BookingDeadline); err == nil {
			trip.BookingDeadline = &d
		} else {
			return nil, fmt.Errorf("invalid booking_deadline: %w", err)
		}
	}

	if err := s.repo.Update(ctx, trip); err != nil {
		return nil, err
	}

	if req.Hotels != nil {
		if _, err := s.tripHotelRepo.ClearByTrip(ctx, id); err != nil {
			return nil, err
		}
		for _, h := range req.Hotels {
			if h.HotelID > 0 {
				th := &models.TripHotel{TripID: id, HotelID: h.HotelID, Nights: h.Nights}
				if err := s.tripHotelRepo.Attach(ctx, th); err != nil {
					return nil, err
				}
			}
		}
	}

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

func (s *TripService) GetMain(ctx context.Context) (*models.Trip, error) {
	return s.repo.GetMain(ctx)
}

func (s *TripService) Popular(ctx context.Context, limit int) ([]models.Trip, error) {
	return s.repo.Popular(ctx, limit)
}

func (s *TripService) IncrementViews(ctx context.Context, id int) error {
	return s.repo.IncrementViews(ctx, id)
}

func (s *TripService) IncrementBuys(ctx context.Context, id int) error {
	return s.repo.IncrementBuys(ctx, id)
}

func (s *TripService) Buy(ctx context.Context, id int, req models.BuyRequest) error {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTripNotFound
		}
		return err
	}

	var tripID models.NullInt32
	if trip != nil {
		tripID = models.NullInt32{NullInt32: sql.NullInt32{Int32: int32(trip.ID), Valid: true}}
	} else {
		tripID = models.NullInt32{NullInt32: sql.NullInt32{Valid: false}}
	}

	order := models.Order{
		TripID:    tripID,
		UserName:  req.UserName,
		UserPhone: req.UserPhone,
		Status:    "pending",
	}

	if err := s.orderRepo.Create(ctx, &order); err != nil {
		return err
	}

	price := formatPrice(trip.Price)

	msg := fmt.Sprintf(
		"🛒 <b>Новый заказ!</b>\n\n"+
			"📅 <b>Дата:</b> %s\n"+
			"👤 <b>Имя:</b> %s\n"+
			"📞 <b>Телефон:</b> <a href=\"tel:%s\">%s</a>\n\n"+
			"🌍 <b>Тур:</b> %s\n"+
			"💰 <b>Цена:</b> %s руб.",
		time.Now().Format("02.01.2006 15:04"),
		order.UserName,
		order.UserPhone, order.UserPhone,
		trip.Title,
		price,
	)

	//if s.telegram != nil {
	//	if err := s.telegram.SendMessage(msg); err != nil {
	//		s.log.Errorw("Ошибка отправки заказа в Telegram", "order_id", order.ID, "err", err)
	//		return err
	//	}
	//}
	if s.telegram != nil {
		link := fmt.Sprintf("%s/api/v1/trips/%d", strings.TrimRight(s.frontendURL, "/"), trip.ID)

		if err := s.telegram.SendMessageWithButton(msg, "Открыть тур", link); err != nil {
			s.log.Errorw("Ошибка отправки заказа в Telegram", "order_id", order.ID, "err", err)
			return err
		}
	}

	go func() {
		if err := s.repo.IncrementBuys(context.Background(), id); err != nil {
			s.log.Errorw("increment_buys_failed", "id", id, "err", err)
		}
	}()

	return nil
}

// BuyWithoutTrip — заявка без привязки к туру
func (s *TripService) BuyWithoutTrip(ctx context.Context, req models.BuyRequest) error {
	order := models.Order{
		// TripID:    0, // закомментировано
		UserName:  req.UserName,
		UserPhone: req.UserPhone,
		Status:    "pending",
	}

	if err := s.orderRepo.Create(ctx, &order); err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"🛒 <b>Новый заказ!</b>\n\n"+
			"📅 <b>Дата:</b> %s\n"+
			"👤 <b>Имя:</b> %s\n"+
			"📞 <b>Телефон:</b> <a href=\"tel:%s\">%s</a>",
		time.Now().Format("02.01.2006 15:04"),
		order.UserName,
		order.UserPhone, order.UserPhone,
	)

	if s.telegram != nil {
		if err := s.telegram.SendMessage(msg); err != nil {
			s.log.Errorw("Ошибка отправки заказа в Telegram", "order_id", order.ID, "err", err)
			return err
		}
	}

	return nil
}

func formatPrice(price float64) string {
	s := strconv.FormatInt(int64(price), 10)
	n := len(s)
	if n <= 3 {
		return s
	}

	var out strings.Builder
	mod := n % 3
	if mod > 0 {
		out.WriteString(s[:mod])
		if n > mod {
			out.WriteString(" ")
		}
	}

	for i := mod; i < n; i += 3 {
		out.WriteString(s[i : i+3])
		if i+3 < n {
			out.WriteString(" ")
		}
	}
	return out.String()
}

func (s *TripService) CreateHotel(ctx context.Context, hotel *models.Hotel) error {
	return s.tripHotelRepo.Create(ctx, hotel)
}

func (s *TripService) CreateRoute(ctx context.Context, tripID int, req models.TripRouteRequest) (*models.TripRoute, error) {
	return s.routeRepo.Create(ctx, tripID, req)
}
