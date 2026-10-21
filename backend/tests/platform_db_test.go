package tests

import (
	"net/http"
	"slices"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/stats"
	"github.com/kayigmb/chaintrack/backend/internal/user"
)

// The platform hears about new organisations and their documents; it does not
// see what organisations are told, and organisations don't see its notices.
func TestPlatformNotifications(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.register(tAcme, org.KindProducer)
	admin, slug := e.login(tAcme), e.slug(tAcme)

	type page struct{ Data []notify.Notification }
	list := func(path, token string) []notify.Notification {
		t.Helper()
		rec := do(e.h, http.MethodGet, path, token, "")
		want(t, "list "+path, rec, http.StatusOK)
		return decodeTo[page](t, rec).Data
	}
	find := func(ns []notify.Notification, kind, subject string) *notify.Notification {
		for i := range ns {
			if ns[i].Kind == kind && ns[i].SubjectID == subject {
				return &ns[i]
			}
		}
		return nil
	}

	registered := find(list("/notifications?size=100", root), notify.KindOrgRegistered, acme.ID)
	if registered == nil || registered.OrganisationID != nil {
		t.Fatalf("platform notice of the new organisation: %+v", registered)
	}

	want(t, "upload", e.upload(admin, slug, acme.ID, org.DocRegistration, pdfBytes), http.StatusCreated)
	if find(list("/notifications?size=100", root), notify.KindDocumentsUploaded, acme.ID) == nil {
		t.Fatal("platform not told about uploaded documents")
	}

	// Something only the organisation is told about.
	e.exec(`INSERT INTO notifications (organisation_id, kind, subject_id, message) VALUES ($1::uuid, $2, $1::uuid, 'for acme')`,
		acme.ID, notify.KindHandoverStarted)
	own := list(as(slug, "/notifications?size=100"), admin)
	if find(own, notify.KindHandoverStarted, acme.ID) == nil {
		t.Fatal("organisation does not see its own notification")
	}
	if find(own, notify.KindOrgRegistered, acme.ID) != nil || find(own, notify.KindDocumentsUploaded, acme.ID) != nil {
		t.Fatal("organisation sees platform notifications")
	}
	if find(list("/notifications?size=100", root), notify.KindHandoverStarted, acme.ID) != nil {
		t.Fatal("platform sees an organisation's notification")
	}

	want(t, "organisation reads a platform notice", do(e.h, http.MethodPost, as(slug, "/notifications/"+registered.ID+"/read"), admin, ""), http.StatusNotFound)
	want(t, "platform reads its notice", do(e.h, http.MethodPost, "/notifications/"+registered.ID+"/read", root, ""), http.StatusOK)
	unread := list("/notifications?size=100&unread=true", root)
	if find(unread, notify.KindOrgRegistered, acme.ID) != nil {
		t.Fatal("read notice still listed as unread")
	}
}

// Profile: my organisations, my name, my password; the platform's list of every user.
func TestProfileAndPlatformUsers(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.register(tAcme, org.KindProducer)
	admin := e.login(tAcme)

	// The switcher's list works on the platform address too (no organisation acting).
	mine := decodeTo[[]user.MyOrganisation](t, do(e.h, http.MethodGet, "/me/organisations", admin, ""))
	if len(mine) != 1 || mine[0].ID != acme.ID || mine[0].Slug != e.slug(tAcme) || mine[0].Role != auth.RoleOrgAdmin {
		t.Fatalf("my organisations: %+v", mine)
	}
	if got := decodeTo[[]user.MyOrganisation](t, do(e.h, http.MethodGet, "/me/organisations", root, "")); len(got) != 0 {
		t.Fatalf("platform admin belongs to no organisation: %+v", got)
	}

	// /me says whether two-factor sign-in is on, for the profile page.
	twoFactor := func() bool {
		t.Helper()
		return decodeTo[struct {
			TwoFactor bool `json:"two_factor"`
		}](t, do(e.h, http.MethodGet, "/me", admin, "")).TwoFactor
	}
	if twoFactor() {
		t.Fatal("two-factor reported on before it was turned on")
	}
	e.exec(`UPDATE users SET totp_enabled = true WHERE email = $1`, e.email(tAcme))
	if !twoFactor() {
		t.Fatal("two-factor not reported on")
	}
	e.exec(`UPDATE users SET totp_enabled = false WHERE email = $1`, e.email(tAcme))

	want(t, "empty name", do(e.h, http.MethodPut, "/me", admin, `{"name":"  "}`), http.StatusBadRequest)
	rec := do(e.h, http.MethodPut, "/me", admin, `{"name":"  Amina K.  "}`)
	want(t, "rename", rec, http.StatusOK)
	if u := decodeTo[auth.User](t, rec); u.Name != "Amina K." || u.Email != e.email(tAcme) {
		t.Fatalf("renamed user: %+v", u)
	}

	newPassword := "A-brand-new-password1!"
	want(t, "wrong current password", do(e.h, http.MethodPost, "/auth/password", admin,
		`{"current":"not-it-at-all","new":"`+newPassword+`"}`), http.StatusForbidden)
	want(t, "short new password", do(e.h, http.MethodPost, "/auth/password", admin,
		`{"current":"`+testPassword+`","new":"short"}`), http.StatusBadRequest)
	want(t, "weak new password", do(e.h, http.MethodPost, "/auth/password", admin,
		`{"current":"`+testPassword+`","new":"lowercase-password1"}`), http.StatusBadRequest)
	want(t, "change password", do(e.h, http.MethodPost, "/auth/password", admin,
		`{"current":"`+testPassword+`","new":"`+newPassword+`"}`), http.StatusOK)
	want(t, "old password", do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tAcme), testPassword)), http.StatusUnauthorized)
	want(t, "new password", do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(tAcme), newPassword)), http.StatusOK)

	// The platform sees every user with every membership; an organisation only its members.
	all := decodeTo[listq.Page[user.Account]](t, do(e.h, http.MethodGet, "/users?size=100", root, "")).Data
	at := slices.IndexFunc(all, func(a user.Account) bool { return a.Email == e.email(tAcme) })
	if at < 0 || len(all[at].Memberships) != 1 || all[at].Memberships[0].Organisation != tAcme ||
		all[at].Memberships[0].Role != auth.RoleOrgAdmin || all[at].Memberships[0].Status != user.StatusActive || all[at].Name != "Amina K." {
		t.Fatalf("platform users list, acme admin: %+v", all)
	}
	at = slices.IndexFunc(all, func(a user.Account) bool { return a.Email == e.email(tRoot) })
	if at < 0 || len(all[at].Memberships) != 1 || all[at].Memberships[0].OrganisationID != nil || all[at].Memberships[0].Organisation != "" {
		t.Fatalf("platform users list, root: %+v", all[at])
	}
}

// The platform dashboard: one answer with every block; organisations get none.
func TestPlatformStats(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)

	rec := do(e.h, http.MethodGet, "/stats", root, "")
	want(t, "platform stats", rec, http.StatusOK)
	st := decodeTo[stats.Stats](t, rec)
	if len(st.Activity) != stats.Days || st.Users < 2 || len(st.Organisations) == 0 || st.Activity[len(st.Activity)-1].Signups < 1 {
		t.Fatalf("stats: %+v", st)
	}
	want(t, "an organisation asks", do(e.h, http.MethodGet, as(e.slug(tAcme), "/stats"), e.login(tAcme), ""), http.StatusForbidden)
}
