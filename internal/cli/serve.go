package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/app"
	"github.com/Ramcache/travel-backend/internal/config"
	"github.com/Ramcache/travel-backend/internal/server"
	"github.com/Ramcache/travel-backend/internal/storage"
)

func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start API server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Load()

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			sugar := logger.Sugar()

			ctx := context.Background()

			// 🔌 DB
			pool, err := storage.NewPostgres(ctx, storage.PostgresConfig{
				DSN:         cfg.DB.URL,
				MaxConns:    cfg.DB.MaxConns,
				MinConns:    cfg.DB.MinConns,
				ConnTimeout: cfg.DB.ConnTimeout,
				IdleTimeout: cfg.DB.IdleTimeout,
			})
			if err != nil {
				sugar.Fatalw("db connect error", "err", err)
			}
			defer pool.Close()

			// 🏗️ DI-контейнер
			application := app.New(ctx, cfg, pool)

			// 🌐 Router
			r := server.NewRouter(application.AuthHandler, cfg.JWTSecret)
			addr := fmt.Sprintf(":%s", cfg.AppPort)

			sugar.Infow("server started", "addr", addr)
			srv := server.NewHttpServer(r, addr)

			// 🔄 async run
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					sugar.Fatalw("server error", "err", err)
				}
			}()

			// ⏳ graceful shutdown
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			sugar.Infow("server shutting down")
			if err := srv.Shutdown(ctx); err != nil {
				sugar.Errorw("shutdown error", "err", err)
			}
		},
	}
}
