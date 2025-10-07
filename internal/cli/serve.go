package cli

import (
	"context"
	"fmt"
	"github.com/Ramcache/travel-backend/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ramcache/travel-backend/internal/app"
	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/server"
	"github.com/Ramcache/travel-backend/internal/storage"
	"github.com/spf13/cobra"
)

func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start API server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()

			log := logger.New(cfg.AppEnv)
			defer log.Sync()

			ctx := context.Background()

			pool, err := storage.NewPostgres(ctx, storage.PostgresConfig{
				DSN:         cfg.DB.URL,
				MaxConns:    cfg.DB.MaxConns,
				MinConns:    cfg.DB.MinConns,
				ConnTimeout: cfg.DB.ConnTimeout,
				IdleTimeout: cfg.DB.IdleTimeout,
			})
			if err != nil {
				log.Fatalw("db connect error", "err", err)
			}
			defer pool.Close()

			application := app.New(ctx, cfg, pool, log)
			r := server.NewRouter(application.AuthHandler, application.UserHandler,
				application.CurrencyHandler, application.TripHandler, application.NewsHandler,
				application.ProfileHandler, application.NewsCategoryHandler, application.StatsHandler,
				application.OrderHandler, application.FeedbackHandler, application.HotelHandler, application.SearchHandler,
				application.ReviewsHandler, application.TripRouteHandler, application.TripPageHandler,
				application.DateHandler, application.MediaHandler, cfg.JWTSecret, log, pool)

			addr := fmt.Sprintf(":%s", cfg.AppPort)

			log.Infow("server started", "addr", addr)
			srv := server.NewHttpServer(r, addr)

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalw("server error", "err", err)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			log.Infow("server shutting down")
			if err := srv.Shutdown(ctx); err != nil {
				log.Errorw("shutdown error", "err", err)
			}
		},
	}
}
