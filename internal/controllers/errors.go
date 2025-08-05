// Package controllers provides HTTP request handlers for the ClamAV API endpoints.
// It includes handlers for virus scanning, daemon status checks, and management operations.
package controllers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/lescactus/clamav-api-go/internal/clamav"
)

const (
	// StatusKey is the key used for status in JSON responses.
	StatusKey = "status"
	// MsgKey is the key used for messages in JSON responses.
	MsgKey = "msg"
	// StatusError is the status value used for error responses.
	StatusError = "error"
	// ContentTypeApplicationJSON is the content type for JSON responses.
	ContentTypeApplicationJSON = "application/json"
)

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

// NewErrorResponse creates a new error response with the given message.
func NewErrorResponse(msg string) *ErrorResponse {
	return &ErrorResponse{
		Status: "error",
		Msg:    msg,
	}
}

// SetErrorResponse will attempt to parse the given error
// and set the response status code and using the ResponseWriter
// according to the type of the error.
func SetErrorResponse(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var errResp *ErrorResponse

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)

	if isNetError(err) {
		errResp = NewErrorResponse("something wrong happened while communicating with clamav")
		w.WriteHeader(http.StatusBadGateway)
	} else if errors.Is(err, ErrFormFile) || errors.Is(err, ErrOpenFileHeaders) {
		errResp = NewErrorResponse("bad request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		if errors.Is(err, clamav.ErrUnknownCommand) {
			errResp = NewErrorResponse("unknown command sent to clamav")
			w.WriteHeader(http.StatusInternalServerError)
		} else if errors.Is(err, clamav.ErrUnknownResponse) {
			errResp = NewErrorResponse("unknown response from clamav")
			w.WriteHeader(http.StatusInternalServerError)
		} else if errors.Is(err, clamav.ErrUnexpectedResponse) {
			errResp = NewErrorResponse("unexpected response from clamav")
			w.WriteHeader(http.StatusInternalServerError)
		} else if errors.Is(err, clamav.ErrScanFileSizeLimitExceeded) {
			errResp = NewErrorResponse("clamav: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			errResp = NewErrorResponse(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	resp, _ := json.Marshal(errResp)
	_, _ = w.Write(resp)
}

// isNetError returns true if the error is a net.Error
func isNetError(err error) bool {
	var e net.Error
	return errors.As(err, &e)
}
