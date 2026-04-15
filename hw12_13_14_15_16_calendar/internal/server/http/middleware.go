package internalhttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start).Seconds()

		// Собираем метрики
		statusCode := wrapper.statusCode
		path := r.URL.Path
		method := r.Method

		// Увеличиваем счетчик запросов
		metrics.RequestsTotal.WithLabelValues(method, path, fmt.Sprint(statusCode)).Inc()

		// Записываем время выполнения
		metrics.RequestDuration.WithLabelValues(method, path).Observe(duration)

		// Если статус >= 400, увеличиваем счетчик ошибок
		if statusCode >= 400 {
			metrics.ErrorsTotal.WithLabelValues(method, path, fmt.Sprint(statusCode)).Inc()
		}

		clientIP := r.RemoteAddr
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			clientIP = ip
		}

		s.logger.Info(fmt.Sprintf(
			"IP: %s Time: %s Method: %s Path: %s Proto: %s Status: %d Latency: %v User-Agent: %s",
			clientIP,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			method,
			path,
			r.Proto,
			statusCode,
			time.Since(start),
			r.UserAgent(),
		))
	})
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) metricsHandler() http.Handler {
	return promhttp.Handler()
}
