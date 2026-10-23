package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
)

// A member of two organisations switches between their addresses without
// signing in again: a handoff code works once, only at the address it was
// made for, only for a minute, and only for an organisation they belong to.
func TestHandoff(t *testing.T) {
	e, _, _ := newEnv(t, false)
	ctx := context.Background()
	acme, beta, other := e.register("acme", "producer"), e.register("beta", "exporter"), e.register("other", "importer")
	userID := e.addUser(tStaff, auth.RoleMember, acme.ID)
	e.exec(`INSERT INTO userscope (user_id, role_id, organisation_id)
	        SELECT $1::uuid, id, $2::uuid FROM roles WHERE name = $3`, userID, beta.ID, auth.RoleMember)
	tok := e.login(tStaff)

	start := func(slug string) *http.Response {
		return do(e.h, http.MethodPost, as(acme.Slug, "/auth/handoff"), tok, `{"slug":"`+slug+`"}`).Result()
	}
	code := func(slug string) string {
		t.Helper()
		rec := do(e.h, http.MethodPost, as(acme.Slug, "/auth/handoff"), tok, `{"slug":"`+slug+`"}`)
		want(t, "start handoff", rec, http.StatusOK)
		return decodeTo[struct{ Code string }](t, rec).Code
	}
	accept := func(slug, c string) int {
		path := "/auth/handoff/accept"
		if slug != "" {
			path = as(slug, path)
		}
		return do(e.h, http.MethodPost, path, "", `{"code":"`+c+`"}`).Code
	}

	if got := start(other.Slug).StatusCode; got != http.StatusForbidden {
		t.Fatalf("handoff to an organisation they don't belong to = %d", got)
	}

	c := code(beta.Slug)
	if got := accept(acme.Slug, c); got != http.StatusBadRequest {
		t.Fatalf("code spent at another address = %d", got)
	}
	rec := do(e.h, http.MethodPost, as(beta.Slug, "/auth/handoff/accept"), "", `{"code":"`+c+`"}`)
	want(t, "accept at beta", rec, http.StatusOK)
	betaTok := decodeTo[struct{ Token string }](t, rec).Token
	me := decodeTo[struct{ Organisation *struct{ ID string } }](t, do(e.h, http.MethodGet, as(beta.Slug, "/me"), betaTok, ""))
	if me.Organisation == nil || me.Organisation.ID != beta.ID {
		t.Fatalf("session after handoff acts for %+v, want beta", me.Organisation)
	}
	if got := accept(beta.Slug, c); got != http.StatusBadRequest {
		t.Fatalf("code spent twice = %d", got)
	}
	if got := meCode(e, tok); got != http.StatusOK {
		t.Fatalf("the old address was signed out: %d", got)
	}

	// The bare address (the workspace picker) takes a code made for "", also
	// when it has a first label of its own; an organisation's address doesn't.
	if got := accept("", code("")); got != http.StatusOK {
		t.Fatalf("handoff to the bare address = %d", got)
	}
	if got := accept(acme.Slug, code("")); got != http.StatusBadRequest {
		t.Fatalf("bare-address code spent at an organisation's address = %d", got)
	}

	// A minute later the code is gone.
	late := "late-" + e.suffix
	if err := e.a.Accounts.AddLink(ctx, userID, auth.LinkHandoff, sha256Hex(late+":"+beta.Slug), time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if got := accept(beta.Slug, late); got != http.StatusBadRequest {
		t.Fatalf("expired code = %d", got)
	}
}

// One sign-in in one browser is one session, shared by every address a
// handoff carried it to: signing out at one ends it at all of them, while the
// same person signed in elsewhere (another device) stays signed in.
func TestSignOutEndsTheSharedSession(t *testing.T) {
	e, _, _ := newEnv(t, false)
	acme := e.register("acme", "producer")
	e.addUser(tStaff, auth.RoleMember, acme.ID)
	orgTok := e.login(tStaff)
	otherDevice := e.login(tStaff)

	rec := do(e.h, http.MethodPost, as(acme.Slug, "/auth/handoff"), orgTok, `{"slug":""}`)
	want(t, "start handoff", rec, http.StatusOK)
	rec = do(e.h, http.MethodPost, "/auth/handoff/accept", "", `{"code":"`+decodeTo[struct{ Code string }](t, rec).Code+`"}`)
	want(t, "accept at the bare address", rec, http.StatusOK)
	bareTok := decodeTo[struct{ Token string }](t, rec).Token

	want(t, "sign out at the organisation's address", do(e.h, http.MethodPost, as(acme.Slug, "/auth/logout"), orgTok, ""), http.StatusOK)
	if got := meCode(e, bareTok); got != http.StatusUnauthorized {
		t.Fatalf("the bare address is still signed in after signing out at the organisation's: %d", got)
	}
	if got := meCode(e, orgTok); got != http.StatusUnauthorized {
		t.Fatalf("the organisation's address is still signed in: %d", got)
	}
	if got := meCode(e, otherDevice); got != http.StatusOK {
		t.Fatalf("another device was signed out too: %d", got)
	}
	if got := meCode(e, e.login(tStaff)); got != http.StatusOK {
		t.Fatalf("signing in again doesn't work: %d", got)
	}
}
