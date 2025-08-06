package controllers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		headerName     string
		providedKey    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "authentication disabled (empty api key)",
			apiKey:         "",
			headerName:     "X-API-Key",
			providedKey:    "",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "valid api key",
			apiKey:         "secret123",
			headerName:     "X-API-Key",
			providedKey:    "secret123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "missing api key",
			apiKey:         "secret123",
			headerName:     "X-API-Key",
			providedKey:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"status":"error","msg":"API key required"}`,
		},
		{
			name:           "invalid api key",
			apiKey:         "secret123",
			headerName:     "X-API-Key",
			providedKey:    "wrong-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"status":"error","msg":"Invalid API key"}`,
		},
		{
			name:           "custom header name",
			apiKey:         "secret123",
			headerName:     "Authorization",
			providedKey:    "secret123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Apply authentication middleware
			authMiddleware := APIKeyAuth(tt.apiKey, tt.headerName)
			wrappedHandler := authMiddleware(handler)

			// Create request with logger context
			req := httptest.NewRequest("GET", "/test", nil)
			logger := zerolog.New(io.Discard)
			req = req.WithContext(logger.WithContext(context.Background()))

			// Add API key header if provided
			if tt.providedKey != "" {
				req.Header.Set(tt.headerName, tt.providedKey)
			}

			// Execute request
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			// Check content type for error responses
			if tt.expectedStatus == http.StatusUnauthorized {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				assert.Equal(t, "API-Key", rr.Header().Get("WWW-Authenticate"))
			}
		})
	}
}

func TestConstantTimeEquals(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "secret123",
			b:        "secret123",
			expected: true,
		},
		{
			name:     "different strings same length",
			a:        "secret123",
			b:        "secret456",
			expected: false,
		},
		{
			name:     "different lengths",
			a:        "secret",
			b:        "secret123",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty string",
			a:        "secret",
			b:        "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constantTimeEquals(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPublicEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "ping endpoint",
			path:     "/rest/v1/ping",
			expected: true,
		},
		{
			name:     "health endpoint",
			path:     "/health",
			expected: true,
		},
		{
			name:     "readiness probe",
			path:     "/readiness",
			expected: true,
		},
		{
			name:     "liveness probe",
			path:     "/liveness",
			expected: true,
		},
		{
			name:     "protected scan endpoint",
			path:     "/rest/v1/scan",
			expected: false,
		},
		{
			name:     "protected version endpoint",
			path:     "/rest/v1/version",
			expected: false,
		},
		{
			name:     "other endpoint",
			path:     "/api/data",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPublicEndpoint(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionalAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		apiKey         string
		providedKey    string
		expectedStatus int
	}{
		{
			name:           "public endpoint without api key",
			path:           "/rest/v1/ping",
			apiKey:         "secret123",
			providedKey:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "public endpoint with api key",
			path:           "/rest/v1/ping",
			apiKey:         "secret123",
			providedKey:    "secret123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "protected endpoint with valid api key",
			path:           "/rest/v1/scan",
			apiKey:         "secret123",
			providedKey:    "secret123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "protected endpoint without api key",
			path:           "/rest/v1/scan",
			apiKey:         "secret123",
			providedKey:    "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "protected endpoint with invalid api key",
			path:           "/rest/v1/version",
			apiKey:         "secret123",
			providedKey:    "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Apply conditional authentication middleware
			authMiddleware := ConditionalAPIKeyAuth(tt.apiKey, "X-API-Key")
			wrappedHandler := authMiddleware(handler)

			// Create request with logger context
			req := httptest.NewRequest("GET", tt.path, nil)
			logger := zerolog.New(io.Discard)
			req = req.WithContext(logger.WithContext(context.Background()))

			// Add API key header if provided
			if tt.providedKey != "" {
				req.Header.Set("X-API-Key", tt.providedKey)
			}

			// Execute request
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
