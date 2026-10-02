package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/server"
)

// fakeStore stands in for the database: one user who is a member of "acme".
type fakeStore struct {
	user     auth.User
	disabled bool
}

func (f *fakeStore) UserByEmail(_ context.Context, email string) (auth.User, error) {
	if f.disabled || email != f.user.Email {
		return auth.User{}, auth.ErrNotFound
	}
	return f.user, nil
}

func (f *fakeStore) ActiveUser(_ context.Context, id string) (auth.User, error) {
	if f.disabled || id != f.user.ID {
		return auth.User{}, auth.ErrNotFound
	}
	return f.user, nil
}

func (f *fakeStore) Scope(_ context.Context, _, slug string) (auth.Scope, error) {
	if slug != "acme" {
		return auth.Scope{}, auth.ErrNotFound
	}
	return auth.Scope{
		Organisation: &auth.Organisation{Slug: "acme"},
		Role:         "member",
		Permissions:  []string{"passport:issue"},
	}, nil
}

func newAuth(t *testing.T) (*auth.Auth, *fakeStore) {
	t.Helper()
	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{user: auth.User{ID: "user-1", Email: "ada@acme.test", Name: "Ada", PasswordHash: hash}}
	return &auth.Auth{
		Store:      store,
		Secret:     []byte("test-secret-test-secret-test-secret"),
		TokenTTL:   time.Hour,
		BaseDomain: "chaintrack.test",
	}, store
}

func noPing(context.Context) error { return nil }

func do(h http.Handler, method, url, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestToken(t *testing.T) {
	a, _ := newAuth(t)
	now := time.Now()

	token, err := a.IssueToken("user-1", now)
	if err != nil {
		t.Fatal(err)
	}
	if id, err := a.ParseToken(token); err != nil || id != "user-1" {
		t.Fatalf("round trip: id=%q err=%v", id, err)
	}

	if _, err := a.ParseToken(token[:len(token)-2] + "xx"); err == nil {
		t.Fatal("tampered token accepted")
	}

	expired, _ := a.IssueToken("user-1", now.Add(-2*time.Hour))
	if _, err := a.ParseToken(expired); err == nil {
		t.Fatal("expired token accepted")
	}

	other := *a
	other.Secret = []byte("another-secret-another-secret-123")
	forged, _ := other.IssueToken("user-1", now)
	if _, err := a.ParseToken(forged); err == nil {
		t.Fatal("token signed with another secret accepted")
	}

	// {"alg":"none"} header: an unsigned token must never be trusted.
	if _, err := a.ParseToken("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEiLCJleHAiOjk5OTk5OTk5OTl9."); err == nil {
		t.Fatal("unsigned token accepted")
	}
}

func TestSlug(t *testing.T) {
	cases := []struct {
		name string
		url  string
		dev  bool
		want string
	}{
		{"subdomain", "http://acme.chaintrack.test/me", false, "acme"},
		{"subdomain with port", "http://acme.chaintrack.test:8080/me", false, "acme"},
		{"upper case host", "http://ACME.chaintrack.test/me", false, "acme"},
		{"bare domain", "http://chaintrack.test/me", false, ""},
		{"other domain", "http://acme.evil.test/me", false, ""},
		{"nested subdomain", "http://a.b.chaintrack.test/me", false, ""},
		{"reserved", "http://admin.chaintrack.test/me", false, ""},
		{"pslug in dev", "http://localhost:8080/me?pslug=acme", true, "acme"},
		{"pslug ignored outside dev", "http://localhost:8080/me?pslug=acme", false, ""},
		{"pslug wins over subdomain in dev", "http://other.chaintrack.test/me?pslug=acme", true, "acme"},
		{"invalid pslug", "http://localhost:8080/me?pslug=Acme_Inc", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := newAuth(t)
			a.Dev = tc.dev
			if got := a.Slug(httptest.NewRequest(http.MethodGet, tc.url, nil)); got != tc.want {
				t.Fatalf("slug = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	a, store := newAuth(t)
	h := server.NewRouter(noPing, a)

	ok := do(h, http.MethodPost, "/auth/login", "", `{"email":"ada@acme.test","password":"correct-password"}`)
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

	wrongPass := do(h, http.MethodPost, "/auth/login", "", `{"email":"ada@acme.test","password":"nope"}`)
	wrongEmail := do(h, http.MethodPost, "/auth/login", "", `{"email":"nobody@acme.test","password":"correct-password"}`)
	if wrongPass.Code != http.StatusUnauthorized || wrongEmail.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d and %d, want 401", wrongPass.Code, wrongEmail.Code)
	}
	if wrongPass.Body.String() != wrongEmail.Body.String() {
		t.Fatal("wrong email and wrong password are distinguishable")
	}

	store.disabled = true
	if rec := do(h, http.MethodPost, "/auth/login", "", `{"email":"ada@acme.test","password":"correct-password"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user login status = %d, want 401", rec.Code)
	}
}

func TestAccessControl(t *testing.T) {
	a, store := newAuth(t)
	token, _ := a.IssueToken("user-1", time.Now())

	// A route for each permission outcome, wired the way real routes will be.
	r := chi.NewRouter()
	r.Use(a.Authenticate, a.ResolveOrg)
	okHandler := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	r.With(a.Require("passport:issue")).Get("/issue", okHandler)
	r.With(a.Require("org:verify")).Get("/verify", okHandler)

	cases := []struct {
		name  string
		url   string
		token string
		want  int
	}{
		{"no token", "http://acme.chaintrack.test/issue", "", http.StatusUnauthorized},
		{"bad token", "http://acme.chaintrack.test/issue", "not-a-token", http.StatusUnauthorized},
		{"member with permission", "http://acme.chaintrack.test/issue", token, http.StatusOK},
		{"member without permission", "http://acme.chaintrack.test/verify", token, http.StatusForbidden},
		{"organisation user is not in", "http://rival.chaintrack.test/issue", token, http.StatusForbidden},
		{"no organisation named", "http://chaintrack.test/issue", token, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if rec := do(r, http.MethodGet, tc.url, tc.token, ""); rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body)
			}
		})
	}

	// Disabling the account must lock out a token that is still unexpired.
	store.disabled = true
	if rec := do(r, http.MethodGet, "http://acme.chaintrack.test/issue", token, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user status = %d, want 401", rec.Code)
	}
}
