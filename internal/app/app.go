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

	// services
	AuthService     *services.AuthService
	CurrencyService *services.CurrencyService

	// handlers
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	CurrencyHandler *handlers.CurrencyHandler
}

func New(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, log *zap.SugaredLogger) *App {
	// repositories
	userRepo := repository.NewUserRepository(pool)

	// services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, log)
	currencyService := services.NewCurrencyService(5 * time.Minute)

	// handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	userHandler := handlers.NewUserHandler(userRepo, log)
	currencyHandler := handlers.NewCurrencyHandler(currencyService, log)

	return &App{
		Config:          cfg,
		Pool:            pool,
		Log:             log,
		UserRepo:        userRepo,
		AuthService:     authService,
		AuthHandler:     authHandler,
		UserHandler:     userHandler,
		CurrencyHandler: currencyHandler,
	}
}
