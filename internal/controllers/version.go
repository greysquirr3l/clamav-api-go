package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// VersionResponse represents the json response of a /version endpoint.
type VersionResponse struct {
	Version string `json:"clamav_version"`
}

// Version handles requests for ClamAV version information.
func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	// Get request id for logging purposes
	reqID, _ := hlog.IDFromCtx(r.Context())

	ctx := r.Context()

	version, err := h.Clamav.Version(ctx)
	if err != nil {
		h.Logger.Error().Str("req_id", reqID.String()).Msgf("error while sending version command: %v", err)

		SetErrorResponse(w, err)
		return
	}

	h.Logger.Debug().Str("req_id", reqID.String()).Msg("version command sent successfully")

	v := VersionResponse{
		Version: string(version),
	}

	resp, err := json.Marshal(&v)
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
