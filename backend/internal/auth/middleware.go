package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

type ctxKey int

const (
	userKey ctxKey = iota
	scopeKey
)

// CurrentUser returns the signed-in user. Only valid behind Authenticate.
func CurrentUser(ctx context.Context) User {
	u, _ := ctx.Value(userKey).(User)
	return u
}

// CurrentScope returns the acting scope. Only valid behind ResolveOrg.
func CurrentScope(ctx context.Context) Scope {
	s, _ := ctx.Value(scopeKey).(Scope)
	return s
}

// Authenticate requires a valid sign-in token for a user who is still active.
func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			httpx.Error(w, http.StatusUnauthorized, "sign in required")
			return
		}
		userID, err := a.ParseToken(token)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "sign in required")
			return
		}
		user, err := a.Store.ActiveUser(r.Context(), userID)
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusUnauthorized, "sign in required")
			return
		}
		if err != nil {
			serverError(w, "authenticate", err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}

// ResolveOrg works out which organisation the user is acting for and loads
// their role and permissions there. Must run after Authenticate.
func (a *Auth) ResolveOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := a.Slug(r)
		scope, err := a.Store.Scope(r.Context(), CurrentUser(r.Context()).ID, slug)
		switch {
		case errors.Is(err, ErrNotFound) && slug != "":
			// The request names an organisation the user does not belong to.
			httpx.Error(w, http.StatusForbidden, "you do not have access to this organisation")
			return
		case errors.Is(err, ErrNotFound):
			scope = Scope{Permissions: []string{}} // no organisation named: signed in, no permissions
		case err != nil:
			serverError(w, "resolve organisation", err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), scopeKey, scope)))
	})
}

// Require allows the request only if the acting scope has the permission,
// e.g. Require("passport:issue"). Must run after ResolveOrg.
func (a *Auth) Require(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !slices.Contains(CurrentScope(r.Context()).Permissions, permission) {
				httpx.Error(w, http.StatusForbidden, "you are not allowed to do this")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// dummyHash is compared against when the email is unknown, so a wrong email
// takes as long as a wrong password and the two cannot be told apart.
var dummyHash, _ = HashPassword("chaintrack-dummy-password")

// Login exchanges an email and password for a sign-in token.
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "send email and password as JSON")
		return
	}

	user, err := a.Store.UserByEmail(r.Context(), strings.TrimSpace(in.Email))
	if err != nil && !errors.Is(err, ErrNotFound) {
		serverError(w, "login", err)
		return
	}
	hash := user.PasswordHash
	if err != nil {
		hash = dummyHash
	}
	if !CheckPassword(hash, in.Password) || err != nil {
		httpx.Error(w, http.StatusUnauthorized, "email or password is incorrect")
		return
	}

	token, err := a.IssueToken(user.ID, time.Now())
	if err != nil {
		serverError(w, "issue token", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_in": int(a.TokenTTL.Seconds()),
	})
}

// Me reports who is signed in and what they may do in the acting organisation.
func (a *Auth) Me(w http.ResponseWriter, r *http.Request) {
	scope := CurrentScope(r.Context())
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user":         CurrentUser(r.Context()),
		"organisation": scope.Organisation,
		"role":         scope.Role,
		"permissions":  scope.Permissions,
	})
}

func serverError(w http.ResponseWriter, what string, err error) {
	log.Printf("%s: %v", what, err)
	httpx.Error(w, http.StatusInternalServerError, "something went wrong")
}
