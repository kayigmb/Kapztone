package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/user"
)

// withPIN sends a request with an explicit signing PIN header (send adds testPIN otherwise).
func withPIN(e *env, method, path, token, body, pin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(auth.HeaderSigningPIN, pin)
	return send(e.h, req, token)
}

func TestOrgDirectory(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)
	rival := e.verifiedOrg(tRival, org.KindExporter, root)
	pending := e.register(tStaff, org.KindImporter)
	acme := e.login(tAcme)

	rec := do(e.h, http.MethodGet, as(e.slug(tAcme), "/orgs/directory?size=100&search="+e.suffix), acme, "")
	want(t, "directory", rec, http.StatusOK)
	page := decodeTo[listq.Page[org.Listing]](t, rec)
	seen := map[string]org.Listing{}
	for _, l := range page.Data {
		seen[l.ID] = l
	}
	if got, ok := seen[rival.ID]; !ok || got.Slug != e.slug(tRival) || got.Kind != org.KindExporter || got.Name == "" {
		t.Fatalf("verified rival missing or incomplete: %+v", page.Data)
	}
	if _, ok := seen[pending.ID]; ok {
		t.Fatal("an unverified organisation is listed")
	}
	for _, l := range page.Data {
		if l.Slug == e.slug(tAcme) {
			t.Fatal("the caller's own organisation is listed")
		}
	}
	if data := string(envelopeOf(t, rec).Data); strings.Contains(data, "signing_address") || strings.Contains(data, "status") {
		t.Fatalf("directory shows more than name, kind and address: %s", rec.Body)
	}

	rec = do(e.h, http.MethodGet, as(e.slug(tAcme), "/orgs/directory?kind="+org.KindProducer+"&search="+e.suffix), acme, "")
	if p := decodeTo[listq.Page[org.Listing]](t, rec); len(p.Data) != 0 {
		t.Fatalf("kind filter ignored: %+v", p.Data)
	}
	want(t, "unknown kind", do(e.h, http.MethodGet, as(e.slug(tAcme), "/orgs/directory?kind=castle"), acme, ""), http.StatusBadRequest)
	want(t, "unverified caller", do(e.h, http.MethodGet, as(e.slug(tStaff), "/orgs/directory"), e.login(tStaff), ""), http.StatusForbidden)
}

func TestForwardedAddressDrivesRateLimit(t *testing.T) {
	e, _, cached := newEnv(t, true)
	e.a.Limiter = &auth.Limiter{Cache: cached.Cache}
	e.addUser(tStaff, auth.RoleMember, "")

	attempt := func(forwardedFor string) int {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(loginBody(e.email(tStaff), "wrong password!")))
		req.Header.Set(auth.HeaderForwardedFor, forwardedFor)
		return send(e.h, req, "").Code
	}
	for range auth.LoginRate.Max {
		if code := attempt("198.51.100.1, 203.0.113.7"); code != http.StatusUnauthorized {
			t.Fatalf("wrong password = %d", code)
		}
	}
	if code := attempt("203.0.113.7"); code != http.StatusTooManyRequests {
		t.Fatalf("sixth failure from the same visitor = %d, want 429", code)
	}
	// The last entry is the visitor; an address earlier in the list is not.
	if code := attempt("203.0.113.7, 203.0.113.99"); code != http.StatusUnauthorized {
		t.Fatalf("another visitor was blocked: %d", code)
	}
}

// lotForFlags issues one passport for acme through the API and raises a flag on it.
func lotForFlags(t *testing.T, e *env, root string) flag.Flag {
	t.Helper()
	acme := e.login(tAcme)
	slug := e.slug(tAcme)
	plot := e.approvedFacility(acme, slug, root, facilityInput(newPlot()))

	tplIn := coffeeTemplate("coffee_" + e.suffix)
	tplIn.Name, tplIn.Unit = "Coffee "+e.suffix, plot.CapacityUnit
	for i := range tplIn.Stages {
		tplIn.Stages[i].RequiredEvidence = nil // evidence rules are not what this test is about
	}
	rec := post(e, root, "/templates", tplIn)
	want(t, "create template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)

	rec = post(e, acme, as(slug, "/passports"), passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: 1000})
	want(t, "issue", rec, http.StatusCreated)
	p := decodeTo[passport.Passport](t, rec)

	rec = post(e, acme, as(slug, "/flags"), flag.RaiseInput{PassportID: p.ID, Reason: "Weighed twice at the gate"})
	want(t, "raise flag", rec, http.StatusCreated)
	return decodeTo[flag.Flag](t, rec)
}

func TestRegulatorResolvesEveryFlag(t *testing.T) {
	e, _, cached := newEnv(t, true)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)
	regOrg := e.verifiedOrg("reg", org.KindRegulator, root)
	e.verifiedOrg(tRival, org.KindExporter, root)
	f := lotForFlags(t, e, root)

	reg, regSlug := e.login("reg"), e.slug("reg")
	rival, rivalSlug := e.login(tRival), e.slug(tRival)

	list := func(token, slug string) []flag.Flag {
		rec := do(e.h, http.MethodGet, as(slug, "/flags?size=100"), token, "")
		want(t, "list flags", rec, http.StatusOK)
		return decodeTo[listq.Page[flag.Flag]](t, rec).Data
	}
	has := func(fs []flag.Flag) bool {
		for _, x := range fs {
			if x.ID == f.ID {
				return true
			}
		}
		return false
	}
	if !has(list(reg, regSlug)) {
		t.Fatal("regulator does not see another organisation's flag")
	}
	if has(list(rival, rivalSlug)) {
		t.Fatal("an exporter sees another organisation's flag")
	}
	want(t, "regulator reads the flag", do(e.h, http.MethodGet, as(regSlug, "/flags/"+f.ID), reg, ""), http.StatusOK)
	want(t, "exporter reads the flag", do(e.h, http.MethodGet, as(rivalSlug, "/flags/"+f.ID), rival, ""), http.StatusNotFound)

	resolve := `{"outcome":"` + flag.StatusCleared + `","note":"Scales recalibrated"}`
	want(t, "exporter resolves", do(e.h, http.MethodPost, as(rivalSlug, "/flags/"+f.ID+"/resolve"), rival, resolve), http.StatusForbidden)

	// A regulator's member reads every flag but only its admin resolves.
	e.addUser("regmember", auth.RoleMember, regOrg.ID)
	member := e.login("regmember")
	want(t, "regulator member reads", do(e.h, http.MethodGet, as(regSlug, "/flags/"+f.ID), member, ""), http.StatusOK)
	want(t, "regulator member resolves", do(e.h, http.MethodPost, as(regSlug, "/flags/"+f.ID+"/resolve"), member, resolve), http.StatusForbidden)

	// Resolution is a signing action: the PIN is checked.
	path := as(regSlug, "/flags/"+f.ID+"/resolve")
	want(t, "no PIN in the request", withPIN(e, http.MethodPost, path, reg, resolve, " "), http.StatusForbidden)
	want(t, "wrong PIN", withPIN(e, http.MethodPost, path, reg, resolve, "000000"), http.StatusForbidden)
	rec := do(e.h, http.MethodPost, path, reg, resolve)
	want(t, "regulator resolves", rec, http.StatusOK)
	if got := decodeTo[flag.Flag](t, rec); got.Status != flag.StatusCleared {
		t.Fatalf("flag status = %s", got.Status)
	}

	// A regulator that is not (or no longer) verified is an ordinary organisation.
	e.setOrgStatus(decodeTo[meResponse](t, do(e.h, http.MethodGet, as(regSlug, "/me"), reg, "")).Organisation.ID, org.StatusRejected)
	cached.InvalidateOrg(context.Background(), regSlug)
	if has(list(reg, regSlug)) {
		t.Fatal("an unverified regulator still sees every flag")
	}
}

func TestSigningPIN(t *testing.T) {
	e, _, cached := newEnv(t, true)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)
	e.verifiedOrg("reg", org.KindRegulator, root)
	f := lotForFlags(t, e, root)
	reg, regSlug := e.login("reg"), e.slug("reg")
	// Limits on only now: sign-ups above share one test address across runs.
	e.a.Limiter = &auth.Limiter{Cache: cached.Cache}
	resolve := `{"outcome":"` + flag.StatusCleared + `","note":"ok"}`
	path := as(regSlug, "/flags/"+f.ID+"/resolve")

	setPIN := func(pin, password string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodPost, "/auth/pin", reg, `{"pin":"`+pin+`","password":"`+password+`"}`)
	}
	want(t, "PIN too short", setPIN("12345", testPassword), http.StatusBadRequest)
	want(t, "PIN with letters", setPIN("12a456", testPassword), http.StatusBadRequest)
	want(t, "wrong password", setPIN("135790", "not my password"), http.StatusForbidden)
	want(t, "set PIN", setPIN("135790", testPassword), http.StatusOK)
	want(t, "old PIN no longer works", withPIN(e, http.MethodPost, path, reg, resolve, testPIN), http.StatusForbidden)

	// Five wrong PINs lock signing, even for the right one.
	for range auth.PINRate.Max - 1 {
		want(t, "wrong PIN", withPIN(e, http.MethodPost, path, reg, resolve, "000000"), http.StatusForbidden)
	}
	want(t, "locked", withPIN(e, http.MethodPost, path, reg, resolve, "135790"), http.StatusLocked)
	cached.Cache.Del(context.Background(), "auth:rate:pin:"+decodeTo[struct{ User auth.User }](t, do(e.h, http.MethodGet, "/me", reg, "")).User.ID)
	want(t, "right PIN after the lock", withPIN(e, http.MethodPost, path, reg, resolve, "135790"), http.StatusOK)

	// No PIN set at all.
	e.exec(`UPDATE users SET pin_hash = NULL WHERE email = $1`, e.email("reg"))
	rec := withPIN(e, http.MethodPost, path, reg, resolve, "135790")
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), auth.MsgPINNotSet) {
		t.Fatalf("no PIN set = %d %s", rec.Code, rec.Body)
	}
	if me := decodeTo[struct {
		PINSet bool `json:"pin_set"`
	}](t, do(e.h, http.MethodGet, "/me", reg, "")); me.PINSet {
		t.Fatal("/me says a PIN is set")
	}
}

func TestRecoveryCodes(t *testing.T) {
	e, _, _ := newEnv(t, false)
	now := time.Now()
	e.a.Now = func() time.Time { return now }
	e.addUser(tStaff, auth.RoleMember, "")
	tok := e.login(tStaff)

	rec := do(e.h, http.MethodPost, "/auth/2fa/setup", tok, "")
	want(t, "setup", rec, http.StatusOK)
	secret := decodeTo[struct{ Secret string }](t, rec).Secret
	code := func(at time.Time) string {
		c, err := auth.TOTPCode(secret, at)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}

	rec = do(e.h, http.MethodPost, "/auth/2fa/enable", tok, `{"code":"`+code(now)+`"}`)
	want(t, "enable", rec, http.StatusOK)
	codes := decodeTo[struct {
		RecoveryCodes []string `json:"recovery_codes"`
	}](t, rec).RecoveryCodes
	if len(codes) != auth.RecoveryCodeCount || len(codes[0]) != auth.RecoveryCodeLen {
		t.Fatalf("recovery codes: %v", codes)
	}
	var stored int
	e.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM recovery_codes WHERE code_hash = ANY($1::text[])`, codes).Scan(&stored)
	if stored != 0 {
		t.Fatal("recovery codes are stored in the clear")
	}

	login := func(c string) int {
		return do(e.h, http.MethodPost, "/auth/login", "",
			`{"email":"`+e.email(tStaff)+`","password":"`+testPassword+`","code":"`+c+`"}`).Code
	}
	if got := login(strings.ToLower(codes[0][:5]) + "-" + codes[0][5:]); got != http.StatusOK {
		t.Fatalf("sign-in with a recovery code (lower case, dashed) = %d", got)
	}
	if got := login(codes[0]); got != http.StatusUnauthorized {
		t.Fatalf("a recovery code worked twice: %d", got)
	}

	// Regenerating needs a current app code and retires the old codes.
	want(t, "regenerate without a code", do(e.h, http.MethodPost, "/auth/2fa/recovery", tok, `{"code":""}`), http.StatusBadRequest)
	now = now.Add(30 * time.Second)
	rec = do(e.h, http.MethodPost, "/auth/2fa/recovery", tok, `{"code":"`+code(now)+`"}`)
	want(t, "regenerate", rec, http.StatusOK)
	fresh := decodeTo[struct {
		RecoveryCodes []string `json:"recovery_codes"`
	}](t, rec).RecoveryCodes
	if got := login(codes[1]); got != http.StatusUnauthorized {
		t.Fatalf("an old code still works: %d", got)
	}
	if got := login(fresh[0]); got != http.StatusOK {
		t.Fatalf("a new code = %d", got)
	}

	// Turning 2FA off removes them.
	now = now.Add(30 * time.Second)
	want(t, "disable", do(e.h, http.MethodPost, "/auth/2fa/disable", tok, `{"code":"`+code(now)+`"}`), http.StatusOK)
	var live int
	e.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM recovery_codes r JOIN users u ON u.id = r.user_id
		 WHERE u.email = $1 AND NOT r.is_deleted`, e.email(tStaff)).Scan(&live)
	if live != 0 {
		t.Fatalf("%d recovery codes left after disabling 2FA", live)
	}
}

func TestUserManagement(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)
	e.verifiedOrg(tRival, org.KindExporter, root)
	admin, slug := e.login(tAcme), e.slug(tAcme)

	type inviteOut struct {
		Invite  user.Invite
		Emailed bool
		Link    string
	}
	invite := func(email, role string) inviteOut {
		t.Helper()
		rec := do(e.h, http.MethodPost, as(slug, "/invites"), admin, `{"email":"`+email+`","role":"`+role+`"}`)
		want(t, "invite "+email, rec, http.StatusCreated)
		return decodeTo[inviteOut](t, rec)
	}
	tokenOf := func(link string) string { return link[strings.LastIndex(link, user.InvitePath)+len(user.InvitePath):] }
	accept := func(token, name, password string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodPost, "/invites/"+token+"/accept", "", `{"name":"`+name+`","password":"`+password+`"}`)
	}

	want(t, "bad e-mail", do(e.h, http.MethodPost, as(slug, "/invites"), admin, `{"email":"not an address","role":"member"}`), http.StatusBadRequest)
	want(t, "platform role", do(e.h, http.MethodPost, as(slug, "/invites"), admin, `{"email":"a@b.test","role":"superadmin"}`), http.StatusBadRequest)

	// A new person: in dev the link comes back, since e-mail is usually off.
	newcomer := e.email("newcomer")
	inv := invite(newcomer, auth.RoleMember)
	if inv.Emailed || !strings.Contains(inv.Link, user.InvitePath) {
		t.Fatalf("dev invite: %+v", inv)
	}
	token := tokenOf(inv.Link)
	want(t, "accept without a name", accept(token, "", testPassword), http.StatusBadRequest)
	want(t, "accept with a short password", accept(token, "Nia", "short"), http.StatusBadRequest)
	want(t, "accept with a weak password", accept(token, "Nia", "lowercase-password1"), http.StatusBadRequest)
	rec := accept(token, "Nia", testPassword)
	want(t, "accept", rec, http.StatusOK)
	if got := decodeTo[user.Accepted](t, rec); !got.NewAccount || got.Organisation != slug {
		t.Fatalf("accepted: %+v", got)
	}
	want(t, "accept twice", accept(token, "Nia", testPassword), http.StatusNotFound)
	nia := e.login("newcomer")
	me := decodeTo[meResponse](t, do(e.h, http.MethodGet, as(slug, "/me"), nia, ""))
	if me.Organisation == nil || me.Role != auth.RoleMember {
		t.Fatalf("newcomer's place: %+v", me)
	}

	// An existing account proves itself with its password.
	inv = invite(e.email(tRival), auth.RoleOrgAdmin)
	want(t, "existing account, wrong password", accept(tokenOf(inv.Link), "", "wrong password!"), http.StatusForbidden)
	rec = accept(tokenOf(inv.Link), "", testPassword)
	want(t, "existing account joins", rec, http.StatusOK)
	if decodeTo[user.Accepted](t, rec).NewAccount {
		t.Fatal("an existing account was created again")
	}

	// Expired, and a production invite: e-mailed, link never returned.
	inv = invite(e.email("late"), auth.RoleMember)
	e.exec(`UPDATE invites SET expires_at = now() - interval '1 minute' WHERE id = $1::uuid`, inv.Invite.ID)
	want(t, "expired invite", accept(tokenOf(inv.Link), "Late", testPassword), http.StatusNotFound)
	sent := make(chan [2]string, 1)
	e.users.Dev = false
	e.users.Mail = func(to, _, text, _ string) error { sent <- [2]string{to, text}; return nil }
	inv = invite(e.email("mailed"), auth.RoleMember)
	if !inv.Emailed || inv.Link != "" {
		t.Fatalf("production invite: %+v", inv)
	}
	var sentTo, sentBody string
	select {
	case got := <-sent:
		sentTo, sentBody = got[0], got[1]
	case <-time.After(10 * time.Second):
		t.Fatal("invitation e-mail was not queued")
	}
	if sentTo != e.email("mailed") || !strings.Contains(sentBody, auth.OrgURL(testAppURL, slug)+user.InvitePath) {
		t.Fatalf("production invite sent to %q: %q", sentTo, sentBody)
	}
	var plain int
	e.pool.QueryRow(context.Background(), `SELECT count(*) FROM invites WHERE token_hash = $1`,
		tokenOf(sentBody[:strings.LastIndex(sentBody, "\n")])).Scan(&plain)
	if plain != 0 {
		t.Fatal("an invite token is stored in the clear")
	}

	// Members, roles, the last admin.
	rec = do(e.h, http.MethodGet, as(slug, "/users"), admin, "")
	want(t, "list members", rec, http.StatusOK)
	members := decodeTo[listq.Page[user.Member]](t, rec).Data
	byEmail := map[string]user.Member{}
	for _, m := range members {
		byEmail[m.Email] = m
	}
	niaID, adminID := byEmail[newcomer].ID, byEmail[e.email(tAcme)].ID
	if len(members) != 3 || byEmail[newcomer].Status != user.StatusActive {
		t.Fatalf("members: %+v", members)
	}
	want(t, "member manages users", do(e.h, http.MethodGet, as(slug, "/users"), nia, ""), http.StatusForbidden)
	want(t, "suspend oneself", post(e, admin, as(slug, "/users/"+adminID+"/suspend"), nil), http.StatusConflict)

	rivalAdminID := byEmail[e.email(tRival)].ID
	want(t, "demote the second admin", do(e.h, http.MethodPut, as(slug, "/users/"+rivalAdminID+"/role"), admin, `{"role":"member"}`), http.StatusOK)
	want(t, "demote the last admin", do(e.h, http.MethodPut, as(slug, "/users/"+adminID+"/role"), admin, `{"role":"member"}`), http.StatusConflict)
	want(t, "promote", do(e.h, http.MethodPut, as(slug, "/users/"+niaID+"/role"), admin, `{"role":"org_admin"}`), http.StatusOK)
	if me := decodeTo[meResponse](t, do(e.h, http.MethodGet, as(slug, "/me"), nia, "")); me.Role != auth.RoleOrgAdmin {
		t.Fatalf("promotion not seen: %+v", me)
	}
	// Two admins now, so only the self rule can refuse this.
	want(t, "demote oneself", do(e.h, http.MethodPut, as(slug, "/users/"+adminID+"/role"), admin, `{"role":"member"}`), http.StatusConflict)
	rec = do(e.h, http.MethodGet, as(slug, "/users"), admin, "")
	want(t, "list members again", rec, http.StatusOK)
	stillAdmin := false
	for _, m := range decodeTo[listq.Page[user.Member]](t, rec).Data {
		if m.ID == adminID && m.Role == auth.RoleOrgAdmin {
			stillAdmin = true
		}
	}
	if !stillAdmin {
		t.Fatal("self-demotion went through")
	}
	want(t, "someone else's member", do(e.h, http.MethodPut, as(slug, "/users/"+fakeOrgID+"/role"), admin, `{"role":"member"}`), http.StatusNotFound)

	// Suspend cuts access to this organisation only; reinstate gives it back.
	want(t, "suspend", post(e, admin, as(slug, "/users/"+niaID+"/suspend"), nil), http.StatusOK)
	want(t, "suspended member's passports", do(e.h, http.MethodGet, as(slug, "/passports"), nia, ""), http.StatusForbidden)
	want(t, "suspend twice", post(e, admin, as(slug, "/users/"+niaID+"/suspend"), nil), http.StatusNotFound)
	if m := decodeTo[listq.Page[user.Member]](t, do(e.h, http.MethodGet, as(slug, "/users"), admin, "")).Data; len(m) != 3 {
		t.Fatalf("a suspended member left the list: %+v", m)
	}
	rivalTok := e.login(tRival)
	want(t, "the other organisation is untouched", do(e.h, http.MethodGet, as(e.slug(tRival), "/passports"), rivalTok, ""), http.StatusOK)
	want(t, "reinstate", post(e, admin, as(slug, "/users/"+niaID+"/reinstate"), nil), http.StatusOK)
	want(t, "reinstated member's passports", do(e.h, http.MethodGet, as(slug, "/passports"), nia, ""), http.StatusOK)
	want(t, "reinstate an active member", post(e, admin, as(slug, "/users/"+niaID+"/reinstate"), nil), http.StatusConflict)
}
