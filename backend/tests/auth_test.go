package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/server"
)

func TestHealthz(t *testing.T) {
	a, _ := newAuth(t)
	cases := []struct {
		name     string
		pingErr  error
		wantCode int
		want     string
	}{
		{"db up", nil, http.StatusOK, server.HealthOK},
		{"db down", errors.New("connection refused"), http.StatusServiceUnavailable, server.HealthDBDown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ping := func(context.Context) error { return tc.pingErr }
			rec := do(server.NewRouter(deps(ping, a)), http.MethodGet, "/healthz", "", "")
			var got struct{ Status string }
			json.Unmarshal(rec.Body.Bytes(), &got)
			if rec.Code != tc.wantCode || got.Status != tc.want {
				t.Fatalf("got %d %q, want %d %q", rec.Code, got.Status, tc.wantCode, tc.want)
			}
		})
	}
}

func TestToken(t *testing.T) {
	a, _ := newAuth(t)
	now := time.Now()

	token, err := a.IssueToken(fakeUserID, now)
	if err != nil {
		t.Fatal(err)
	}
	if id, err := a.ParseToken(token); err != nil || id != fakeUserID {
		t.Fatalf("round trip: id=%q err=%v", id, err)
	}

	if _, err := a.ParseToken(token[:len(token)-2] + "xx"); err == nil {
		t.Fatal("tampered token accepted")
	}

	expired, _ := a.IssueToken(fakeUserID, now.Add(-2*a.TokenTTL))
	if _, err := a.ParseToken(expired); err == nil {
		t.Fatal("expired token accepted")
	}

	other := *a
	other.Secret = []byte("another-secret-another-secret-123")
	forged, _ := other.IssueToken(fakeUserID, now)
	if _, err := a.ParseToken(forged); err == nil {
		t.Fatal("token signed with another secret accepted")
	}

	// {"alg":"none"} header: an unsigned token must never be trusted.
	if _, err := a.ParseToken("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEiLCJleHAiOjk5OTk5OTk5OTl9."); err == nil {
		t.Fatal("unsigned token accepted")
	}
}

func TestSlug(t *testing.T) {
	const apiURL = "http://api.internal:8080/me"
	withParam := apiURL + "?" + auth.ParamOrgSlug + "="
	site := func(host string) string { return "https://" + host }

	cases := []struct {
		name   string
		url    string
		origin string // the browser's Origin header: the frontend's address
		dev    bool
		want   string
	}{
		{"frontend subdomain", apiURL, site(fakeSlug + "." + testBaseDomain), false, fakeSlug},
		{"frontend with port", apiURL, site(fakeSlug + ".localhost:3000"), false, fakeSlug},
		{"upper case frontend", apiURL, site(strings.ToUpper(fakeSlug) + "." + testBaseDomain), false, fakeSlug},
		{"frontend wins over the API's own host", orgURL("other", "/me"), site(fakeSlug + "." + testBaseDomain), false, fakeSlug},
		{"no Origin: the request's host is used", orgURL(fakeSlug, "/me"), "", false, fakeSlug},
		{"host with no subdomain", "http://localhost:8080/me", "", false, ""},
		{"IP address", "http://127.0.0.1:8080/me", "", false, ""},
		{"reserved name", apiURL, site(auth.ReservedSlugs[0] + "." + testBaseDomain), false, ""},
		{"not a valid slug", apiURL, site("bad_name." + testBaseDomain), false, ""},
		{"malformed Origin falls back to host", orgURL(fakeSlug, "/me"), "null", false, fakeSlug},
		{"param in dev", withParam + fakeSlug, "", true, fakeSlug},
		{"param ignored outside dev", "http://localhost:8080/me?" + auth.ParamOrgSlug + "=" + fakeSlug, "", false, ""},
		{"param wins over the frontend in dev", withParam + fakeSlug, site("other." + testBaseDomain), true, fakeSlug},
		{"invalid param", "http://localhost:8080/me?" + auth.ParamOrgSlug + "=Acme_Inc", "", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := newAuth(t)
			a.Dev = tc.dev
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			if tc.origin != "" {
				req.Header.Set(auth.HeaderOrigin, tc.origin)
			}
			if got := a.Slug(req); got != tc.want {
				t.Fatalf("slug = %q, want %q", got, tc.want)
			}
		})
	}
}

// The frontend's address is recorded for every request and readable anywhere.
func TestFrontendIsKnownToHandlers(t *testing.T) {
	var got auth.Frontend
	h := auth.WithFrontend(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = auth.FrontendOf(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "http://api.internal:8080/x", nil)
	req.Header.Set(auth.HeaderOrigin, "https://Acme.chaintrack.app:8443")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got.Host != "acme.chaintrack.app" || got.Origin != "https://Acme.chaintrack.app:8443" {
		t.Fatalf("from Origin: %+v", got)
	}

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://acme.localhost:8080/x", nil))
	if got.Host != "acme.localhost" || got.Origin != "http://acme.localhost:8080" {
		t.Fatalf("from Host: %+v", got)
	}
}

func TestGate(t *testing.T) {
	cases := []struct {
		name   string
		scope  auth.Scope
		wantID string
		wantOK bool
	}{
		{"platform sees all", auth.Scope{Platform: true}, "", true},
		{"organisation sees its own", auth.Scope{Organisation: &auth.Organisation{ID: fakeOrgID}}, fakeOrgID, true},
		{"no organisation sees nothing", auth.Scope{}, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if id, ok := tc.scope.Gate(); id != tc.wantID || ok != tc.wantOK {
				t.Fatalf("gate = %q %v, want %q %v", id, ok, tc.wantID, tc.wantOK)
			}
		})
	}
}

func loginBody(email, password string) string {
	return fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
}

func TestLogin(t *testing.T) {
	a, store := newAuth(t)
	h := fakeRouter(a)

	ok := do(h, http.MethodPost, "/auth/login", "", loginBody(fakeEmail, testPassword))
	if ok.Code != http.StatusOK {
		t.Fatalf("login status = %d, body %s", ok.Code, ok.Body)
	}
	var out struct{ Token string }
	if err := json.Unmarshal(ok.Body.Bytes(), &out); err != nil || out.Token == "" {
		t.Fatalf("no token in %s", ok.Body)
	}
	if strings.Contains(ok.Body.String(), "password") {
		t.Fatal("login response leaks password data")
	}

	wrongPass := do(h, http.MethodPost, "/auth/login", "", loginBody(fakeEmail, "nope"))
	wrongEmail := do(h, http.MethodPost, "/auth/login", "", loginBody("nobody@acme.test", testPassword))
	if wrongPass.Code != http.StatusUnauthorized || wrongEmail.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d and %d, want 401", wrongPass.Code, wrongEmail.Code)
	}
	if wrongPass.Body.String() != wrongEmail.Body.String() {
		t.Fatal("wrong email and wrong password are distinguishable")
	}

	store.disabled = true
	if rec := do(h, http.MethodPost, "/auth/login", "", loginBody(fakeEmail, testPassword)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user login status = %d, want 401", rec.Code)
	}
}

func TestAccessControl(t *testing.T) {
	a, store := newAuth(t)
	token, _ := a.IssueToken(fakeUserID, time.Now())

	// A route for each permission outcome, wired the way real routes are.
	r := chi.NewRouter()
	r.Use(a.Authenticate, a.ResolveOrg)
	okHandler := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	r.With(a.Require(auth.PassportIssue)).Get("/issue", okHandler)
	r.With(a.Require(auth.OrgVerify)).Get("/verify", okHandler)

	cases := []struct {
		name  string
		url   string
		token string
		want  int
	}{
		{"no token", orgURL(fakeSlug, "/issue"), "", http.StatusUnauthorized},
		{"bad token", orgURL(fakeSlug, "/issue"), "not-a-token", http.StatusUnauthorized},
		{"member with permission", orgURL(fakeSlug, "/issue"), token, http.StatusOK},
		{"member without permission", orgURL(fakeSlug, "/verify"), token, http.StatusForbidden},
		{"organisation user is not in", orgURL("rival", "/issue"), token, http.StatusForbidden},
		{"no organisation named", "http://" + testBaseDomain + "/issue", token, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(r, http.MethodGet, tc.url, tc.token, ""); rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body)
			}
		})
	}

	// Naming an organisation the user is not in is not an error by itself:
	// they are signed in, acting for nobody, with no permissions.
	me := do(fakeRouter(a), http.MethodGet, orgURL("rival", "/me"), token, "")
	var who struct {
		Organisation *auth.Organisation
		Permissions  []string
	}
	json.Unmarshal(me.Body.Bytes(), &who)
	if me.Code != http.StatusOK || who.Organisation != nil || len(who.Permissions) != 0 {
		t.Fatalf("foreign organisation gave %d %s", me.Code, me.Body)
	}

	// Disabling the account must lock out a token that is still unexpired.
	store.disabled = true
	if rec := do(r, http.MethodGet, orgURL(fakeSlug, "/issue"), token, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user status = %d, want 401", rec.Code)
	}
}

func TestPermissionsAreDeclaredOnce(t *testing.T) {
	format := regexp.MustCompile(`^[a-z_]+:[a-z_]+$`)
	seen := map[string]bool{}
	for _, p := range auth.AllPermissions {
		if !format.MatchString(p) {
			t.Errorf("%q is not resource:action", p)
		}
		if seen[p] {
			t.Errorf("%q listed twice", p)
		}
		seen[p] = true
	}
	for role, perms := range auth.DefaultRoles {
		for _, p := range perms {
			if !seen[p] {
				t.Errorf("role %s uses undeclared permission %q", role, p)
			}
		}
	}
	if len(auth.DefaultRoles[auth.RoleSuperadmin]) != len(auth.AllPermissions) {
		t.Error("superadmin must hold every permission")
	}
}

// Every endpoint must state who may call it. Adding a route without a
// permission, or a new public route, has to be a deliberate change here.
func TestEveryRouteIsGated(t *testing.T) {
	a, _ := newAuth(t)
	token, _ := a.IssueToken(fakeUserID, time.Now())
	h := fakeRouter(a)

	public := []string{"GET /healthz", "POST /auth/login", "POST /orgs", "GET /verify/{id}", "POST /auth/logout",
		"POST /invites/{token}/accept", "POST /auth/password/forgot", "POST /auth/password/reset", "POST /auth/lock"}
	signedIn := []string{"GET /me", "POST /auth/2fa/setup", "POST /auth/2fa/enable", "POST /auth/2fa/disable",
		"POST /auth/2fa/recovery", "POST /auth/pin", "GET /notifications", "POST /notifications/{id}/read",
		"PUT /me", "GET /me/organisations", "POST /auth/password", "POST /auth/sessions/revoke"}
	idInPath := regexp.MustCompile(`\{[^}]+\}`)

	for _, rt := range server.Routes(deps(noPing, a)) {
		name := rt.Method + " " + rt.Path
		url := orgURL(fakeSlug, idInPath.ReplaceAllString(rt.Path, fakeOrgID))

		switch rt.Access {
		case server.Public:
			if !slices.Contains(public, name) {
				t.Errorf("%s is public but not on the public list", name)
			}
			continue
		case server.SignedIn:
			if !slices.Contains(signedIn, name) {
				t.Errorf("%s needs no permission but is not on the signed-in list", name)
			}
		case server.Permission, server.VerifiedPermission:
			if !slices.Contains(auth.AllPermissions, rt.Permission) {
				t.Errorf("%s has unknown permission %q", name, rt.Permission)
			}
			// The fake user holds none of the permissions these routes need.
			if rec := do(h, rt.Method, url, token, ""); rec.Code != http.StatusForbidden {
				t.Errorf("%s without its permission = %d, want 403", name, rec.Code)
			}
		}
		if rec := do(h, rt.Method, url, "", ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s without sign-in = %d, want 401", name, rec.Code)
		}
	}
}
