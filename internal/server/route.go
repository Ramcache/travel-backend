package server

import (
	"github.com/Ramcache/travel-backend/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	docs "github.com/Ramcache/travel-backend/docs"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/middleware"
)

func NewRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler,
	currencyHandler *handlers.CurrencyHandler, tripHandler *handlers.TripHandler,
	newsHandler *handlers.NewsHandler, profileHandler *handlers.ProfileHandler,
	categoryHandler *handlers.NewsCategoryHandler, statsHandler *handlers.StatsHandler,
	jwtSecret string, log *zap.SugaredLogger, db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.CORS())
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.ZapLogger(log))
	r.Use(middleware.Recoverer(log))

	// кастомные 404/405
	r.NotFound(middleware.NotFoundHandler())
	r.MethodNotAllowed(middleware.MethodNotAllowedHandler())

	// swagger
	docs.SwaggerInfo.Title = "Travel API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := storage.Ping(r.Context(), db); err != nil {
			http.Error(w, "db down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204
	})

	// public
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/register", authHandler.Register)
		api.Post("/auth/login", authHandler.Login)

		api.Get("/currency", currencyHandler.GetRates)

		api.Get("/trips", tripHandler.List)
		api.Get("/trips/{id}", tripHandler.Get)
		api.Get("/trips/{id}/countdown", tripHandler.Countdown)
		api.Get("/trips/main", tripHandler.GetMain)

		api.Get("/news", newsHandler.PublicList)
		api.Get("/news/{slug_or_id}", newsHandler.PublicGet)
		api.Get("/news/recent", newsHandler.Recent)
		api.Get("/news/popular", newsHandler.Popular)

		api.Post("/trips/{id}/buy", tripHandler.Buy)
		api.Get("/trips/popular", tripHandler.Popular)

		// profile (требует JWT)
		api.Group(func(pr chi.Router) {
			pr.Use(middleware.JWTAuth(jwtSecret))
			pr.Get("/profile", profileHandler.Get)
			pr.Put("/profile", profileHandler.Update)
		})

		// admin (JWT + роль 2)
		api.Group(func(admin chi.Router) {
			admin.Use(middleware.JWTAuth(jwtSecret))
			admin.Use(middleware.RoleAuth(2))

			admin.Get("/admin/users", userHandler.List)
			admin.Get("/admin/users/{id}", userHandler.Get)
			admin.Post("/admin/users", userHandler.Create)
			admin.Put("/admin/users/{id}", userHandler.Update)
			admin.Delete("/admin/users/{id}", userHandler.Delete)

			admin.Get("/admin/trips", tripHandler.List)
			admin.Get("/admin/trips/{id}", tripHandler.Get)
			admin.Post("/admin/trips", tripHandler.Create)
			admin.Put("/admin/trips/{id}", tripHandler.Update)
			admin.Delete("/admin/trips/{id}", tripHandler.Delete)

			admin.Get("/admin/news", newsHandler.AdminList)
			admin.Post("/admin/news", newsHandler.Create)
			admin.Put("/admin/news/{id}", newsHandler.Update)
			admin.Delete("/admin/news/{id}", newsHandler.Delete)

			admin.Get("/admin/news/categories", categoryHandler.List)
			admin.Get("/admin/news/categories/{id}", categoryHandler.Get)
			admin.Post("/admin/news/categories", categoryHandler.Create)
			admin.Put("/admin/news/categories/{id}", categoryHandler.Update)
			admin.Delete("/admin/news/categories/{id}", categoryHandler.Delete)

			admin.Get("/admin/stats", statsHandler.Get)
		})
	})

	return r
}
