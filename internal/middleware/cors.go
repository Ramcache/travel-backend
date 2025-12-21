package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://web95.tech",
			"https://www.web95.tech",
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		ExposedHeaders: []string{
			"Link",
		},
		AllowCredentials: false,
		MaxAge:           300,
	})
}
