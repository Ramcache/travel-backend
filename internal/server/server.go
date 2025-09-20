package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/Ramcache/travel-backend/internal/handlers"
	"github.com/Ramcache/travel-backend/internal/middleware"
)

func NewRouter(authHandler *handlers.AuthHandler, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	// public
	r.Post("/api/v1/auth/register", authHandler.Register)
	r.Post("/api/v1/auth/login", authHandler.Login)

	// protected
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.JWTAuth(jwtSecret))
		pr.Get("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("You are authorized"))
		})
	})

	return r
}
