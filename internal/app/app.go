package app

import (
	"context"
	"github.com/Ramcache/travel-backend/internal/helpers"
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
	statsRepo        *repository.StatsRepository
	orderRepo        *repository.OrderRepo

	// services
	AuthService         *services.AuthService
	CurrencyService     *services.CurrencyService
	TripService         *services.TripService
	newsService         *services.NewsService
	newsCategoryService *services.NewsCategoryService
	statsService        *services.StatsService
	orderService        *services.OrderService

	// handlers
	AuthHandler         *handlers.AuthHandler
	UserHandler         *handlers.UserHandler
	CurrencyHandler     *handlers.CurrencyHandler
	TripHandler         *handlers.TripHandler
	NewsHandler         *handlers.NewsHandler
	ProfileHandler      *handlers.ProfileHandler
	NewsCategoryHandler *handlers.NewsCategoryHandler
	StatsHandler        *handlers.StatsHandler
	OrderHandler        *handlers.OrderHandler
}

func New(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, log *zap.SugaredLogger) *App {
	// repositories
	userRepo := repository.NewUserRepository(pool)
	tripRepo := repository.NewTripRepository(pool)
	newsRepo := repository.NewNewsRepository(pool)
	newsCategoryRepo := repository.NewNewsCategoryRepository(pool)
	statsRepo := repository.NewStatsRepository(pool)
	orderRepo := repository.NewOrderRepo(pool)
	// helpers
	telegramClient := helpers.NewTelegramClient(cfg.TG.TelegramToken, cfg.TG.TelegramChat)

	// services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, log)
	currencyService := services.NewCurrencyService(5*time.Minute, log)
	tripService := services.NewTripService(tripRepo, telegramClient, cfg.FrontendURL, log)
	newsService := services.NewNewsService(newsRepo, newsCategoryRepo, log)
	newsCategoryService := services.NewNewsCategoryService(newsCategoryRepo, log)
	statsService := services.NewStatsService(statsRepo)
	orderService := services.NewOrderService(orderRepo)
	// handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	userHandler := handlers.NewUserHandler(userRepo, log)
	currencyHandler := handlers.NewCurrencyHandler(currencyService, log)
	tripHandler := handlers.NewTripHandler(tripService, orderService, log)
	newsHandler := handlers.NewNewsHandler(newsService, log)
	profileHandler := handlers.NewProfileHandler(authService, log)
	newsCategoryHandler := handlers.NewNewsCategoryHandler(newsCategoryService, log)
	statsHandler := handlers.NewStatsHandler(statsService, log)
	orderHandler := handlers.NewOrderHandler(orderService, log)

	return &App{
		Config:              cfg,
		Pool:                pool,
		Log:                 log,
		UserRepo:            userRepo,
		tripRepo:            tripRepo,
		newsRepo:            newsRepo,
		newsCategoryRepo:    newsCategoryRepo,
		statsRepo:           statsRepo,
		AuthService:         authService,
		CurrencyService:     currencyService,
		TripService:         tripService,
		newsService:         newsService,
		newsCategoryService: newsCategoryService,
		statsService:        statsService,
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		CurrencyHandler:     currencyHandler,
		TripHandler:         tripHandler,
		NewsHandler:         newsHandler,
		ProfileHandler:      profileHandler,
		NewsCategoryHandler: newsCategoryHandler,
		StatsHandler:        statsHandler,
		OrderHandler:        orderHandler,
	}
}
