package auth

import (
	"net/http"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// A signing action (issue, handover, transform, certificate, flag resolution)
// needs the user's signing PIN on top of the sign-in, so a session left open
// on a shared device cannot sign for the organisation.

func validPIN(pin string) bool {
	if len(pin) != PINLen {
		return false
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// CheckPIN checks the X-Signing-PIN header against the signed-in user's PIN,
// or writes the refusal. Five wrong PINs lock signing for the window. Must
// run after Authenticate.
func (a *Auth) CheckPIN(w http.ResponseWriter, r *http.Request) bool {
	ctx, u := r.Context(), CurrentUser(r.Context())
	if a.Limiter.blocked(ctx, PINRate, u.ID) {
		httpx.Error(w, http.StatusLocked, MsgPINLocked)
		return false
	}
	hash, err := a.Store.PINHash(ctx, u.ID)
	if err != nil && err != ErrNotFound {
		httpx.ServerError(w, "load signing pin", err)
		return false
	}
	pin := strings.TrimSpace(r.Header.Get(HeaderSigningPIN))
	switch {
	case hash == "":
		httpx.Error(w, http.StatusForbidden, MsgPINNotSet)
		return false
	case pin == "":
		httpx.Error(w, http.StatusForbidden, MsgPINNeeded)
		return false
	case !CheckPassword(hash, pin):
		a.Limiter.record(ctx, PINRate, u.ID)
		httpx.Error(w, http.StatusForbidden, MsgBadPIN)
		return false
	}
	a.Limiter.reset(ctx, PINRate, u.ID)
	return true
}

// SetPIN sets or changes the signing PIN. It needs the account password, so
// an open session alone cannot change it.
func (a *Auth) SetPIN(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PIN      string `json:"pin"`
		Password string `json:"password"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if !validPIN(in.PIN) {
		httpx.Error(w, http.StatusBadRequest, MsgPINFormat)
		return
	}
	u, ok := a.freshUser(w, r)
	if !ok {
		return
	}
	ctx, ip := r.Context(), clientIP(r)
	if a.Limiter.blocked(ctx, LoginRate, u.Email, ip) {
		httpx.Error(w, http.StatusTooManyRequests, MsgTooManyAttempts)
		return
	}
	if !CheckPassword(u.PasswordHash, in.Password) {
		a.Limiter.record(ctx, LoginRate, u.Email, ip)
		httpx.Error(w, http.StatusForbidden, MsgBadCredentials)
		return
	}
	hash, err := HashPassword(in.PIN)
	if err != nil {
		httpx.ServerError(w, "hash signing pin", err)
		return
	}
	if err := a.Store.SetPIN(ctx, u.ID, hash); err != nil {
		httpx.ServerError(w, "store signing pin", err)
		return
	}
	a.Limiter.reset(ctx, PINRate, u.ID)
	httpx.OK(w, http.StatusOK, MsgPINSet, map[string]bool{"pin_set": true})
}
