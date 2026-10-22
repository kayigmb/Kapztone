package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
)

// mailbox captures the account emails (sent in the background) and hands
// back the token of the link each one carries.
type mailbox chan string

var linkToken = regexp.MustCompile(`/(?:reset|lock)/([A-Za-z0-9_-]+)`)

func captureMail(e *env) mailbox {
	box := make(mailbox, 8)
	e.a.Mail = func(to, subject, text, _ string) error {
		box <- subject + "\n" + text
		return nil
	}
	return box
}

func (m mailbox) next(t *testing.T, subject string) (token string) {
	t.Helper()
	select {
	case mail := <-m:
		if !strings.HasPrefix(mail, subject) {
			t.Fatalf("mail %q, want subject %q", mail, subject)
		}
		match := linkToken.FindStringSubmatch(mail)
		if match == nil {
			t.Fatalf("mail without a link: %q", mail)
		}
		// Never the address the request claimed (see the forgot request below).
		if !strings.Contains(mail, testAppURL+"/") {
			t.Fatalf("link not on %s: %q", testAppURL, mail)
		}
		return match[1]
	case <-time.After(5 * time.Second):
		t.Fatalf("no %q mail", subject)
	}
	return ""
}

func (m mailbox) empty(t *testing.T) {
	t.Helper()
	select {
	case mail := <-m:
		t.Fatalf("unexpected mail: %q", mail)
	case <-time.After(200 * time.Millisecond):
	}
}

func meCode(e *env, token string) int { return do(e.h, http.MethodGet, "/me", token, "").Code }

func TestPasswordResetAndLock(t *testing.T) {
	e, _, _ := newEnv(t, false)
	box := captureMail(e)
	ctx := context.Background()
	userID := e.addUser(tStaff, auth.RoleMember, "")
	before := e.login(tStaff)

	resets := func() (n int) {
		e.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets WHERE user_id = $1::uuid`, userID).Scan(&n)
		return n
	}
	var allBefore, allAfter int
	e.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets`).Scan(&allBefore)

	// An unknown email gets the same answer and nothing is stored or sent.
	want(t, "forgot, unknown email", do(e.h, http.MethodPost, "/auth/password/forgot", "",
		`{"email":"nobody-`+e.suffix+`@chaintrack.test"}`), http.StatusAccepted)
	box.empty(t)
	e.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets`).Scan(&allAfter)
	if allAfter != allBefore {
		t.Fatal("a reset link was stored for an unknown email")
	}

	// Another site's page is refused outright.
	forgot := func(host string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/password/forgot",
			strings.NewReader(`{"email":"`+strings.ToUpper(e.email(tStaff))+`"}`))
		req.Host = host
		req.Header.Set(auth.HeaderOrigin, "https://evil.test")
		rec := httptest.NewRecorder()
		e.h.ServeHTTP(rec, req)
		return rec
	}
	want(t, "forgot from another site", forgot("example.com"), http.StatusForbidden)
	// A script forging Host and Origin alike still cannot aim the link elsewhere.
	want(t, "forgot", forgot("evil.test"), http.StatusAccepted)
	token := box.next(t, auth.MailResetSubject)
	if resets() != 1 {
		t.Fatal("reset link not stored")
	}
	var stored int
	e.pool.QueryRow(ctx, `SELECT count(*) FROM password_resets WHERE token_hash = $1`, token).Scan(&stored)
	if stored != 0 {
		t.Fatal("reset token stored in the clear")
	}

	reset := func(tok, pw string) int {
		return do(e.h, http.MethodPost, "/auth/password/reset", "", `{"token":"`+tok+`","new":"`+pw+`"}`).Code
	}
	const newPassword = "Harbour-lantern-41!"
	if got := reset(token, "short"); got != http.StatusBadRequest {
		t.Fatalf("short password = %d", got)
	}
	if got := reset(token, "lowercase-password1"); got != http.StatusBadRequest {
		t.Fatalf("weak password = %d", got)
	}
	if got := reset(token, testPassword); got != http.StatusBadRequest {
		t.Fatalf("current password reused = %d", got)
	}
	if got := reset("not-a-token", newPassword); got != http.StatusNotFound {
		t.Fatalf("unknown token = %d", got)
	}
	// The refused attempts left the link usable.
	if got := reset(token, newPassword); got != http.StatusOK {
		t.Fatalf("reset = %d", got)
	}
	if got := reset(token, "Another-good-pass1!"); got != http.StatusNotFound {
		t.Fatalf("link used twice = %d", got)
	}
	if got := meCode(e, before); got != http.StatusUnauthorized {
		t.Fatalf("a token from before the reset still works: %d", got)
	}
	if rec := do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tStaff), newPassword)); rec.Code != http.StatusOK {
		t.Fatalf("sign in with the new password = %d", rec.Code)
	}
	lock := box.next(t, auth.MailChangedSubject)

	// An expired link is refused.
	expired := "expired-" + e.suffix
	if err := e.a.Accounts.AddLink(ctx, userID, auth.LinkReset, sha256Hex(expired), time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if got := reset(expired, "Yet-another-pass1!"); got != http.StatusNotFound {
		t.Fatalf("expired link = %d", got)
	}

	// "Lock my account" disables it and signs everyone out.
	after := decodeTo[struct{ Token string }](t,
		do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tStaff), newPassword))).Token
	want(t, "lock", do(e.h, http.MethodPost, "/auth/lock", "", `{"token":"`+lock+`"}`), http.StatusOK)
	want(t, "lock twice", do(e.h, http.MethodPost, "/auth/lock", "", `{"token":"`+lock+`"}`), http.StatusNotFound)
	if got := meCode(e, after); got != http.StatusUnauthorized {
		t.Fatalf("locked account still signed in: %d", got)
	}
	if rec := do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tStaff), newPassword)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("locked account signed in: %d", rec.Code)
	}
}

func TestPasswordChangeHistoryAndSessions(t *testing.T) {
	e, _, _ := newEnv(t, false)
	box := captureMail(e)
	e.addUser(tStaff, auth.RoleMember, "")
	first := e.login(tStaff)
	second := e.login(tStaff)

	change := func(tok, cur, next string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodPost, "/auth/password", tok, `{"current":"`+cur+`","new":"`+next+`"}`)
	}
	rec := change(first, testPassword, "Harbour-lantern-41!")
	want(t, "change", rec, http.StatusOK)
	fresh := decodeTo[struct{ Token string }](t, rec).Token
	box.next(t, auth.MailChangedSubject)
	if meCode(e, first) != http.StatusUnauthorized || meCode(e, second) != http.StatusUnauthorized {
		t.Fatal("other sessions survived a password change")
	}
	if meCode(e, fresh) != http.StatusOK {
		t.Fatal("the session that changed the password was signed out")
	}
	// The last passwords can't come back.
	want(t, "back to the old password", change(fresh, "Harbour-lantern-41!", testPassword), http.StatusBadRequest)

	// Sign out other devices: the caller stays signed in with a new token.
	other := decodeTo[struct{ Token string }](t,
		do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tStaff), "harbour-lantern-41"))).Token
	rec = do(e.h, http.MethodPost, "/auth/sessions/revoke", fresh, "")
	want(t, "revoke", rec, http.StatusOK)
	kept := decodeTo[struct{ Token string }](t, rec).Token
	if meCode(e, other) != http.StatusUnauthorized || meCode(e, fresh) != http.StatusUnauthorized {
		t.Fatal("old sessions survived sign-out-everywhere")
	}
	if meCode(e, kept) != http.StatusOK {
		t.Fatal("the caller was signed out")
	}
}

func TestAttemptsLeftAndRemember(t *testing.T) {
	e, _, cached := newEnv(t, true)
	e.addUser(tStaff, auth.RoleMember, "")
	e.a.Limiter = &auth.Limiter{Cache: cached.Cache}

	for want := auth.LoginRate.Max - 1; want >= auth.LoginRate.Max-2; want-- {
		rec := do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tStaff), "wrong-password-x"))
		got := decodeTo[struct {
			AttemptsLeft int64 `json:"attempts_left"`
		}](t, rec).AttemptsLeft
		if rec.Code != http.StatusUnauthorized || got != want {
			t.Fatalf("attempts_left = %d (%d), want %d", got, rec.Code, want)
		}
	}

	rec := do(e.h, http.MethodPost, "/auth/login", "",
		`{"email":"`+e.email(tStaff)+`","password":"`+testPassword+`","remember":true}`)
	want(t, "remembered sign-in", rec, http.StatusOK)
	if got := decodeTo[struct {
		ExpiresIn int `json:"expires_in"`
	}](t, rec).ExpiresIn; got != int(e.a.RememberTTL.Seconds()) {
		t.Fatalf("remembered sign-in lasts %d s", got)
	}
	if c := cookieNamed(rec, auth.SessionCookie); c == nil || c.MaxAge != int(e.a.RememberTTL.Seconds()) {
		t.Fatalf("remembered cookie: %+v", c)
	}
}

func TestTrustedDevice(t *testing.T) {
	e, _, _ := newEnv(t, false)
	now := time.Now()
	e.a.Now = func() time.Time { return now }
	e.addUser(tStaff, auth.RoleMember, "")
	tok := e.login(tStaff)

	rec := do(e.h, http.MethodPost, "/auth/2fa/setup", tok, "")
	want(t, "setup", rec, http.StatusOK)
	secret := decodeTo[struct{ Secret string }](t, rec).Secret
	code := func() string {
		now = now.Add(30 * time.Second) // each code works once: move to a fresh step
		c, err := auth.TOTPCode(secret, now)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	want(t, "enable", do(e.h, http.MethodPost, "/auth/2fa/enable", tok, `{"code":"`+code()+`"}`), http.StatusOK)

	login := func(body string, trust *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		if trust != nil {
			req.AddCookie(trust)
		}
		return send(e.h, req, "")
	}
	noCode := `{"email":"` + e.email(tStaff) + `","password":"` + testPassword + `"}`
	if rec := login(noCode, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("2FA skipped without trust: %d", rec.Code)
	}
	rec = login(`{"email":"`+e.email(tStaff)+`","password":"`+testPassword+`","code":"`+code()+`","trust_device":true}`, nil)
	want(t, "sign in and trust", rec, http.StatusOK)
	trust := cookieNamed(rec, auth.TrustCookie)
	if trust == nil || !trust.HttpOnly || trust.MaxAge != int(auth.TrustTTL.Seconds()) {
		t.Fatalf("trust cookie: %+v", trust)
	}
	signedIn := decodeTo[struct{ Token string }](t, rec).Token
	want(t, "trusted device skips the code", login(noCode, trust), http.StatusOK)

	forged := *trust
	forged.Value = trust.Value[:len(trust.Value)-3] + "abc"
	if rec := login(noCode, &forged); rec.Code != http.StatusUnauthorized {
		t.Fatalf("forged trust cookie accepted: %d", rec.Code)
	}

	// Signing out everywhere bumps the token version: the device is no longer trusted.
	want(t, "revoke", do(e.h, http.MethodPost, "/auth/sessions/revoke", signedIn, ""), http.StatusOK)
	if rec := login(noCode, trust); rec.Code != http.StatusUnauthorized {
		t.Fatalf("trust survived sign-out-everywhere: %d", rec.Code)
	}

	// After its 30 days the cookie no longer counts.
	rec = login(`{"email":"`+e.email(tStaff)+`","password":"`+testPassword+`","code":"`+code()+`","trust_device":true}`, nil)
	trust = cookieNamed(rec, auth.TrustCookie)
	now = now.Add(auth.TrustTTL + time.Minute)
	if rec := login(noCode, trust); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expired trust cookie accepted: %d", rec.Code)
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func cookieNamed(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
