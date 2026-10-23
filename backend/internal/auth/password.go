package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/mailtmpl"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

// The account flows: change, forget and reset a password, sign out every
// other device, and lock the account from the email sent after a change.
// Each one bumps the user's token version, which voids every sign-in token
// and trusted-device cookie issued before.

// invalidate drops the cached user, so the new token version is seen at once.
func (a *Auth) invalidate(ctx context.Context, userID string) {
	if c, ok := a.Store.(interface {
		InvalidateUser(context.Context, string)
	}); ok {
		c.InvalidateUser(ctx, userID)
	}
}

// signIn issues a token at the user's current version, sets the session
// cookie, and writes the token for callers that are not browsers.
// signIn sets this address's cookie for the browser session sid; "" opens a new one.
func (a *Auth) signIn(w http.ResponseWriter, r *http.Request, userID string, ver int, sid string, ttl time.Duration, message string) {
	if sid == "" {
		var err error
		if sid, err = a.Store.StartSession(r.Context(), userID); err != nil {
			httpx.ServerError(w, "start session", err)
			return
		}
	}
	token, err := a.issue(userID, ver, sid, ttl, a.now())
	if err != nil {
		httpx.ServerError(w, "issue token", err)
		return
	}
	http.SetCookie(w, a.sessionCookie(token, int(ttl.Seconds())))
	httpx.OK(w, http.StatusOK, message, map[string]any{"token": token, "expires_in": int(ttl.Seconds())})
}

// passwordProblem says what is wrong with a new password, or "" if nothing is.
func (a *Auth) passwordProblem(ctx context.Context, userID, password string) (string, error) {
	if problem := PasswordFormatProblem(password); problem != "" {
		return problem, nil
	}
	recent, err := a.Accounts.RecentPasswordHashes(ctx, userID, PasswordHistoryLen)
	if err != nil {
		return "", err
	}
	for _, h := range recent {
		if CheckPassword(h, password) {
			return MsgPasswordReused, nil
		}
	}
	return "", nil
}

// PasswordFormatProblem enforces the policy for every newly chosen password.
// The byte limit matches bcrypt's 72-byte input limit.
func PasswordFormatProblem(password string) string {
	if n := len(password); n < MinPasswordLen || n > MaxPasswordLen {
		return MsgPasswordLength
	}
	var lower, upper, digit, symbol bool
	for _, r := range password {
		lower = lower || unicode.IsLower(r)
		upper = upper || unicode.IsUpper(r)
		digit = digit || unicode.IsNumber(r)
		symbol = symbol || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	if !lower || !upper || !digit || !symbol {
		return MsgPasswordComplexity
	}
	return ""
}

// newLinkToken is a random token for an emailed link and the hash that is stored.
func newLinkToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, linkHash(token), nil
}

func linkHash(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

// sendLink stores a one-use link and emails it. In dev the link is also
// logged (never returned: the reset form must not say whether an account
// exists). A mail failure is logged, not returned, for the same reason.
func (a *Auth) sendLink(ctx context.Context, u User, kind, path string, ttl time.Duration, subject, bodyFmt, actionLabel, footnote string) (link string, err error) {
	token, hash, err := newLinkToken()
	if err != nil {
		return "", err
	}
	if err := a.Accounts.AddLink(ctx, u.ID, kind, hash, a.now().Add(ttl)); err != nil {
		return "", err
	}
	link = a.AppURL + path + token
	if a.Dev {
		logx.From(ctx).Info("dev: link not e-mailed", "kind", kind, "user_id", u.ID, "link", link)
	}
	if a.Mail != nil {
		// Sent in the background so a slow mail server doesn't make an existing
		// account answer more slowly than an unknown email.
		text := fmt.Sprintf(bodyFmt, link)
		html := mailtmpl.Render(mailtmpl.Page{
			Subject:     subject,
			Preheader:   subject,
			Paragraphs:  mailtmpl.Split(text),
			ActionLabel: actionLabel,
			ActionURL:   link,
			Footnote:    footnote,
		})
		go func() {
			if err := a.Mail(u.Email, subject, text, html); err != nil {
				logx.From(ctx).Error("mail link failed", "kind", kind, logx.KeyUser, u.ID, logx.KeyError, err)
			}
		}()
	}
	return link, nil
}

// passwordChanged tells the user, with a link to lock the account if it wasn't them.
func (a *Auth) passwordChanged(ctx context.Context, u User) {
	if _, err := a.sendLink(ctx, u, LinkLock, PathLock, LockLinkTTL, MailChangedSubject, MailChangedBody,
		"Lock my account", "This link works once, for 7 days."); err != nil {
		logx.From(ctx).Error("password changed notice failed", logx.KeyUser, u.ID, logx.KeyError, err)
	}
}

// ChangePassword answers POST /auth/password: the signed-in person proves the
// current password and sets a new one. A wrong current password counts toward
// the sign-in lockout, like a failed sign-in. Every other device is signed
// out; this one gets a new token.
func (a *Auth) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
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
	if !CheckPassword(u.PasswordHash, in.Current) {
		a.Limiter.record(ctx, LoginRate, u.Email, ip)
		httpx.Error(w, http.StatusForbidden, MsgBadCredentials)
		return
	}
	ver, ok := a.setPassword(w, r, u, in.New)
	if !ok {
		return
	}
	a.Limiter.reset(ctx, LoginRate, u.Email, ip)
	a.passwordChanged(ctx, u)
	a.signIn(w, r, u.ID, ver, CurrentSession(r.Context()), a.TokenTTL, MsgPasswordChanged)
}

// setPassword checks and stores a new password, bumping the token version.
func (a *Auth) setPassword(w http.ResponseWriter, r *http.Request, u User, password string) (int, bool) {
	ctx := r.Context()
	msg, err := a.passwordProblem(ctx, u.ID, password)
	if err != nil {
		httpx.ServerError(w, "check password history", err)
		return 0, false
	}
	if msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return 0, false
	}
	hash, err := HashPassword(password)
	if err != nil {
		httpx.ServerError(w, "hash password", err)
		return 0, false
	}
	ver, err := a.Accounts.SetPassword(ctx, u.ID, hash)
	if err != nil {
		httpx.ServerError(w, "store password", err)
		return 0, false
	}
	a.invalidate(ctx, u.ID)
	return ver, true
}

// ForgotPassword answers POST /auth/password/forgot (public). The answer is
// the same whether or not the email has an account.
func (a *Auth) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	// Answered before the account is even looked up, so a known and an unknown
	// email take the same time and get the same answer, failures included.
	// WithoutCancel keeps the request's values (the site address for the link).
	ctx, email := context.WithoutCancel(r.Context()), strings.ToLower(strings.TrimSpace(in.Email))
	httpx.OK(w, http.StatusAccepted, MsgResetSent, nil)
	go func() {
		u, err := a.Store.UserByEmail(ctx, email)
		if err == nil {
			_, err = a.sendLink(ctx, u, LinkReset, PathReset, ResetLinkTTL, MailResetSubject, MailResetBody,
				"Choose a new password", "This link works once, for 30 minutes. If you didn't ask for it, ignore this email.")
		}
		if err != nil && !errors.Is(err, ErrNotFound) {
			logx.From(ctx).Error("reset link failed", logx.KeyError, err)
		}
	}()
}

// ResetPassword answers POST /auth/password/reset (public): an emailed link
// sets a new password once. Every device is signed out.
func (a *Auth) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
		New   string `json:"new"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := PasswordFormatProblem(in.New); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	ctx, hash := r.Context(), linkHash(in.Token)
	u, ok := a.resetUser(w, r, hash, in.New)
	if !ok {
		return
	}
	if _, err := a.Accounts.UseLink(ctx, LinkReset, hash, a.now()); errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, MsgLinkInvalid) // spent by a concurrent request
		return
	} else if err != nil {
		httpx.ServerError(w, "spend reset link", err)
		return
	}
	if _, ok := a.setPassword(w, r, u, in.New); !ok {
		return
	}
	a.passwordChanged(ctx, u)
	httpx.OK(w, http.StatusOK, MsgPasswordReset, nil)
}

// resetUser finds the link's user and checks the new password against their
// history before the link is spent, so a refused password leaves it usable.
func (a *Auth) resetUser(w http.ResponseWriter, r *http.Request, hash, password string) (User, bool) {
	ctx := r.Context()
	userID, err := a.Accounts.LinkUser(ctx, LinkReset, hash, a.now())
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, MsgLinkInvalid)
		return User{}, false
	}
	if err != nil {
		httpx.ServerError(w, "read reset link", err)
		return User{}, false
	}
	u, err := a.Accounts.UserByID(ctx, userID)
	if err != nil {
		httpx.ServerError(w, "load user", err)
		return User{}, false
	}
	msg, err := a.passwordProblem(ctx, u.ID, password)
	switch {
	case err != nil:
		httpx.ServerError(w, "check password history", err)
	case msg != "":
		httpx.Error(w, http.StatusBadRequest, msg)
	default:
		return u, true
	}
	return User{}, false
}

// SendVerify emails the user a link that confirms their email.
func (a *Auth) SendVerify(ctx context.Context, u User) error {
	_, err := a.sendVerify(ctx, u)
	return err
}

func (a *Auth) sendVerify(ctx context.Context, u User) (string, error) {
	return a.sendLink(ctx, u, LinkVerify, PathVerify, VerifyLinkTTL, MailVerifySubject, MailVerifyBody,
		"Confirm my email", "This link works once, for 24 hours.")
}

// SendVerifyTo sends the confirmation link to a just-registered email. A
// failure is logged: the account exists, and signing in sends a new link.
// In dev it also returns the link, as invites do, so scripts and the browser
// tests can confirm without a mailbox; elsewhere it returns "".
func (a *Auth) SendVerifyTo(ctx context.Context, email string) (devLink string) {
	u, err := a.Store.UserByEmail(ctx, email)
	var link string
	if err == nil {
		link, err = a.sendVerify(ctx, u)
	}
	if err != nil {
		logx.From(ctx).Error("verify link failed", logx.KeyError, err)
	}
	if !a.Dev {
		return ""
	}
	return link
}

// VerifyEmail answers POST /auth/email/verify (public), the link in the
// confirmation email.
func (a *Auth) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	ctx := r.Context()
	userID, err := a.Accounts.UseLink(ctx, LinkVerify, linkHash(in.Token), a.now())
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, MsgLinkInvalid)
		return
	}
	if err == nil {
		err = a.Accounts.MarkVerified(ctx, userID)
	}
	if err != nil {
		httpx.ServerError(w, "verify email", err)
		return
	}
	httpx.OK(w, http.StatusOK, MsgEmailVerified, nil)
}

// LockAccount answers POST /auth/lock (public), the link in the "password
// changed" email: the account is disabled and every device signed out.
// Unlocking is a platform admin's job.
func (a *Auth) LockAccount(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token string `json:"token"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	ctx := r.Context()
	userID, err := a.Accounts.UseLink(ctx, LinkLock, linkHash(in.Token), a.now())
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, MsgLinkInvalid)
		return
	}
	if err == nil {
		err = a.Accounts.Lock(ctx, userID)
	}
	if err != nil {
		httpx.ServerError(w, "lock account", err)
		return
	}
	a.invalidate(ctx, userID)
	httpx.OK(w, http.StatusOK, MsgAccountLocked, nil)
}

// RevokeSessions answers POST /auth/sessions/revoke: every other device is
// signed out; the caller gets a new token and stays signed in.
func (a *Auth) RevokeSessions(w http.ResponseWriter, r *http.Request) {
	ctx, userID := r.Context(), CurrentUser(r.Context()).ID
	ver, err := a.Accounts.BumpTokenVersion(ctx, userID)
	if err != nil {
		httpx.ServerError(w, "sign out other devices", err)
		return
	}
	a.invalidate(ctx, userID)
	a.signIn(w, r, userID, ver, CurrentSession(r.Context()), a.TokenTTL, MsgSessionsRevoked)
}
