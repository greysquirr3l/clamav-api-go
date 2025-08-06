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

func TestHandlerFreshClam(t *testing.T) {
	logger := zerolog.New(io.Discard)
	mockClamav := &MockClamav{}

	type args struct {
		scenario MockScenario
	}
	type want struct {
		status int
		body   []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "no error",
			args: args{
				scenario: ScenarioNoError,
			},
			want: want{
				status: http.StatusOK,
				body:   []byte(`{"status":"success","message":"virus definitions updated successfully","output":"Database updated successfully"}`),
			},
		},
		{
			name: "error is net error",
			args: args{
				scenario: ScenarioNetError,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","message":"freshclam update failed","output":"network error"}`),
			},
		},
		{
			name: "error is ErrUnknownCommand",
			args: args{
				scenario: ScenarioErrUnknownCommand,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","message":"freshclam update failed","output":"ERROR: Command not found"}`),
			},
		},
		{
			name: "error is ErrUnknownResponse",
			args: args{
				scenario: ScenarioErrUnknownResponse,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","message":"freshclam update failed","output":"ERROR: Unknown response"}`),
			},
		},
		{
			name: "error is ErrUnexpectedResponse",
			args: args{
				scenario: ScenarioErrUnexpectedResponse,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","message":"freshclam update failed","output":"ERROR: Unexpected response"}`),
			},
		},
		{
			name: "error is ErrScanFileSizeLimitExceeded",
			args: args{
				scenario: ScenarioErrScanFileSizeLimitExceeded,
			},
			want: want{
				status: http.StatusInternalServerError,
				body:   []byte(`{"status":"error","message":"freshclam update failed","output":"ERROR: Size limit exceeded"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&logger, mockClamav)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.FreshClam)

			ctx := context.WithValue(context.Background(), MockScenario(""), tt.args.scenario)
			req, err := http.NewRequestWithContext(ctx, "POST", "/rest/v1/freshclam", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler.ServeHTTP(rr, req)

			resp := rr.Result()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.want.status, resp.StatusCode)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.body, body)
		})
	}
}
