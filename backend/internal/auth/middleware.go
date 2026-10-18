package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

type ctxKey int

const (
	userKey ctxKey = iota
	scopeKey
	frontendKey
)

// WithFrontend records, for every request, which site the user's browser is
// on, so any handler can ask with FrontendOf.
func WithFrontend(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), frontendKey, frontendOf(r))))
	})
}

// FrontendOf returns the site the request came from. Valid behind WithFrontend.
func FrontendOf(ctx context.Context) Frontend {
	f, _ := ctx.Value(frontendKey).(Frontend)
	return f
}

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

// sessionToken finds the sign-in token: in the Authorization header (scripts,
// other services), else in the session cookie (browsers).
func sessionToken(r *http.Request) (token string, fromCookie bool) {
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), bearerPrefix); ok {
		return token, false
	}
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value, true
	}
	return "", false
}

// sameSite reports whether a browser's request came from a page of the site
// it is addressed to. A browser attaches the session cookie to any request
// for this site, whichever site's page caused it, and names that page's site
// in Origin; a request that changes something must come from our own pages.
// SameSite on the cookie does not cover this alone: another organisation's
// subdomain counts as the same site.
func sameSite(r *http.Request) bool {
	u, err := url.Parse(r.Header.Get(HeaderOrigin))
	return err == nil && u.Host != "" && strings.EqualFold(u.Host, siteHost(r))
}

func (a *Auth) sessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name: SessionCookie, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true,                 // never readable by scripts
		Secure:   !a.Dev,               // https only, except on a developer's machine
		SameSite: http.SameSiteLaxMode, // not sent on requests other sites make in the background
	}
}

// Authenticate requires a valid sign-in token for a user who is still active.
func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, fromCookie := sessionToken(r)
		if token == "" {
			httpx.Error(w, http.StatusUnauthorized, MsgSignInRequired)
			return
		}
		reads := r.Method == http.MethodGet || r.Method == http.MethodHead
		if fromCookie && !reads && !sameSite(r) {
			httpx.Error(w, http.StatusForbidden, MsgForeignOrigin)
			return
		}
		userID, ver, err := a.parse(token)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, MsgSignInRequired)
			return
		}
		user, err := a.Store.ActiveUser(r.Context(), userID)
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusUnauthorized, MsgSignInRequired)
			return
		}
		if err != nil {
			httpx.ServerError(w, "authenticate", err)
			return
		}
		// Issued before a password change, sign-out-everywhere or lock: void.
		if ver != user.TokenVersion {
			httpx.Error(w, http.StatusUnauthorized, MsgSignInRequired)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}

// ResolveOrg works out which organisation the user is acting for and loads
// their role and permissions there. Must run after Authenticate.
func (a *Auth) ResolveOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope, err := a.Store.Scope(r.Context(), CurrentUser(r.Context()).ID, a.Slug(r))
		switch {
		case errors.Is(err, ErrNotFound):
			// No organisation named, or one the user does not belong to:
			// signed in, acting for nobody, holding no permissions.
			scope = Scope{Permissions: []string{}}
		case err != nil:
			httpx.ServerError(w, "resolve organisation", err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), scopeKey, scope)))
	})
}

// Require allows the request only if the acting scope has the permission,
// e.g. Require(PassportIssue). Must run after ResolveOrg.
func (a *Auth) Require(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !slices.Contains(CurrentScope(r.Context()).Permissions, permission) {
				httpx.Error(w, http.StatusForbidden, MsgNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireVerified allows the request only when the acting organisation has
// passed KYB. Must run after ResolveOrg.
func (a *Auth) RequireVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := CurrentScope(r.Context()).Organisation
		if o == nil {
			httpx.Error(w, http.StatusForbidden, MsgNoOrgSelected)
			return
		}
		if o.Status != OrgVerified {
			httpx.Error(w, http.StatusForbidden, MsgOrgNotVerified)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Gate returns the organisation filter for this request (see Scope.Gate), or
// writes 403 when the caller has no data access.
func Gate(w http.ResponseWriter, r *http.Request) (orgID string, ok bool) {
	orgID, ok = CurrentScope(r.Context()).Gate()
	if !ok {
		httpx.Error(w, http.StatusForbidden, MsgNoOrgSelected)
	}
	return orgID, ok
}

// dummyHash is compared against when the email is unknown, so a wrong email
// takes as long as a wrong password and the two cannot be told apart.
var dummyHash, _ = HashPassword("chaintrack-dummy-password")

// Login exchanges an email and password (and a 2FA code, if the user has
// 2FA on and this device isn't trusted) for a sign-in token. Remember asks
// for the longer sign-in; TrustDevice, with a valid app code, lets this
// device skip the code for TrustTTL.
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		Code        string `json:"code"`
		Remember    bool   `json:"remember"`
		TrustDevice bool   `json:"trust_device"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	ctx, ip := r.Context(), clientIP(r)

	if a.Limiter.blocked(ctx, LoginRate, in.Email, ip) {
		httpx.Error(w, http.StatusTooManyRequests, MsgTooManyAttempts)
		return
	}
	fail := func(body map[string]any) {
		a.Limiter.record(ctx, LoginRate, in.Email, ip)
		body["attempts_left"] = a.Limiter.left(ctx, LoginRate, in.Email, ip)
		httpx.WriteJSON(w, http.StatusUnauthorized, body)
	}

	user, err := a.Store.UserByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		httpx.ServerError(w, "login", err)
		return
	}
	hash := user.PasswordHash
	if err != nil {
		hash = dummyHash
	}
	if !CheckPassword(hash, in.Password) || err != nil {
		fail(map[string]any{"error": MsgBadCredentials})
		return
	}

	now := a.now()
	setTrust := false
	if user.TOTPEnabled && !a.trusted(r, user, now) {
		valid, err := a.checkCode(ctx, user, in.Code, now)
		setTrust = valid && in.TrustDevice // an app code, not a recovery code
		if err == nil && !valid && strings.TrimSpace(in.Code) != "" {
			// A lost phone: a recovery code works once in place of the app code.
			valid, err = a.Store.UseRecoveryCode(ctx, user.ID, recoveryHash(in.Code))
		}
		if err != nil {
			httpx.ServerError(w, "check 2fa code", err)
			return
		}
		if !valid {
			msg := MsgBadCode
			if strings.TrimSpace(in.Code) == "" {
				msg = MsgTwoFactorCode
			}
			fail(map[string]any{"error": msg, "two_factor_required": true})
			return
		}
	}
	a.Limiter.reset(ctx, LoginRate, in.Email, ip)

	if setTrust {
		http.SetCookie(w, a.trustCookie(user, now))
	}
	ttl := a.TokenTTL
	if in.Remember && a.RememberTTL > 0 {
		ttl = a.RememberTTL
	}
	// Browsers keep the token in the cookie. It is also returned for callers
	// that are not browsers.
	a.signIn(w, user.ID, user.TokenVersion, ttl)
}

// Logout clears the session cookie. The page cannot do it: scripts have no
// access to the cookie. It needs no sign-in, since it only removes one.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, a.sessionCookie("", -1))
	w.WriteHeader(http.StatusNoContent)
}

// Me reports who is signed in and what they may do in the acting organisation.
func (a *Auth) Me(w http.ResponseWriter, r *http.Request) {
	scope := CurrentScope(r.Context())
	pin, err := a.Store.PINHash(r.Context(), CurrentUser(r.Context()).ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		httpx.ServerError(w, "load signing pin", err)
		return
	}
	full, ok := a.freshUser(w, r) // the cached user carries no 2FA state
	if !ok {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user":         map[string]string{"id": full.ID, "email": full.Email, "name": full.Name},
		"organisation": scope.Organisation,
		"role":         scope.Role,
		"permissions":  scope.Permissions,
		"pin_set":      pin != "",
		"two_factor":   full.TOTPEnabled,
	})
}
