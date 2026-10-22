package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/user"
)

// A new sign-up confirms its email before signing in; signing in first sends a new link.
func TestEmailVerification(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.register(tAcme, org.KindProducer)
	e.exec(`UPDATE users SET email_verified_at = NULL WHERE email = $1`, e.email(tAcme))

	rec := do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tAcme), testPassword))
	want(t, "sign in before confirming", rec, http.StatusForbidden)
	var links int
	e.pool.QueryRow(t.Context(),
		`SELECT count(*) FROM password_resets pr JOIN users u ON u.id = pr.user_id WHERE u.email = $1 AND pr.kind = $2`,
		e.email(tAcme), auth.LinkVerify).Scan(&links)
	if links != 1 {
		t.Fatalf("confirmation links sent on sign-in = %d, want 1", links)
	}

	// The emailed token is never stored, so the test plants one of its own.
	token := "test-verify-" + e.suffix
	sum := sha256.Sum256([]byte(token))
	e.exec(`INSERT INTO password_resets (user_id, kind, token_hash, expires_at)
	        SELECT id, $2, $3, now() + interval '1 hour' FROM users WHERE email = $1`,
		e.email(tAcme), auth.LinkVerify, hex.EncodeToString(sum[:]))
	body := `{"token":"` + token + `"}`
	want(t, "confirm", do(e.h, http.MethodPost, "/auth/email/verify", "", body), http.StatusOK)
	want(t, "confirm twice", do(e.h, http.MethodPost, "/auth/email/verify", "", body), http.StatusNotFound)
	e.login(tAcme)

	// Dev hands the link back too (scripts, browser tests); prod never does.
	if link := e.a.SendVerifyTo(t.Context(), e.email(tAcme)); !strings.HasPrefix(link, testAppURL+auth.PathVerify) {
		t.Fatalf("dev link = %q", link)
	}
	e.a.Dev = false
	defer func() { e.a.Dev = true }()
	if link := e.a.SendVerifyTo(t.Context(), e.email(tAcme)); link != "" {
		t.Fatalf("prod returned the link: %q", link)
	}
}

// The platform admin turns accounts, memberships and organisations off and on.
func TestPlatformAdminActions(t *testing.T) {
	e, _, _ := newEnv(t, false)
	rootID := e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.register(tAcme, org.KindProducer)
	rival := e.register(tRival, org.KindProducer)
	staffID := e.addUser(tStaff, auth.RoleMember, acme.ID)
	admin, slug := e.login(tAcme), e.slug(tAcme)

	// One account, and its memberships.
	rec := do(e.h, http.MethodGet, "/users/"+staffID, root, "")
	want(t, "user detail", rec, http.StatusOK)
	if a := decodeTo[user.Account](t, rec); a.Email != e.email(tStaff) || len(a.Memberships) != 1 || !a.EmailVerified {
		t.Fatalf("user detail: %+v", a)
	}
	want(t, "org admin reads user detail", do(e.h, http.MethodGet, as(slug, "/users/"+staffID), admin, ""), http.StatusForbidden)

	// A membership, suspended and reinstated by the platform.
	scope := "/users/" + staffID + "/%s?organisation_id=" + acme.ID
	want(t, "suspend without organisation", do(e.h, http.MethodPost, "/users/"+staffID+"/suspend", root, ""), http.StatusBadRequest)
	want(t, "platform suspends membership", do(e.h, http.MethodPost, fmt.Sprintf(scope, "suspend"), root, ""), http.StatusOK)
	want(t, "platform reinstates membership", do(e.h, http.MethodPost, fmt.Sprintf(scope, "reinstate"), root, ""), http.StatusOK)

	// The account itself.
	want(t, "deactivate self", do(e.h, http.MethodPost, "/users/"+rootID+"/deactivate", root, ""), http.StatusConflict)
	want(t, "org admin deactivates", do(e.h, http.MethodPost, as(slug, "/users/"+staffID+"/deactivate"), admin, ""), http.StatusForbidden)
	want(t, "deactivate", do(e.h, http.MethodPost, "/users/"+staffID+"/deactivate", root, ""), http.StatusOK)
	want(t, "sign in deactivated", do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tStaff), testPassword)), http.StatusUnauthorized)
	want(t, "activate", do(e.h, http.MethodPost, "/users/"+staffID+"/activate", root, ""), http.StatusOK)
	e.login(tStaff)

	// One organisation's members, and lists narrowed to it.
	rec = do(e.h, http.MethodGet, "/orgs/"+acme.ID+"/members", root, "")
	want(t, "organisation members", rec, http.StatusOK)
	if p := decodeTo[listq.Page[user.Member]](t, rec); p.Total != 2 {
		t.Fatalf("acme members = %d, want 2", p.Total)
	}
	rec = do(e.h, http.MethodGet, "/orgs?organisation_id="+rival.ID, root, "")
	want(t, "narrowed list", rec, http.StatusOK)
	if p := decodeTo[listq.Page[org.Organisation]](t, rec); p.Total != 1 || p.Data[0].ID != rival.ID {
		t.Fatalf("narrowed to rival: %+v", p)
	}
	want(t, "bad organisation_id", do(e.h, http.MethodGet, "/orgs?organisation_id=x", root, ""), http.StatusBadRequest)

	// Suspend and reactivate the organisation.
	want(t, "suspend pending", do(e.h, http.MethodPost, "/orgs/"+acme.ID+"/suspend", root, `{"reason":"fraud"}`), http.StatusConflict)
	e.exec(`UPDATE organisations SET status = $2 WHERE id = $1::uuid`, acme.ID, org.StatusVerified)
	want(t, "suspend without reason", do(e.h, http.MethodPost, "/orgs/"+acme.ID+"/suspend", root, `{}`), http.StatusBadRequest)
	want(t, "suspend", do(e.h, http.MethodPost, "/orgs/"+acme.ID+"/suspend", root, `{"reason":"fraud"}`), http.StatusOK)
	want(t, "suspended organisation acts", do(e.h, http.MethodPost, as(slug, "/facilities"), admin, `{}`), http.StatusForbidden)
	want(t, "suspended organisation invites", do(e.h, http.MethodPost, as(slug, "/invites"), admin,
		`{"email":"`+e.email("late")+`","role":"`+auth.RoleMember+`"}`), http.StatusForbidden)
	want(t, "suspended organisation still reads", do(e.h, http.MethodGet, as(slug, "/users"), admin, ""), http.StatusOK)
	want(t, "reactivate", do(e.h, http.MethodPost, "/orgs/"+acme.ID+"/reactivate", root, ""), http.StatusOK)
	if rec := do(e.h, http.MethodPost, as(slug, "/facilities"), admin, `{}`); rec.Code == http.StatusForbidden {
		t.Fatalf("reactivated organisation still refused: %s", rec.Body)
	}
}
