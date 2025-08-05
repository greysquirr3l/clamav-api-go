package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// ReloadResponse represents the json response of a /reload endpoint.
type ReloadResponse struct {
	Status string `json:"status"`
}

// Reload handles requests to reload ClamAV configuration and databases.
func (h *Handler) Reload(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	reqID, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	err := h.Clamav.Reload(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", reqID.String()).Msgf("error while sending reload command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", reqID.String()).Msg("reload command sent successfully")

	rr := ReloadResponse{
		Status: "Reloading",
	}

	resp, err := json.Marshal(&rr)
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
