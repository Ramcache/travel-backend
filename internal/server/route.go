package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	docs "github.com/Ramcache/travel-backend/docs"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/middleware"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	currencyHandler *handlers.CurrencyHandler,
	tripHandler *handlers.TripHandler,
	jwtSecret string,
	log *zap.SugaredLogger,
) http.Handler {
	r := chi.NewRouter()

	// middlewares
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(middleware.ZapLogger(log))
	r.Use(middleware.CORS())

	// swagger
	docs.SwaggerInfo.Title = "Travel API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// public
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/register", authHandler.Register)
		api.Post("/auth/login", authHandler.Login)
		api.Get("/currency", currencyHandler.GetRates)
		// public
		api.Get("/trips", tripHandler.List)
		api.Get("/trips/{id}", tripHandler.Get)

		api.Group(func(pr chi.Router) {
			pr.Use(middleware.JWTAuth(jwtSecret))
			pr.Get("/auth/me", authHandler.Me)
		})

		// admin
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
		})
	})

	return r
}
