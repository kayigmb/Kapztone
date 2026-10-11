// Package httpx holds the HTTP helpers shared by handlers and middleware.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
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

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// ServerError logs the real cause and tells the caller nothing about it.
func ServerError(w http.ResponseWriter, what string, err error) {
	log.Printf("%s: %v", what, err)
	Error(w, http.StatusInternalServerError, MsgServerError)
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID guards ids taken from URLs before they reach a uuid cast in SQL.
func IsUUID(s string) bool { return uuidPattern.MatchString(s) }

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
