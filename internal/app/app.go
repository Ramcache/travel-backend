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
	UserRepo         *repository.UserRepository
	tripRepo         *repository.TripRepository
	newsRepo         *repository.NewsRepository
	newsCategoryRepo *repository.NewsCategoryRepository

	// services
	AuthService         *services.AuthService
	CurrencyService     *services.CurrencyService
	tripService         *services.TripService
	newsService         *services.NewsService
	newsCategoryService *services.NewsCategoryService

	// handlers
	AuthHandler         *handlers.AuthHandler
	UserHandler         *handlers.UserHandler
	CurrencyHandler     *handlers.CurrencyHandler
	TripHandler         *handlers.TripHandler
	NewsHandler         *handlers.NewsHandler
	ProfileHandler      *handlers.ProfileHandler
	NewsCategoryHandler *handlers.NewsCategoryHandler
}

func New(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, log *zap.SugaredLogger) *App {
	// repositories
	userRepo := repository.NewUserRepository(pool)
	tripRepo := repository.NewTripRepository(pool)
	newsRepo := repository.NewNewsRepository(pool)
	newsCategoryRepo := repository.NewNewsCategoryRepository(pool)

	// services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, log)
	currencyService := services.NewCurrencyService(5*time.Minute, log)
	tripService := services.NewTripService(tripRepo, log)
	newsService := services.NewNewsService(newsRepo, newsCategoryRepo, log)
	newsCategoryService := services.NewNewsCategoryService(newsCategoryRepo, log)

	// handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	userHandler := handlers.NewUserHandler(userRepo, log)
	currencyHandler := handlers.NewCurrencyHandler(currencyService, log)
	tripHandler := handlers.NewTripHandler(tripService, log)
	newsHandler := handlers.NewNewsHandler(newsService, log)
	profileHandler := handlers.NewProfileHandler(authService, log)
	newsCategoryHandler := handlers.NewNewsCategoryHandler(newsCategoryService, log)

	return &App{
		Config:              cfg,
		Pool:                pool,
		Log:                 log,
		UserRepo:            userRepo,
		tripRepo:            tripRepo,
		newsRepo:            newsRepo,
		AuthService:         authService,
		CurrencyService:     currencyService,
		tripService:         tripService,
		newsService:         newsService,
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		CurrencyHandler:     currencyHandler,
		TripHandler:         tripHandler,
		NewsHandler:         newsHandler,
		ProfileHandler:      profileHandler,
		NewsCategoryHandler: newsCategoryHandler,
	}
}
