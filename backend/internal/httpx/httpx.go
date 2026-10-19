// Package httpx holds the HTTP helpers shared by handlers and middleware.
package httpx

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

// Messages shown to callers.
const (
	MsgServerError = "something went wrong"
	MsgNotFound    = "not found"
	MsgBadJSON     = "request body must be valid JSON"
)

// ParamID is the URL parameter holding a row id, as in /things/{id}.
const ParamID = "id"

const maxJSONBody = 64 << 10

// Envelope is every JSON answer: the HTTP status again, a sentence the app
// shows as it is (empty for plain reads), and the data.
type Envelope struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func write(w http.ResponseWriter, e Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	json.NewEncoder(w).Encode(e)
}

// OK answers with data and, for a change, the sentence that confirms it.
func OK(w http.ResponseWriter, status int, message string, data any) {
	write(w, Envelope{Status: status, Message: message, Data: data})
}

// Error refuses with a message safe to show.
func Error(w http.ResponseWriter, status int, message string) {
	write(w, Envelope{Status: status, Message: message})
}

// ErrorData refuses with a message and details the app uses (attempts left, a difference…).
func ErrorData(w http.ResponseWriter, status int, message string, data any) {
	write(w, Envelope{Status: status, Message: message, Data: data})
}

// ServerError logs the real cause and tells the caller nothing about it.
func ServerError(w http.ResponseWriter, what string, err error) {
	logx.RecordError(w, what, err) // reported on the request's access line, with its request id
	Error(w, http.StatusInternalServerError, MsgServerError)
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID guards ids taken from URLs before they reach a uuid cast in SQL.
func IsUUID(s string) bool { return uuidPattern.MatchString(s) }

// ParseTime accepts a date (2026-01-31) or an RFC 3339 timestamp.
func ParseTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.DateOnly, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// DecodeJSON reads a size-limited JSON body into dst, or writes 400.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody)).Decode(dst); err != nil {
		Error(w, http.StatusBadRequest, MsgBadJSON)
		return false
	}
	return true
}

// URLID returns the {id} URL parameter, or writes 404 if it is not a uuid.
func URLID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, ParamID)
	if !IsUUID(id) {
		Error(w, http.StatusNotFound, MsgNotFound)
		return "", false
	}
	return id, true
}
