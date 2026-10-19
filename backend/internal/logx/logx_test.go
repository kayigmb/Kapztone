package logx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

// One access line per request, carrying the request id, the route pattern,
// who acted, and the error a handler answered 500 for.
func TestAccessLine(t *testing.T) {
	var buf bytes.Buffer
	if _, err := logx.Setup(&buf, logx.FormatJSON, "info"); err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	r.Use(logx.Middleware(func(*http.Request) string { return "192.0.2.1" }))
	r.Get("/things/{id}", func(w http.ResponseWriter, r *http.Request) {
		logx.SetUser(r.Context(), "u-1")
		logx.SetOrg(r.Context(), "acme")
		logx.RecordError(w, "load thing", errors.New("db down"))
		w.WriteHeader(http.StatusInternalServerError)
	})
	r.Get("/boom", func(http.ResponseWriter, *http.Request) { panic("boom") })

	req := httptest.NewRequest(http.MethodGet, "/things/42", nil)
	req.Header.Set(logx.HeaderRequestID, "web-request-0001")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Header().Get(logx.HeaderRequestID) != "web-request-0001" {
		t.Fatal("the request id from the web app is not echoed")
	}
	var line map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &line); err != nil {
		t.Fatalf("not one JSON line: %s", buf.String())
	}
	for k, want := range map[string]any{
		logx.KeyRequestID: "web-request-0001", logx.KeyRoute: "/things/{id}", logx.KeyStatus: float64(500),
		logx.KeyUser: "u-1", logx.KeyOrg: "acme", logx.KeyError: "db down", logx.KeyErrorAt: "load thing", "level": "ERROR",
	} {
		if line[k] != want {
			t.Errorf("%s = %v, want %v", k, line[k], want)
		}
	}

	buf.Reset()
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))
	if rec.Code != http.StatusInternalServerError || !bytes.Contains(buf.Bytes(), []byte(`"panic":"boom"`)) ||
		len(rec.Header().Get(logx.HeaderRequestID)) < 8 {
		t.Fatalf("panic not recovered and logged with a fresh request id: %d %s", rec.Code, buf.String())
	}
	if _, err := logx.Setup(&buf, "xml", "info"); err == nil {
		t.Fatal("an unknown format is accepted")
	}
}
