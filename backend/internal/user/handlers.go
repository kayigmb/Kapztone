package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

type Handlers struct {
	Store Store
	// InvalidateUser clears cached access after a membership change.
	InvalidateUser func(ctx context.Context, userID string)
	// Mail sends one e-mail; nil when e-mail is off.
	Mail func(to, subject, body string) error
	// Dev also returns the invitation link (and logs it), since a developer
	// usually has e-mail off. Never in production: the link is a credential.
	Dev bool
	Now func() time.Time // nil means time.Now
}

func (h *Handlers) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// actingOrg is the organisation the admin manages; platform scopes act for none.
func actingOrg(w http.ResponseWriter, r *http.Request) (*auth.Organisation, bool) {
	o := auth.CurrentScope(r.Context()).Organisation
	if o == nil {
		httpx.Error(w, http.StatusForbidden, auth.MsgNoOrgSelected)
	}
	return o, o != nil
}

func (h *Handlers) fail(w http.ResponseWriter, what string, err error) {
	switch {
	case errors.Is(err, ErrNotMember):
		httpx.Error(w, http.StatusNotFound, MsgNotMember)
	case errors.Is(err, ErrLastAdmin):
		httpx.Error(w, http.StatusConflict, MsgLastAdmin)
	case errors.Is(err, ErrNotSuspended):
		httpx.Error(w, http.StatusConflict, MsgNotSuspended)
	case errors.Is(err, ErrAlreadyMember):
		httpx.Error(w, http.StatusConflict, MsgAlreadyMember)
	case errors.Is(err, ErrInviteInvalid):
		httpx.Error(w, http.StatusNotFound, MsgInviteInvalid)
	case errors.Is(err, ErrNameNeeded):
		httpx.Error(w, http.StatusBadRequest, MsgName)
	case errors.Is(err, ErrBadPassword):
		httpx.Error(w, http.StatusForbidden, auth.MsgBadCredentials)
	default:
		httpx.ServerError(w, what, err)
	}
}

// List answers GET /users: the acting organisation's members, or for the
// platform every user with all their memberships.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	if gate, ok := auth.CurrentScope(r.Context()).Gate(); ok && gate == "" {
		accounts, err := h.Store.All(r.Context())
		if err != nil {
			httpx.ServerError(w, "list users", err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, accounts)
		return
	}
	o, ok := actingOrg(w, r)
	if !ok {
		return
	}
	members, err := h.Store.List(r.Context(), o.ID)
	if err != nil {
		httpx.ServerError(w, "list members", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, members)
}

type roleInput struct {
	Role string `json:"role"`
}

func validRole(w http.ResponseWriter, role string) bool {
	if !slices.Contains(AssignableRoles, role) {
		httpx.Error(w, http.StatusBadRequest, "role must be one of: "+strings.Join(AssignableRoles, ", "))
		return false
	}
	return true
}

// Invite e-mails a one-time link to join the acting organisation with a role.
func (h *Handlers) Invite(w http.ResponseWriter, r *http.Request) {
	o, ok := actingOrg(w, r)
	if !ok {
		return
	}
	var in struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if a, err := mail.ParseAddress(in.Email); err != nil || a.Address != in.Email || len(in.Email) > MaxEmailLen {
		httpx.Error(w, http.StatusBadRequest, MsgEmail)
		return
	}
	if !validRole(w, in.Role) {
		return
	}

	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		httpx.ServerError(w, "invite token", err)
		return
	}
	token := hex.EncodeToString(raw)
	inv, err := h.Store.CreateInvite(r.Context(), o.ID, in.Email, in.Role, tokenHash(token),
		auth.CurrentUser(r.Context()).ID, h.now().Add(InviteTTL))
	if err != nil {
		httpx.ServerError(w, "create invite", err)
		return
	}

	// The link opens the invitation on the organisation's own address.
	link := auth.FrontendOf(r.Context()).Origin + InvitePath + token
	out := map[string]any{"invite": inv, "emailed": false}
	if h.Mail != nil {
		body := fmt.Sprintf(MsgInviteBodyFmt, o.Name, in.Role, link)
		if err := h.Mail(in.Email, InviteSubject, body); err != nil {
			log.Printf("invite %s: send e-mail: %v", inv.ID, err)
		} else {
			out["emailed"] = true
		}
	}
	if h.Dev {
		log.Printf("invite %s for %s: %s", inv.ID, in.Email, link)
		out["link"] = link
	}
	httpx.WriteJSON(w, http.StatusCreated, out)
}

// Accept is public: the token in the link is the proof of invitation. A new
// person sets their name and password; an existing account gives its password.
func (h *Handlers) Accept(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, ParamToken)
	var in struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) > MaxNameLen {
		httpx.Error(w, http.StatusBadRequest, MsgName)
		return
	}
	if n := len(in.Password); n < auth.MinPasswordLen || n > auth.MaxPasswordLen {
		// An existing account's password met the same rule when it was set.
		httpx.Error(w, http.StatusBadRequest, fmt.Sprintf(MsgPasswordFormat, auth.MinPasswordLen, auth.MaxPasswordLen))
		return
	}
	acc, err := h.Store.Accept(r.Context(), tokenHash(token), in.Name, in.Password)
	if err != nil {
		h.fail(w, "accept invite", err)
		return
	}
	h.InvalidateUser(r.Context(), acc.UserID)
	httpx.WriteJSON(w, http.StatusOK, acc)
}

func (h *Handlers) SetRole(w http.ResponseWriter, r *http.Request) {
	o, ok := actingOrg(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in roleInput
	if !httpx.DecodeJSON(w, r, &in) || !validRole(w, in.Role) {
		return
	}
	if err := h.Store.SetRole(r.Context(), o.ID, id, in.Role); err != nil {
		h.fail(w, "set role", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"role": in.Role})
}

// Suspend removes the person's access to the acting organisation only.
func (h *Handlers) Suspend(w http.ResponseWriter, r *http.Request) {
	o, ok := actingOrg(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	if id == auth.CurrentUser(r.Context()).ID {
		httpx.Error(w, http.StatusConflict, MsgSelf)
		return
	}
	if err := h.Store.Suspend(r.Context(), o.ID, id); err != nil {
		h.fail(w, "suspend member", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": StatusSuspended})
}

func (h *Handlers) Reinstate(w http.ResponseWriter, r *http.Request) {
	o, ok := actingOrg(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	if err := h.Store.Reinstate(r.Context(), o.ID, id); err != nil {
		h.fail(w, "reinstate member", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": StatusActive})
}
