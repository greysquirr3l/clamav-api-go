package controllers

import (
	"net/http"

	"github.com/lescactus/clamav-api-go/internal/clamav"
	"github.com/rs/zerolog"
)

// Handler provides HTTP request handlers for ClamAV API endpoints.
type Handler struct {
	Clamav clamav.Clamaver
	Logger *zerolog.Logger
}

// NewHandler creates a new Handler with the provided logger and ClamAV client.
func NewHandler(logger *zerolog.Logger, clamav clamav.Clamaver) *Handler {
	return &Handler{Logger: logger, Clamav: clamav}
}

// MaxReqSize is a HTTP middleware limiting the size of the request.
// by using http.MaxBytesReader() on the request body.
func MaxReqSize(maxReqSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxReqSize)
			next.ServeHTTP(w, r)
		})
	}
}
