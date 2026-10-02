// Package server builds the HTTP router for the API.
package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// NewRouter takes ping so the health check can be tested without a database.
func NewRouter(ping func(context.Context) error, a *auth.Auth) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			log.Printf("healthz: %v", err)
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unavailable"})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/auth/login", a.Login)

	// Everything below needs a signed-in user and a resolved organisation.
	// Guard individual routes with a.Require("resource:action").
	r.Group(func(r chi.Router) {
		r.Use(a.Authenticate, a.ResolveOrg)
		r.Get("/me", a.Me)
	})

	return r
}
