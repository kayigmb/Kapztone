package auth

import (
	"errors"
	"net/http"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// The session cookie belongs to one address (acme.chaintrack.app), so moving
// to another organisation's address would mean signing in again. A handoff
// carries the session across instead: the current address asks for a code
// bound to the target address (StartHandoff), the target spends it once
// (AcceptHandoff) and gets its own session. The code lives a minute, and
// only its hash with the target's slug is stored, so it works nowhere else.

func handoffHash(code, slug string) string { return linkHash(code + ":" + slug) }

// StartHandoff answers POST /auth/handoff {slug}: a code for that
// organisation's address, or for the bare address when slug is "".
func (a *Auth) StartHandoff(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Slug string `json:"slug"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	ctx, userID := r.Context(), CurrentUser(r.Context()).ID
	if in.Slug != "" {
		scope, err := a.Store.Scope(ctx, userID, in.Slug)
		if err != nil && !errors.Is(err, ErrNotFound) {
			httpx.ServerError(w, "handoff scope", err)
			return
		}
		if scope.Organisation == nil || scope.Organisation.Slug != in.Slug {
			httpx.Error(w, http.StatusForbidden, MsgNotMember)
			return
		}
	}
	code, _, err := newLinkToken()
	if err == nil {
		err = a.Accounts.AddLink(ctx, userID, LinkHandoff, handoffHash(code, in.Slug), a.now().Add(HandoffTTL))
	}
	if err != nil {
		httpx.ServerError(w, "start handoff", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", map[string]string{"code": code})
}

// AcceptHandoff answers POST /auth/handoff/accept {code} on the target
// address: a code made for this address signs the person in here.
func (a *Auth) AcceptHandoff(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Code string `json:"code"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	ctx, slug := r.Context(), a.Slug(r)
	userID, err := a.Accounts.UseLink(ctx, LinkHandoff, handoffHash(in.Code, slug), a.now())
	if errors.Is(err, ErrNotFound) && slug != "" {
		// The bare address may still have a first label (chaintrack in
		// chaintrack.app): a code for the bare address works there, but never
		// at the address of an organisation the person belongs to.
		userID, err = a.Accounts.UseLink(ctx, LinkHandoff, handoffHash(in.Code, ""), a.now())
		if err == nil {
			var scope Scope
			if scope, err = a.Store.Scope(ctx, userID, slug); err == nil && scope.Organisation != nil && scope.Organisation.Slug == slug {
				err = ErrNotFound
			} else if errors.Is(err, ErrNotFound) {
				err = nil
			}
		}
	}
	var u User
	if err == nil {
		u, err = a.Store.ActiveUser(ctx, userID)
	}
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusBadRequest, MsgHandoffInvalid)
	case err != nil:
		httpx.ServerError(w, "accept handoff", err)
	default:
		a.signIn(w, u.ID, u.TokenVersion, a.TokenTTL, "")
	}
}
