package middleware

import (
	"net/http"
	"time"

	"musicapp/internal/logging"
)

type LoggingMiddleware struct {
	logger *logging.Logger
}

func NewLoggingMiddleware(logger *logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

func (l *LoggingMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Extract user ID from context if available
		var userID *string
		if id, ok := GetUserIDFromContext(r.Context()); ok {
			userID = &id
		}

		// Log request
		l.logger.LogRequest(r.Method, r.URL.Path, r.UserAgent(), r.RemoteAddr, userID)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log response
		l.logger.LogResponse(r.Method, r.URL.Path, wrapped.statusCode, duration, userID)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
