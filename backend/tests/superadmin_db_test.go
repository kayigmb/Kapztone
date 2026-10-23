package tests

import (
	"net/http"
	"slices"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/org"
)

// At an organisation's own address a superadmin acts inside it: that
// organisation's records only, with what its admin may do. Anyone else who
// doesn't belong there acts for nobody, and /me names the address so the
// app can send them to their own organisations.
func TestSuperadminInsideOrganisation(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.verifiedOrg(tAcme, org.KindProducer, root)
	rival := e.verifiedOrg(tRival, org.KindExporter, root)
	acmeSlug, rivalSlug := e.slug(tAcme), e.slug(tRival)

	me := func(token, path string) meResponse {
		rec := do(e.h, http.MethodGet, path, token, "")
		want(t, "/me "+path, rec, http.StatusOK)
		return decodeTo[meResponse](t, rec)
	}

	got := me(root, as(acmeSlug, "/me"))
	if got.Organisation == nil || got.Organisation.ID != acme.ID || got.Role != auth.RoleSuperadmin {
		t.Fatalf("superadmin at acme's address: %+v", got)
	}
	if !slices.Equal(got.Permissions, slices.Sorted(slices.Values(auth.DefaultRoles[auth.RoleOrgAdmin]))) {
		t.Fatalf("superadmin inside an organisation should hold its admin's permissions, got %v", got.Permissions)
	}
	if got := me(root, "/me"); got.Organisation != nil || !slices.Contains(got.Permissions, auth.OrgVerify) {
		t.Fatalf("superadmin at the bare address should keep the platform scope: %+v", got)
	}
	if got := me(root, as("t-no-such-org", "/me")); got.Organisation != nil || got.AddressSlug != "t-no-such-org" {
		t.Fatalf("an address naming no organisation keeps the platform scope: %+v", got)
	}

	// Only acme's rows, and no platform-only actions.
	rec := do(e.h, http.MethodGet, as(acmeSlug, "/orgs"), root, "")
	want(t, "orgs inside acme", rec, http.StatusOK)
	if p := decodeTo[struct {
		Total int
		Data  []org.Organisation
	}](t, rec); p.Total != 1 || p.Data[0].ID != acme.ID {
		t.Fatalf("superadmin inside acme sees %d organisations", p.Total)
	}
	want(t, "verify from inside acme", post(e, root, as(acmeSlug, "/orgs/"+rival.ID+"/verify"), org.Review{Decision: org.StatusVerified}), http.StatusForbidden)

	// A superadmin may carry the session to any organisation's address; a member only to their own.
	want(t, "superadmin handoff", post(e, root, "/auth/handoff", map[string]string{"slug": rivalSlug}), http.StatusOK)
	e.addUser(tStaff, auth.RoleMember, acme.ID)
	staff := e.login(tStaff)
	want(t, "member handoff elsewhere", post(e, staff, as(acmeSlug, "/auth/handoff"), map[string]string{"slug": rivalSlug}), http.StatusForbidden)
	if got := me(staff, as(rivalSlug, "/me")); got.Organisation != nil || got.AddressSlug != rivalSlug {
		t.Fatalf("member at another organisation's address: %+v", got)
	}

	// The bare address (AppURL's host) names no organisation, though it has a first label.
	if got := me(staff, testAppURL+"/me"); got.AddressSlug != "" {
		t.Fatalf("bare address read as organisation %q", got.AddressSlug)
	}
}
