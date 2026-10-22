package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/mailtmpl"
)

type Handlers struct {
	Store Store
	// InvalidateUser clears cached access after a membership change.
	InvalidateUser func(ctx context.Context, userID string)
	// Mail sends one e-mail; nil when e-mail is off.
	Mail func(to, subject, text, html string) error
	// LogoURL is the absolute URL of the ChainTrack mark shown atop HTML
	// mail; empty means the header shows the name as text.
	LogoURL string
	// AppURL is the web app's bare address; invitation links start with it.
	AppURL string
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

// isPlatform reports whether the caller acts for the whole platform (the superadmin).
func isPlatform(r *http.Request) bool {
	gate, ok := auth.CurrentScope(r.Context()).Gate()
	return ok && gate == ""
}

// platformOnly writes 403 unless the caller acts for the whole platform.
func platformOnly(w http.ResponseWriter, r *http.Request) bool {
	if !isPlatform(r) {
		httpx.Error(w, http.StatusForbidden, auth.MsgNotAllowed)
		return false
	}
	return true
}

// targetOrg is the organisation a membership change applies to: the acting
// one, or for the platform the one named by ?organisation_id=.
func targetOrg(w http.ResponseWriter, r *http.Request) (string, bool) {
	if isPlatform(r) {
		id := r.URL.Query().Get(auth.ParamOrganisation)
		if !httpx.IsUUID(id) {
			httpx.Error(w, http.StatusBadRequest, auth.MsgOrganisationParam)
			return "", false
		}
		return id, true
	}
	o, ok := actingOrg(w, r)
	if !ok {
		return "", false
	}
	return o.ID, true
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
	case errors.Is(err, ErrPasswordPolicy):
		httpx.Error(w, http.StatusBadRequest, auth.MsgPasswordComplexity)
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
		q, ok := listq.FromRequest(w, r, AccountList)
		if !ok {
			return
		}
		accounts, err := h.Store.All(r.Context(), q, "")
		if err != nil {
			httpx.ServerError(w, "list users", err)
			return
		}
		httpx.OK(w, http.StatusOK, "", accounts)
		return
	}
	o, ok := actingOrg(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, MemberList)
	if !ok {
		return
	}
	members, err := h.Store.List(r.Context(), q, o.ID)
	if err != nil {
		httpx.ServerError(w, "list members", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", members)
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
	link := auth.OrgURL(h.AppURL, o.Slug) + InvitePath + token
	out := map[string]any{"invite": inv, "emailed": false}
	if h.Mail != nil {
		// Queued, not sent: the invitation must not wait on the mail server
		// (a slow or unreachable server must not slow this response down).
		// A failure is logged; the admin re-invites and a new link is made.
		text := fmt.Sprintf(MsgInviteBodyFmt, o.Name, in.Role, link)
		html := mailtmpl.Render(mailtmpl.Page{
			Subject:     InviteSubject,
			Preheader:   "You have been invited to join " + o.Name + " on ChainTrack.",
			Paragraphs:  mailtmpl.Split(text),
			ActionLabel: "Accept invitation",
			ActionURL:   link,
			Footnote:    "This link works once, for 24 hours.",
			LogoURL:     h.LogoURL,
		})
		mail, to := h.Mail, in.Email
		go func() {
			if err := mail(to, InviteSubject, text, html); err != nil {
				logx.From(context.Background()).Error("invite e-mail failed", "invite_id", inv.ID, logx.KeyError, err)
			}
		}()
		out["emailed"] = true
	}
	if h.Dev {
		logx.From(r.Context()).Info("dev: invite link not e-mailed", "invite_id", inv.ID, "link", link)
		out["link"] = link
	}
	msg := MsgInviteLink
	if out["emailed"] == true {
		msg = MsgInviteSent(inv.Email)
	}
	httpx.OK(w, http.StatusCreated, msg, out)
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
	httpx.OK(w, http.StatusOK, MsgInviteAccepted, acc)
}

func (h *Handlers) SetRole(w http.ResponseWriter, r *http.Request) {
	orgID, ok := targetOrg(w, r)
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
	if id == auth.CurrentUser(r.Context()).ID {
		httpx.Error(w, http.StatusConflict, MsgSelfRole)
		return
	}
	if err := h.Store.SetRole(r.Context(), orgID, id, in.Role); err != nil {
		h.fail(w, "set role", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	httpx.OK(w, http.StatusOK, MsgRoleChanged, map[string]string{"role": in.Role})
}

// Suspend removes the person's access to the acting organisation only.
func (h *Handlers) Suspend(w http.ResponseWriter, r *http.Request) {
	orgID, ok := targetOrg(w, r)
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
	if err := h.Store.Suspend(r.Context(), orgID, id); err != nil {
		h.fail(w, "suspend member", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	httpx.OK(w, http.StatusOK, MsgSuspended, map[string]string{"status": StatusSuspended})
}

func (h *Handlers) Reinstate(w http.ResponseWriter, r *http.Request) {
	orgID, ok := targetOrg(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	if err := h.Store.Reinstate(r.Context(), orgID, id); err != nil {
		h.fail(w, "reinstate member", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	httpx.OK(w, http.StatusOK, MsgReinstated, map[string]string{"status": StatusActive})
}

// Get answers GET /users/{id} for the platform: one account with every membership.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok || !platformOnly(w, r) {
		return
	}
	a, err := h.Store.Account(r.Context(), id)
	if errors.Is(err, ErrNotMember) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if err != nil {
		httpx.ServerError(w, "get user", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", a)
}

// Deactivate answers POST /users/{id}/deactivate (platform): the account can't
// sign in anywhere and every device is signed out.
func (h *Handlers) Deactivate(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, false, auth.MsgUserDeactivated)
}

// Activate answers POST /users/{id}/activate (platform).
func (h *Handlers) Activate(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, true, auth.MsgUserActivated)
}

func (h *Handlers) setActive(w http.ResponseWriter, r *http.Request, active bool, msg string) {
	id, ok := httpx.URLID(w, r)
	if !ok || !platformOnly(w, r) {
		return
	}
	if !active && id == auth.CurrentUser(r.Context()).ID {
		httpx.Error(w, http.StatusConflict, auth.MsgCannotDisableSelf)
		return
	}
	if err := h.Store.SetActive(r.Context(), id, active); err != nil {
		h.fail(w, "set account status", err)
		return
	}
	h.InvalidateUser(r.Context(), id)
	status := auth.UserDisabled
	if active {
		status = auth.UserActive
	}
	httpx.OK(w, http.StatusOK, msg, map[string]string{"status": status})
}

// OrgMembers answers GET /orgs/{id}/members for the platform: everyone who
// belongs or belonged to that organisation.
func (h *Handlers) OrgMembers(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok || !platformOnly(w, r) {
		return
	}
	q, ok := listq.FromRequest(w, r, MemberList)
	if !ok {
		return
	}
	members, err := h.Store.List(r.Context(), q, id)
	if err != nil {
		httpx.ServerError(w, "list organisation members", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", members)
}
