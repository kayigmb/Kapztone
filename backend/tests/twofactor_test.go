package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
)

func TestSecretBox(t *testing.T) {
	box := testBox(t)
	const plain = "JBSWY3DPEHPK3PXP"

	sealed, err := box.Seal(plain)
	if err != nil || strings.Contains(sealed, plain) {
		t.Fatalf("seal failed or leaked the value: %q %v", sealed, err)
	}
	if again, _ := box.Seal(plain); again == sealed {
		t.Fatal("sealing the same value twice gave the same output")
	}
	if got, err := box.Open(sealed); err != nil || got != plain {
		t.Fatalf("round trip: %q %v", got, err)
	}

	// Any change to the stored value must be detected.
	tampered := []byte(sealed)
	tampered[len(tampered)-3] ^= 1
	if _, err := box.Open(string(tampered)); err == nil {
		t.Fatal("tampered value opened")
	}
	if _, err := box.Open("not base64!"); err == nil {
		t.Fatal("malformed value opened")
	}
	other, _ := secret.New(strings.Repeat("ab", secret.KeyBytes))
	if _, err := other.Open(sealed); err == nil {
		t.Fatal("value opened with a different key")
	}

	for _, bad := range []string{"", "abcd", strings.Repeat("zz", secret.KeyBytes), strings.Repeat("ab", secret.KeyBytes-1)} {
		if _, err := secret.New(bad); err == nil {
			t.Errorf("key %q accepted", bad)
		}
	}
}

func codeBody(code string) string { return fmt.Sprintf(`{"code":%q}`, code) }

func loginWithCode(email, code string) string {
	return fmt.Sprintf(`{"email":%q,"password":%q,"code":%q}`, email, testPassword, code)
}

func TestTwoFactorFlow(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.register(tAcme, org.KindProducer)
	email := e.email(tAcme)
	token := e.login(tAcme)
	post := func(path, body string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodPost, path, token, body)
	}
	// The test owns the clock, so codes never expire mid-test. Each accepted
	// code burns its 30-second step and all earlier ones, so every step that
	// should succeed moves the clock forward.
	now := time.Now()
	e.a.Now = func() time.Time { return now }
	tick := func() { now = now.Add(30 * time.Second) }
	code := func(secret string) string {
		t.Helper()
		c, err := auth.TOTPCode(secret, now)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}

	want(t, "enable before setup", post("/auth/2fa/enable", codeBody("123456")), http.StatusConflict)
	want(t, "disable when off", post("/auth/2fa/disable", codeBody("123456")), http.StatusConflict)

	setup := post("/auth/2fa/setup", "")
	want(t, "setup", setup, http.StatusOK)
	out := decodeTo[struct {
		Secret     string
		OtpauthURL string `json:"otpauth_url"`
	}](t, setup)
	if out.Secret == "" || !strings.HasPrefix(out.OtpauthURL, "otpauth://totp/") {
		t.Fatalf("setup response wrong: %s", setup.Body)
	}

	// The database must not hold the secret in the clear.
	var stored string
	e.pool.QueryRow(t.Context(), `SELECT totp_secret FROM users WHERE email = $1`, email).Scan(&stored)
	if stored == "" || strings.Contains(stored, out.Secret) {
		t.Fatalf("secret stored unsealed: %q", stored)
	}

	// Until a code proves the app works, sign-in is unchanged.
	want(t, "login before enabling", do(e.h, http.MethodPost, "/auth/login", "", loginBody(email, testPassword)), http.StatusOK)
	want(t, "enable with wrong code", post("/auth/2fa/enable", codeBody("000000")), http.StatusBadRequest)

	first := code(out.Secret)
	want(t, "enable", post("/auth/2fa/enable", codeBody(first)), http.StatusOK)
	want(t, "setup while on", post("/auth/2fa/setup", ""), http.StatusConflict)

	// Sign-in now needs the code.
	noCode := do(e.h, http.MethodPost, "/auth/login", "", loginBody(email, testPassword))
	want(t, "login without code", noCode, http.StatusUnauthorized)
	if !decodeTo[struct {
		Required bool `json:"two_factor_required"`
	}](t, noCode).Required {
		t.Fatalf("response does not ask for a code: %s", noCode.Body)
	}
	login := func(code string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodPost, "/auth/login", "", loginWithCode(email, code))
	}
	want(t, "login with the code that enabled 2FA", login(first), http.StatusUnauthorized)
	want(t, "login with wrong code", login("000000"), http.StatusUnauthorized)
	tick()
	second := code(out.Secret)
	want(t, "login with code", login(second), http.StatusOK)
	want(t, "same code again", login(second), http.StatusUnauthorized)

	// A code from a much earlier or later time is not accepted.
	stale, _ := auth.TOTPCode(out.Secret, now.Add(-10*time.Minute))
	want(t, "stale code", login(stale), http.StatusUnauthorized)
	tick()
	third := code(out.Secret)

	// A wrong password must not reveal that 2FA is on.
	wrong := do(e.h, http.MethodPost, "/auth/login", "", fmt.Sprintf(`{"email":%q,"password":"nope","code":%q}`, email, third))
	want(t, "wrong password", wrong, http.StatusUnauthorized)
	if strings.Contains(wrong.Body.String(), "two_factor") {
		t.Fatalf("wrong password response mentions 2FA: %s", wrong.Body)
	}

	want(t, "disable with wrong code", post("/auth/2fa/disable", codeBody("000000")), http.StatusBadRequest)
	want(t, "disable", post("/auth/2fa/disable", codeBody(third)), http.StatusOK)
	want(t, "login after disabling", do(e.h, http.MethodPost, "/auth/login", "", loginBody(email, testPassword)), http.StatusOK)
}

func TestRateLimits(t *testing.T) {
	redisURL := os.Getenv(envTestRedis)
	if redisURL == "" {
		t.Skip(envTestRedis + " not set")
	}
	e, _, _ := newEnv(t, false)
	c, err := cache.New(redisURL, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	// Register before switching the limiter on: sign-ups from the test address
	// would otherwise count against the shared registration limit on every run.
	e.register(tAcme, org.KindProducer)
	e.register(tRival, org.KindExporter)
	e.a.Limiter = &auth.Limiter{Cache: c}
	t.Cleanup(func() { c.DelPattern(context.Background(), "*"+e.suffix+"*") }) // drop this test's counters
	email := e.email(tAcme)
	login := func(email, password string) int {
		return do(e.h, http.MethodPost, "/auth/login", "", loginBody(email, password)).Code
	}
	failures := func(n int64) {
		t.Helper()
		for i := int64(0); i < n; i++ {
			if code := login(email, "wrong-password"); code != http.StatusUnauthorized {
				t.Fatalf("failed sign-in %d = %d, want 401", i+1, code)
			}
		}
	}

	// A successful sign-in clears the count, so honest typos never add up.
	failures(auth.LoginRate.Max - 1)
	if code := login(email, testPassword); code != http.StatusOK {
		t.Fatalf("sign-in below the limit = %d", code)
	}
	failures(auth.LoginRate.Max)

	// At the limit even the right password is refused, for this account only.
	if code := login(email, testPassword); code != http.StatusTooManyRequests {
		t.Fatalf("sign-in over the limit = %d, want 429", code)
	}
	if code := login(e.email(tRival), testPassword); code != http.StatusOK {
		t.Fatalf("another account was locked out too: %d", code)
	}

	// LimitIP: at most Max calls per address per window.
	rate := auth.Rate{Name: "test-" + e.suffix, Max: 2, Window: time.Minute}
	h := e.a.LimitIP(rate, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	for i, wantCode := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests} {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != wantCode {
			t.Fatalf("call %d = %d, want %d", i+1, rec.Code, wantCode)
		}
	}

	// With no limiter configured, nothing is limited.
	open := (&auth.Auth{}).LimitIP(rate, func(w http.ResponseWriter, _ *http.Request) { httpx.WriteJSON(w, http.StatusOK, nil) })
	for i := int64(0); i <= rate.Max; i++ {
		rec := httptest.NewRecorder()
		open(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("unlimited call %d = %d", i+1, rec.Code)
		}
	}
}
