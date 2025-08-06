// Package controllers provides authentication middleware for the ClamAV API.
package controllers

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/hlog"
)

// APIKeyAuth returns a middleware that validates API key authentication.
// If apiKey is empty, authentication is disabled and all requests are allowed.
// If apiKey is provided, requests must include the API key in the specified header.
func APIKeyAuth(apiKey, headerName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no API key is configured, authentication is disabled
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Get request ID for logging
			reqID, _ := hlog.IDFromCtx(r.Context())
			logger := hlog.FromRequest(r)

			// Extract API key from header
			providedKey := r.Header.Get(headerName)
			if providedKey == "" {
				logger.Warn().Str("req_id", reqID.String()).
					Str("expected_header", headerName).
					Msg("API key authentication required but no key provided")

				writeAPIKeyErrorResponse(w, "API key required")
				return
			}

			// Validate API key (constant-time comparison for security)
			if !constantTimeEquals(providedKey, apiKey) {
				logger.Warn().Str("req_id", reqID.String()).
					Str("client_ip", r.RemoteAddr).
					Str("user_agent", r.UserAgent()).
					Msg("Invalid API key provided")

				writeAPIKeyErrorResponse(w, "Invalid API key")
				return
			}

			// API key is valid, log successful authentication
			logger.Debug().Str("req_id", reqID.String()).
				Msg("API key authentication successful")

			next.ServeHTTP(w, r)
		})
	}
}

// constantTimeEquals performs a constant-time comparison of two strings
// to prevent timing attacks on API key validation.
func constantTimeEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// writeAPIKeyErrorResponse writes a standardized error response for authentication failures.
func writeAPIKeyErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Header().Set("WWW-Authenticate", "API-Key")
	w.WriteHeader(http.StatusUnauthorized)

	response := []byte(`{"status":"error","msg":"` + message + `"}`)
	if _, err := w.Write(response); err != nil {
		// Log error but don't expose it to client
		return
	}
}

// IsPublicEndpoint returns true if the endpoint should be accessible without authentication.
// Typically health check endpoints should remain public.
func IsPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/rest/v1/ping", // Health check should remain accessible
		"/health",       // Alternative health check endpoint
		"/readiness",    // Kubernetes readiness probe
		"/liveness",     // Kubernetes liveness probe
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// ConditionalAPIKeyAuth returns a middleware that applies API key authentication
// only to non-public endpoints. Public endpoints (like health checks) bypass authentication.
func ConditionalAPIKeyAuth(apiKey, headerName string) func(next http.Handler) http.Handler {
	authMiddleware := APIKeyAuth(apiKey, headerName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a public endpoint
			if IsPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Apply authentication for protected endpoints
			authMiddleware(next).ServeHTTP(w, r)
		})
	}
}
