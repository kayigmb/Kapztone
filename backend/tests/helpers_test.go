package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
	"github.com/kayigmb/chaintrack/backend/internal/server"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
)

const (
	testSecret   = "test-secret-test-secret-test-secret"
	testPassword = "Correct-password1!"
	// testPIN is every database test user's signing PIN; send attaches it to every request.
	testPIN        = "246810"
	testBaseDomain = "chaintrack.test"
	testKey        = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f" // 32 bytes, hex
	fakeUserID     = "user-1"
	fakeEmail      = "ada@acme.test"
	fakeSlug       = "acme"
	fakeOrgID      = "11111111-1111-1111-1111-111111111111"
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

func (f *fakeStore) SetTOTP(context.Context, string, string, bool) error { return nil }

func (f *fakeStore) UseTOTPStep(context.Context, string, int64) (bool, error) { return true, nil }

func (f *fakeStore) PINHash(context.Context, string) (string, error)               { return "", nil }
func (f *fakeStore) SetPIN(context.Context, string, string) error                  { return nil }
func (f *fakeStore) SetRecoveryCodes(context.Context, string, []string) error      { return nil }
func (f *fakeStore) UseRecoveryCode(context.Context, string, string) (bool, error) { return false, nil }

func (f *fakeStore) Scope(_ context.Context, _, slug string) (auth.Scope, error) {
	if slug != fakeSlug {
		return auth.Scope{}, auth.ErrNotFound
	}
	return auth.Scope{
		Organisation: &auth.Organisation{ID: fakeOrgID, Slug: fakeSlug},
		Role:         auth.RoleMember,
		Permissions:  []string{auth.PassportIssue},
	}, nil
}

func newAuth(t *testing.T) (*auth.Auth, *fakeStore) {
	t.Helper()
	hash, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{user: auth.User{ID: fakeUserID, Email: fakeEmail, Name: "Ada", PasswordHash: hash}}
	return &auth.Auth{
		Store:    store,
		Secret:   []byte(testSecret),
		TokenTTL: time.Hour,
		Secrets:  testBox(t),
	}, store
}

func noPing(context.Context) error { return nil }

// deps wires the real route table around a; the feature handlers are present
// but have no database behind them.
func deps(ping func(context.Context) error, a *auth.Auth) server.Deps {
	return server.Deps{
		Ping: ping, Auth: a,
		Org: &org.Handlers{}, Facility: &facility.Handlers{}, Template: &template.Handlers{},
		Passport: &passport.Handlers{}, Attestation: &attestation.Handlers{}, Flag: &flag.Handlers{},
		Notify: &notify.Handlers{}, Verify: &verify.Handlers{},
	}
}

// fakeRouter is the real router over the fake store.
func fakeRouter(a *auth.Auth) http.Handler { return server.NewRouter(deps(noPing, a)) }

func orgURL(slug, path string) string { return "http://" + slug + "." + testBaseDomain + path }

func do(h http.Handler, method, url, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	return send(h, req, token)
}

func send(h http.Handler, req *http.Request, token string) *httptest.ResponseRecorder {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if req.Header.Get(auth.HeaderSigningPIN) == "" {
		req.Header.Set(auth.HeaderSigningPIN, testPIN)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func testBox(t *testing.T) *secret.Box {
	t.Helper()
	box, err := secret.New(testKey)
	if err != nil {
		t.Fatal(err)
	}
	return box
}
