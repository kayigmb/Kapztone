package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/server"
)

func TestHealthz(t *testing.T) {
	cases := []struct {
		name     string
		pingErr  error
		wantCode int
		wantBody string
	}{
		{"db up", nil, http.StatusOK, `{"status":"ok"}`},
		{"db down", errors.New("connection refused"), http.StatusServiceUnavailable, `{"status":"db unavailable"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ping := func(context.Context) error { return tc.pingErr }
			rec := httptest.NewRecorder()
			server.NewRouter(ping, &auth.Auth{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantCode)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tc.wantBody {
				t.Fatalf("body = %s, want %s", got, tc.wantBody)
			}
		})
	}
}
