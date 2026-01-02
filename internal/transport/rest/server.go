package rest

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/koan6gi/notifier/internal/config"
	v1 "github.com/koan6gi/notifier/internal/transport/rest/v1"
	"github.com/koan6gi/notifier/pkg/logger"
)

type Server struct {
	srv *http.Server
}

func NewServer(cfg *config.ServerConfig) *Server {
	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return &Server{
		srv: srv,
	}
}

func (s *Server) RegisterHandlers(h *v1.Handlers, lg *logger.Logger, u v1.Updater) {
	mux := http.NewServeMux()

	lastItem := v1.UpdateMiddleware(http.HandlerFunc(h.Last), u)

	mux.Handle("GET /", lastItem)
	mux.HandleFunc("GET /list", h.List)

	s.srv.Handler = v1.LoggingMiddleware(mux, lg)
}

func (s *Server) Run() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
