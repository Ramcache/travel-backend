package server

import (
	"net/http"
	"time"

	"github.com/Ramcache/travel-backend/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	docs "github.com/Ramcache/travel-backend/docs"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/middleware"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	currencyHandler *handlers.CurrencyHandler,
	tripHandler *handlers.TripHandler,
	newsHandler *handlers.NewsHandler,
	profileHandler *handlers.ProfileHandler,
	categoryHandler *handlers.NewsCategoryHandler,
	statsHandler *handlers.StatsHandler,
	orderHandler *handlers.OrderHandler,
	feedbackHandler *handlers.FeedbackHandler,
	hotelHandler *handlers.HotelHandler,
	searchHandler *handlers.SearchHandler,
	reviewHandler *handlers.ReviewHandler,
	tripRouteHandler *handlers.TripRouteHandler,
	tripPageHandler *handlers.TripPageHandler,
	dateHandler *handlers.DateHandler,
	mediaHandler *handlers.MediaHandler,
	cloudflareHandler *handlers.CloudflareHandler,
	jwtSecret string,
	log *zap.SugaredLogger,
	db *pgxpool.Pool,
) http.Handler {
	r := chi.NewRouter()

	// ---- Rate limiters (IP-based) ----
	// TTL чтобы map лимитеров не рос бесконечно
	ttl := 10 * time.Minute

	// Global: 10 rps, burst 20 (под фоновые сканеры уже хватает)
	globalLimiter := middleware.NewIPLimiter(rate.Limit(10), 20, ttl)

	// Auth: 1 rps, burst 3 (защита от брутфорса)
	authLimiter := middleware.NewIPLimiter(rate.Limit(1), 3, ttl)

	// Buy/feedback: 0.5 rps (~1 запрос в 2 секунды), burst 2
	buyLimiter := middleware.NewIPLimiter(rate.Limit(0.5), 2, ttl)

	// Admin upload/cleanup: 0.2 rps (~1 запрос в 5 секунд), burst 1
	adminUploadLimiter := middleware.NewIPLimiter(rate.Limit(0.2), 1, ttl)

	// middlewares
	r.Use(middleware.CORS())
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)

	// Важно: rate limit после RealIP, чтобы лимитироваться по реальному IP за nginx
	r.Use(middleware.RateLimit(globalLimiter))

	r.Use(middleware.ZapLogger(log))
	r.Use(middleware.Recoverer(log))
	r.Use(middleware.MetricsMiddleware)

	// кастомные 404/405
	r.NotFound(middleware.NotFoundHandler())
	r.MethodNotAllowed(middleware.MethodNotAllowedHandler())

	// swagger
	docs.SwaggerInfo.Title = "Travel API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	// infra
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := storage.Ping(r.Context(), db); err != nil {
			http.Error(w, "db down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent) // 204
	})
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// public + api
	r.Route("/api/v1", func(api chi.Router) {
		// auth — отдельный tight лимит
		api.Group(func(a chi.Router) {
			a.Use(middleware.RateLimit(authLimiter))
			a.Post("/auth/register", authHandler.Register)
			a.Post("/auth/login", authHandler.Login)
		})

		api.Get("/date/today", dateHandler.Today)
		api.Get("/currency", currencyHandler.GetRates)

		api.Get("/trips", tripHandler.List)
		api.Get("/trips/{id}", tripHandler.Get)
		api.Get("/trips/{id}/countdown", tripHandler.Countdown)
		api.Get("/trips/{id}/page", tripPageHandler.Get)
		api.Get("/trips/main", tripHandler.GetMain)

		api.Get("/news", newsHandler.PublicList)
		api.Get("/news/{slug_or_id}", newsHandler.PublicGet)
		api.Get("/news/recent", newsHandler.Recent)
		api.Get("/news/popular", newsHandler.Popular)

		api.Get("/trips/popular", tripHandler.Popular)
		api.Get("/trips/full", tripPageHandler.ListAll)
		api.Get("/trips/relations", tripPageHandler.ListWithRelations)
		api.Get("/trips/{id}/relations", tripPageHandler.GetWithRelations)

		api.Get("/search", searchHandler.GlobalSearch)

		api.Route("/trips/{trip_id}/reviews", func(rr chi.Router) {
			rr.Get("/", reviewHandler.ListByTrip)
			rr.Post("/", reviewHandler.Create)
		})

		// маршруты тура — публичный список
		api.Get("/trips/{id}/routes", tripRouteHandler.GetTripRoutesCities)
		api.Get("/trips/{id}/routes/ui", tripRouteHandler.ListUI)

		// buy + feedback — отдельный лимит (обычно достаточно жёсткий)
		api.Group(func(b chi.Router) {
			b.Use(middleware.RateLimit(buyLimiter))
			b.Post("/trips/{id}/buy", tripHandler.Buy)
			b.Post("/trips/buy", tripHandler.BuyWithoutTrip)
			b.Post("/feedback", feedbackHandler.Create)
		})

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
			admin.Post("/admin/cloudflare/purge-cache", cloudflareHandler.PurgeCache)

			admin.Get("/admin/users", userHandler.List)
			admin.Get("/admin/users/{id}", userHandler.Get)
			admin.Post("/admin/users", userHandler.Create)
			admin.Put("/admin/users/{id}", userHandler.Update)
			admin.Delete("/admin/users/{id}", userHandler.Delete)

			admin.Get("/admin/trips", tripHandler.List)
			admin.Get("/admin/trips/{id}", tripHandler.Get)
			admin.Post("/admin/trips", tripHandler.Create)
			admin.Post("/admin/tours", tripHandler.CreateTour)
			admin.Put("/admin/trips/{id}", tripHandler.Update)
			admin.Delete("/admin/trips/{id}", tripHandler.Delete)
			admin.Put("/admin/trips/{id}/full", tripHandler.UpdateTour)
			admin.Get("/admin/trips/{id}/full", tripHandler.GetFull)

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

			admin.Get("/admin/orders", orderHandler.List)
			admin.Post("/admin/orders/{id}/status", orderHandler.UpdateStatus)
			admin.Post("/admin/orders/{id}/read", orderHandler.MarkAsRead)
			admin.Delete("/admin/orders/{id}", orderHandler.Delete)

			admin.Get("/admin/feedbacks", feedbackHandler.List)
			admin.Post("/admin/feedbacks/{id}/read", feedbackHandler.MarkAsRead)
			admin.Delete("/admin/feedbacks/{id}", feedbackHandler.Delete)

			// hotels CRUD
			admin.Get("/admin/hotels", hotelHandler.List)
			admin.Get("/admin/hotels/{id}", hotelHandler.Get)
			admin.Post("/admin/hotels", hotelHandler.Create)
			admin.Put("/admin/hotels/{id}", hotelHandler.Update)
			admin.Delete("/admin/hotels/{id}", hotelHandler.Delete)
			admin.Post("/admin/trips/{id}/hotels", hotelHandler.AttachHotelToTrip)

			// routes CRUD
			admin.Post("/admin/trips/{id}/routes/batch", tripRouteHandler.CreateBatch)
			admin.Put("/admin/trips/{id}/routes/{route_id}", tripRouteHandler.Update)
			admin.Delete("/admin/trips/{id}/routes/{route_id}", tripRouteHandler.Delete)

			// upload/cleanup — отдельный строгий лимит
			admin.Group(func(up chi.Router) {
				up.Use(middleware.RateLimit(adminUploadLimiter))
				up.Post("/admin/upload", mediaHandler.Upload)
				up.Post("/admin/media/cleanup", mediaHandler.CleanupUnused)
			})

			// оставшиеся admin media endpoints (как у вас)
			admin.Get("/admin/uploads", mediaHandler.ListUploads)
			admin.Delete("/admin/upload", mediaHandler.DeleteUpload)
		})
	})

	return r
}
