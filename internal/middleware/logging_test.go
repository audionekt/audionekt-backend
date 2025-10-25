package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"musicapp/internal/logging"
)

func TestNewLoggingMiddleware(t *testing.T) {
	logger, err := logging.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	middleware := NewLoggingMiddleware(logger)

	if middleware == nil {
		t.Error("Expected middleware to be created, got nil")
	}
}

func TestLoggingMiddleware_Logging(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "successful GET request",
			method:         "GET",
			path:           "/api/users",
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found request",
			method:         "GET",
			path:           "/api/nonexistent",
			statusCode:     http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "POST request",
			method:         "POST",
			path:           "/api/posts",
			statusCode:     http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PUT request",
			method:         "PUT",
			path:           "/api/users/123",
			statusCode:     http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE request",
			method:         "DELETE",
			path:           "/api/posts/456",
			statusCode:     http.StatusNoContent,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "internal server error",
			method:         "GET",
			path:           "/api/error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that sets the status code
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("test response"))
			})

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply logging middleware
			logger, err := logging.NewDevelopment()
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			middleware := NewLoggingMiddleware(logger)
			handler := middleware.Logging(testHandler)
			
			// Record start time to verify logging doesn't take too long
			start := time.Now()
			handler.ServeHTTP(w, req)
			duration := time.Since(start)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify response body
			if w.Body.String() != "test response" {
				t.Errorf("Expected body 'test response', got '%s'", w.Body.String())
			}

			// Verify that logging doesn't take too long (should be very fast)
			if duration > 100*time.Millisecond {
				t.Errorf("Logging middleware took too long: %v", duration)
			}
		})
	}
}

func TestResponseWriter(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "status code 200",
			statusCode: http.StatusOK,
		},
		{
			name:       "status code 404",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "status code 500",
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test response recorder
			w := httptest.NewRecorder()

			// Create response writer wrapper
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default
			}

			// Test WriteHeader
			rw.WriteHeader(tt.statusCode)

			// Verify status code was set
			if rw.statusCode != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, rw.statusCode)
			}

			// Verify underlying response writer also got the status code
			if w.Code != tt.statusCode {
				t.Errorf("Expected underlying status code %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
