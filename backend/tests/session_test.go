package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
)

const (
	siteAcme  = "acme." + testBaseDomain
	siteRival = "rival." + testBaseDomain
	apiInside = "api.internal:8080" // the API's own address behind the web app
)

// browser sends a request the way a browser does: cookie, and headers the test names.
func browser(h http.Handler, method, url string, session *http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(""))
	if session != nil {
		req.AddCookie(session)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == auth.SessionCookie {
			return c
		}
	}
	t.Fatalf("no %s cookie in the answer", auth.SessionCookie)
	return nil
}

// The session lives in a cookie the API sets: scripts cannot read it, it is
// https-only outside dev, and only the site's own pages may use it to change anything.
func TestSessionCookie(t *testing.T) {
	a, _ := newAuth(t)
	h := fakeRouter(a)
	signIn := func() *http.Cookie {
		t.Helper()
		rec := do(h, http.MethodPost, "/auth/login", "", loginBody(fakeEmail, testPassword))
		want(t, "sign in", rec, http.StatusOK)
		return sessionCookie(t, rec)
	}

	c := signIn()
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.MaxAge <= 0 || c.Domain != "" {
		t.Fatalf("session cookie in production: %+v", c)
	}
	a.Dev = true
	if dev := signIn(); !dev.HttpOnly || dev.Secure {
		t.Fatalf("session cookie in dev must stay HttpOnly and drop Secure: %+v", dev)
	}
	a.Dev = false

	// ---- The cookie signs the browser in; the site's address names the organisation.
	me := func(url string, headers map[string]string) meResponse {
		t.Helper()
		rec := browser(h, http.MethodGet, url, c, headers)
		want(t, "GET "+url, rec, http.StatusOK)
		return decodeTo[meResponse](t, rec)
	}
	if got := me("http://"+siteAcme+"/me", nil); got.Organisation == nil || got.Organisation.Slug != fakeSlug {
		t.Fatalf("cookie on the organisation's address: %+v", got)
	}
	// Behind the web app the API sees its own address; the forwarded host says which site it is.
	forwarded := map[string]string{auth.HeaderForwardedHost: siteAcme}
	if got := me("http://"+apiInside+"/me", forwarded); got.Organisation == nil || got.Organisation.Slug != fakeSlug {
		t.Fatalf("forwarded host did not select the organisation: %+v", got)
	}
	if got := me("http://"+apiInside+"/me", nil); got.Organisation != nil {
		t.Fatalf("the API's own address named an organisation: %+v", got)
	}
	want(t, "no cookie", browser(h, http.MethodGet, "http://"+siteAcme+"/me", nil, nil), http.StatusUnauthorized)
	want(t, "forged cookie", browser(h, http.MethodGet, "http://"+siteAcme+"/me",
		&http.Cookie{Name: auth.SessionCookie, Value: c.Value + "x"}, nil), http.StatusUnauthorized)

	// ---- Changing something with the cookie: only from the site's own pages.
	const change = "/auth/2fa/setup"
	post := func(url string, headers map[string]string) int {
		return browser(h, http.MethodPost, url, c, headers).Code
	}
	for name, headers := range map[string]map[string]string{
		"no origin":                     nil,
		"another site":                  {auth.HeaderOrigin: "https://evil.example"},
		"another organisation's site":   {auth.HeaderOrigin: "http://" + siteRival},
		"right name, wrong port":        {auth.HeaderOrigin: "http://" + siteAcme + ":9999"},
		"forwarded host of another one": {auth.HeaderOrigin: "http://" + siteRival, auth.HeaderForwardedHost: siteAcme},
	} {
		if code := post("http://"+siteAcme+change, headers); code != http.StatusForbidden {
			t.Errorf("POST with the cookie, %s = %d, want 403", name, code)
		}
	}
	if code := post("http://"+siteAcme+change, map[string]string{auth.HeaderOrigin: "http://" + siteAcme}); code != http.StatusOK {
		t.Fatalf("POST from the site's own page = %d, want 200", code)
	}
	if code := post("http://"+apiInside+change, map[string]string{auth.HeaderOrigin: "https://" + siteAcme, auth.HeaderForwardedHost: siteAcme}); code != http.StatusOK {
		t.Fatalf("POST from the site's own page, through the web app = %d, want 200", code)
	}

	// A token in the header is not attached by browsers on their own, so no origin is asked of it.
	req := httptest.NewRequest(http.MethodPost, "http://"+siteAcme+change, nil)
	req.Header.Set(auth.HeaderOrigin, "https://evil.example")
	want(t, "bearer token from anywhere", send(h, req, c.Value), http.StatusOK)

	// ---- Signing out: the API takes the cookie back.
	rec := browser(h, http.MethodPost, "http://"+siteAcme+"/auth/logout", c, nil)
	want(t, "sign out", rec, http.StatusOK)
	if gone := sessionCookie(t, rec); gone.MaxAge >= 0 || gone.Value != "" || !gone.HttpOnly {
		t.Fatalf("sign-out cookie: %+v", gone)
	}
}
