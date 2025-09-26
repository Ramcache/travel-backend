package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Ramcache/travel-backend/internal/helpers"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func ZapLogger(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sww := &statusWriter{ResponseWriter: w}
			reqID := chimw.GetReqID(r.Context())

			next.ServeHTTP(sww, r)

			var userID interface{} = nil
			if v := r.Context().Value(helpers.UserIDKey); v != nil {
				userID = v
			}

			log.Infow("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sww.status,
				"size", sww.size,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"request_id", reqID,
				"user_id", userID,
			)
		})
	}
}

func Recoverer(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Errorf("panic: %v\n%s", rec, debug.Stack())
					helpers.Error(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helpers.Error(w, http.StatusNotFound, "Ресурс не найден")
	}
}

func MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helpers.Error(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}
