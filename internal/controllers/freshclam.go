package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// FreshClamResponse represents the json response of a /freshclam endpoint.
type FreshClamResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// FreshClam handles requests to update ClamAV virus definitions.
// This endpoint executes the freshclam command to download the latest
// virus definition updates from ClamAV servers.
func (h *Handler) FreshClam(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	reqID, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	h.Logger.Debug().Str("req_id", reqID.String()).Msg("starting freshclam update")

	output, err := h.Clamav.FreshClam(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", reqID.String()).Msgf("error while running freshclam: %v", err)

		// Return the output even on error, as it may contain useful information
		fcr := FreshClamResponse{
			Status:  "error",
			Message: "freshclam update failed",
			Output:  string(output),
		}

		resp, marshalErr := json.Marshal(&fcr)
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", ContentTypeApplicationJSON)
		w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := w.Write(resp); writeErr != nil {
			h.Logger.Error().Str("req_id", reqID.String()).Msgf("failed to write response: %v", writeErr)
		}
		return
	}

	h.Logger.Info().Str("req_id", reqID.String()).Msg("freshclam update completed successfully")

	fcr := FreshClamResponse{
		Status:  "success",
		Message: "virus definitions updated successfully",
		Output:  string(output),
	}

	resp, err := json.Marshal(&fcr)
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
