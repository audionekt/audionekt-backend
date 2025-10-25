package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedHeaders map[string]string
	}{
		{
			name:   "GET request with CORS headers",
			method: "GET",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
		{
			name:   "POST request with CORS headers",
			method: "POST",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
		{
			name:   "OPTIONS request (preflight)",
			method: "OPTIONS",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
		{
			name:   "PUT request with CORS headers",
			method: "PUT",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
		{
			name:   "DELETE request with CORS headers",
			method: "DELETE",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test response"))
			})

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply CORS middleware
			handler := CORS(testHandler)
			handler.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify CORS headers
			for headerName, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(headerName)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s: %s, got %s", headerName, expectedValue, actualValue)
				}
			}

			// For OPTIONS requests, verify that the handler wasn't called
			if tt.method == "OPTIONS" {
				if w.Body.String() != "" {
					t.Error("Expected empty body for OPTIONS request")
				}
			} else {
				if w.Body.String() != "test response" {
					t.Errorf("Expected body 'test response', got '%s'", w.Body.String())
				}
			}
		})
	}
}
