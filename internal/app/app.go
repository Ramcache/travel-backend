package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/repository"
	"github.com/Ramcache/travel-backend/internal/services"
)

type App struct {
	Config *config.Config
	Pool   *pgxpool.Pool

	// repositories
	UserRepo *repository.UserRepository

	// services
	AuthService *services.AuthService

	// handlers
	AuthHandler *handlers.AuthHandler
}

func New(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool) *App {
	// repositories
	userRepo := repository.NewUserRepository(pool)

	// services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)

	// handlers
	authHandler := handlers.NewAuthHandler(authService)

	return &App{
		Config:      cfg,
		Pool:        pool,
		UserRepo:    userRepo,
		AuthService: authService,
		AuthHandler: authHandler,
	}
}
