package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/services"
)

type App struct {
	Config *config.Config
	Pool   *pgxpool.Pool
	Log    *zap.SugaredLogger

	// repositories
	UserRepo *repository.UserRepository
	tripRepo *repository.TripRepository
	newsRepo *repository.NewsRepository

	// services
	AuthService     *services.AuthService
	CurrencyService *services.CurrencyService
	tripService     *services.TripService
	newsService     *services.NewsService

	// handlers
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	CurrencyHandler *handlers.CurrencyHandler
	TripHandler     *handlers.TripHandler
	NewsHandler     *handlers.NewsHandler
	ProfileHandler  *handlers.ProfileHandler
}

func New(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, log *zap.SugaredLogger) *App {
	// repositories
	userRepo := repository.NewUserRepository(pool)
	tripRepo := repository.NewTripRepository(pool)
	newsRepo := repository.NewNewsRepository(pool)

	// services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, log)
	currencyService := services.NewCurrencyService(5*time.Minute, log) // 🟢 добавил log
	tripService := services.NewTripService(tripRepo, log)
	newsService := services.NewNewsService(newsRepo, log)

	// handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	userHandler := handlers.NewUserHandler(userRepo, log)
	currencyHandler := handlers.NewCurrencyHandler(currencyService, log)
	tripHandler := handlers.NewTripHandler(tripService, log) // 🟢 добавил log
	newsHandler := handlers.NewNewsHandler(newsService, log)
	profileHandler := handlers.NewProfileHandler(authService, log) // 🟢 добавил log

	return &App{
		Config:          cfg,
		Pool:            pool,
		Log:             log,
		UserRepo:        userRepo,
		tripRepo:        tripRepo,
		newsRepo:        newsRepo,
		AuthService:     authService,
		CurrencyService: currencyService,
		tripService:     tripService,
		newsService:     newsService,
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		CurrencyHandler: currencyHandler,
		TripHandler:     tripHandler,
		NewsHandler:     newsHandler,
		ProfileHandler:  profileHandler,
	}
}
