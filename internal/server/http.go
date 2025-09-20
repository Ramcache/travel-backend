package server

import (
	"context"
	"net/http"
	"time"
)

type HTTPServer struct {
	srv *http.Server
}

func NewHttpServer(handler http.Handler, addr string) *HTTPServer {
	return &HTTPServer{
		srv: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (h *HTTPServer) ListenAndServe() error {
	return h.srv.ListenAndServe()
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}
