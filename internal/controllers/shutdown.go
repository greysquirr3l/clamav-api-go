package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// ShutdownResponse represents the json response of a /shutdown endpoint.
type ShutdownResponse struct {
	Status string `json:"status"`
}

// Shutdown handles requests to gracefully shutdown ClamAV daemon.
func (h *Handler) Shutdown(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	reqID, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	err := h.Clamav.Shutdown(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", reqID.String()).Msgf("error while sending shutdown command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", reqID.String()).Msg("shutdown command sent successfully")

	sr := ShutdownResponse{
		Status: "Shutting down",
	}

	resp, err := json.Marshal(&sr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resp); err != nil {
		h.Logger.Error().Str("req_id", reqID.String()).Msgf("failed to write response: %v", err)
	}
}
