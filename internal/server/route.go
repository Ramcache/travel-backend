package server

import (
	"net/http"

	"github.com/Ramcache/travel-backend/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	docs "github.com/Ramcache/travel-backend/docs"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/middleware"
)

// RouterDeps содержит все зависимости для построения роутера.
// Так мы избегаем длинной сигнатуры NewRouter.
type RouterDeps struct {
	Auth     *handlers.AuthHandler
	User     *handlers.UserHandler
	Currency *handlers.CurrencyHandler
	Trip     *handlers.TripHandler
	News     *handlers.NewsHandler
	Profile  *handlers.ProfileHandler
	Category *handlers.NewsCategoryHandler
	Stats    *handlers.StatsHandler
	Order    *handlers.OrderHandler
	Feedback *handlers.FeedbackHandler
	Hotel    *handlers.HotelHandler
	Search   *handlers.SearchHandler
	Review   *handlers.ReviewHandler

	JWTSecret string
	Log       *zap.SugaredLogger
	DB        *pgxpool.Pool
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// --- Middlewares ---
	r.Use(middleware.CORS())
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.ZapLogger(deps.Log))
	r.Use(middleware.Recoverer(deps.Log))
	r.Use(middleware.MetricsMiddleware)

	// кастомные 404/405
	r.NotFound(middleware.NotFoundHandler())
	r.MethodNotAllowed(middleware.MethodNotAllowedHandler())

	// --- Routes ---
	registerSystemRoutes(r, deps)
	registerPublicRoutes(r, deps)
	registerProfileRoutes(r, deps)
	registerAdminRoutes(r, deps)

	return r
}

// --- System endpoints (metrics, swagger, health) ---
func registerSystemRoutes(r chi.Router, deps RouterDeps) {
	docs.SwaggerInfo.Title = "Travel API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		if err := storage.Ping(req.Context(), deps.DB); err != nil {
			http.Error(w, "db down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// --- Public API (/api/v1/...) ---
func registerPublicRoutes(r chi.Router, deps RouterDeps) {
	r.Route("/api/v1", func(api chi.Router) {
		// Auth
		api.Post("/auth/register", deps.Auth.Register)
		api.Post("/auth/login", deps.Auth.Login)

		// Currency
		api.Get("/currency", deps.Currency.GetRates)

		// Trips
		api.Get("/trips", deps.Trip.List)
		api.Get("/trips/{id}", deps.Trip.Get)
		api.Get("/trips/{id}/countdown", deps.Trip.Countdown)
		api.Get("/trips/main", deps.Trip.GetMain)
		api.Get("/trips/popular", deps.Trip.Popular)
		api.Post("/trips/{id}/buy", deps.Trip.Buy)
		api.Post("/trips/buy", deps.Trip.BuyWithoutTrip)

		// News
		api.Get("/news", deps.News.PublicList)
		api.Get("/news/{slug_or_id}", deps.News.PublicGet)
		api.Get("/news/recent", deps.News.Recent)
		api.Get("/news/popular", deps.News.Popular)

		// Feedback
		api.Post("/feedback", deps.Feedback.Create)

		// Search
		api.Get("/search", deps.Search.GlobalSearch)

		// Reviews (исправлено: теперь внутри /api/v1)
		api.Route("/trips/{trip_id}/reviews", func(r chi.Router) {
			r.Get("/", deps.Review.ListByTrip)
			r.Post("/", deps.Review.Create)
		})
	})
}

// --- Profile routes (requires JWT) ---
func registerProfileRoutes(r chi.Router, deps RouterDeps) {
	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(pr chi.Router) {
			pr.Use(middleware.JWTAuth(deps.JWTSecret))
			pr.Get("/profile", deps.Profile.Get)
			pr.Put("/profile", deps.Profile.Update)
		})
	})
}

// --- Admin routes (/api/v1/admin/...) ---
func registerAdminRoutes(r chi.Router, deps RouterDeps) {
	r.Route("/api/v1/admin", func(admin chi.Router) {
		admin.Use(middleware.JWTAuth(deps.JWTSecret))
		admin.Use(middleware.RoleAuth(2))

		// Users
		admin.Get("/users", deps.User.List)
		admin.Get("/users/{id}", deps.User.Get)
		admin.Post("/users", deps.User.Create)
		admin.Put("/users/{id}", deps.User.Update)
		admin.Delete("/users/{id}", deps.User.Delete)

		// Trips
		admin.Get("/trips", deps.Trip.List)
		admin.Get("/trips/{id}", deps.Trip.Get)
		admin.Post("/trips", deps.Trip.Create)
		admin.Put("/trips/{id}", deps.Trip.Update)
		admin.Delete("/trips/{id}", deps.Trip.Delete)
		admin.Post("/trips/{id}/hotels", deps.Hotel.AttachHotelToTrip)

		// News
		admin.Get("/news", deps.News.AdminList)
		admin.Post("/news", deps.News.Create)
		admin.Put("/news/{id}", deps.News.Update)
		admin.Delete("/news/{id}", deps.News.Delete)

		// News categories
		admin.Get("/news/categories", deps.Category.List)
		admin.Get("/news/categories/{id}", deps.Category.Get)
		admin.Post("/news/categories", deps.Category.Create)
		admin.Put("/news/categories/{id}", deps.Category.Update)
		admin.Delete("/news/categories/{id}", deps.Category.Delete)

		// Stats
		admin.Get("/stats", deps.Stats.Get)

		// Orders
		admin.Get("/orders", deps.Order.List)
		admin.Post("/orders/{id}/status", deps.Order.UpdateStatus)
		admin.Post("/orders/{id}/read", deps.Order.MarkAsRead)
		admin.Delete("/orders/{id}", deps.Order.Delete)

		// Feedbacks
		admin.Get("/feedbacks", deps.Feedback.List)
		admin.Post("/feedbacks/{id}/read", deps.Feedback.MarkAsRead)
		admin.Delete("/feedbacks/{id}", deps.Feedback.Delete)

		// Hotels
		admin.Get("/hotels", deps.Hotel.List)
		admin.Get("/hotels/{id}", deps.Hotel.Get)
		admin.Post("/hotels", deps.Hotel.Create)
		admin.Put("/hotels/{id}", deps.Hotel.Update)
		admin.Delete("/hotels/{id}", deps.Hotel.Delete)
	})
}
