package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		clientIP := r.RemoteAddr
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			clientIP = ip
		}

		s.logger.Info(fmt.Sprintf(
			"IP: %s Time: %s Method: %s Path: %s Proto: %s Status: %d Latency: %v User-Agent: %s",
			clientIP,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.Path,
			r.Proto,
			wrapper.statusCode,
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
