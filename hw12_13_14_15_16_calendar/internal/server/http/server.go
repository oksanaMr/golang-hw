package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", helloHandler)

	handler := s.loggingMiddleware(mux)

	s.httpServer = &http.Server{ //nolint:gosec
		Addr:    host + ":" + port,
		Handler: handler,
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

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(50 * time.Millisecond)
	fmt.Fprintf(w, "Hello, World!")
}
