package internalhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/server/api"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/server/handlers"
)

type Server struct {
	logger     *logger.Logger
	app        *app.App
	httpServer *http.Server
}

func NewServer(logger *logger.Logger, app *app.App) *Server {
	return &Server{
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(_ context.Context, host, port string) error {
	strictHandler := handlers.NewStrictCalendarHandler(s.app)
	handler := api.NewStrictHandler(strictHandler, nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(s.loggingMiddleware)

	httpHandler := api.HandlerFromMux(handler, r)

	s.httpServer = &http.Server{ //nolint:gosec
		Addr:    host + ":" + port,
		Handler: httpHandler,
	}

	s.logger.Info("Server starting...")
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping server...")
	return s.httpServer.Shutdown(ctx)
}
