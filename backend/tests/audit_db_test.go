package tests

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

const routeDocuments = "/orgs/{id}/documents"

func TestAuditLog(t *testing.T) {
	e, _, _ := newEnv(t, false)
	ctx := context.Background()
	rootID := e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.register(tAcme, org.KindProducer)
	rival := e.register(tRival, org.KindExporter)
	admin, rivalAdmin := e.login(tAcme), e.login(tRival)
	var adminID string
	e.pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email = $1`, e.email(tAcme)).Scan(&adminID)

	rows := func(route, subject string) (n int) {
		e.pool.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE route = $1 AND subject_id = $2`, route, subject).Scan(&n)
		return n
	}

	// A change is logged with who acted, for which organisation, and on what.
	want(t, "upload", e.upload(admin, e.slug(tAcme), acme.ID, org.DocLicence, pdfBytes), http.StatusCreated)
	want(t, "rival upload", e.upload(rivalAdmin, e.slug(tRival), rival.ID, org.DocLicence, pdfBytes), http.StatusCreated)
	var actor, orgID, method string
	var status int
	err := e.pool.QueryRow(ctx, `SELECT actor_user_id::text, organisation_id::text, method, status FROM audit_log
	                             WHERE route = $1 AND subject_id = $2`, routeDocuments, acme.ID).Scan(&actor, &orgID, &method, &status)
	if err != nil || actor != adminID || orgID != acme.ID || method != http.MethodPost || status != http.StatusCreated {
		t.Fatalf("upload logged wrongly: actor=%s org=%s method=%s status=%d (%v)", actor, orgID, method, status, err)
	}

	// Refused changes and reads are not logged.
	bad, _ := json.Marshal(org.Review{Decision: org.StatusRejected}) // no reason: 400
	want(t, "refused review", do(e.h, http.MethodPost, "/orgs/"+acme.ID+"/verify", root, string(bad)), http.StatusBadRequest)
	want(t, "read", do(e.h, http.MethodGet, as(e.slug(tAcme), "/orgs/"+acme.ID), admin, ""), http.StatusOK)
	if n := rows("/orgs/{id}/verify", acme.ID); n != 0 {
		t.Fatalf("refused change logged %d times", n)
	}
	if n := rows("/orgs/{id}", acme.ID); n != 0 {
		t.Fatalf("read logged %d times", n)
	}

	// Public changes are logged without an actor; no password reaches the log.
	var publicRows, secrets int
	e.pool.QueryRow(ctx, `SELECT count(*) FROM audit_log WHERE route = '/auth/login' AND actor_user_id IS NULL`).Scan(&publicRows)
	e.pool.QueryRow(ctx, `SELECT count(*) FROM audit_log a WHERE row_to_json(a)::text LIKE '%' || $1 || '%'`, testPassword).Scan(&secrets)
	if publicRows == 0 || secrets != 0 {
		t.Fatalf("sign-ins logged %d, rows holding a password %d", publicRows, secrets)
	}

	// An organisation admin sees only their organisation's rows; superadmin filters by actor and route.
	page := decodeTo[listq.Page[audit.Entry]](t, do(e.h, http.MethodGet, as(e.slug(tAcme), "/audit?size=100"), admin, ""))
	if page.Total == 0 {
		t.Fatal("organisation admin sees no rows")
	}
	for _, r := range page.Data {
		if r.OrganisationID == nil || *r.OrganisationID != acme.ID {
			t.Fatalf("organisation admin sees another organisation's row: %+v", r)
		}
	}
	rec := do(e.h, http.MethodGet, "/audit?actor="+adminID+"&route="+routeDocuments, root, "")
	want(t, "superadmin filter", rec, http.StatusOK)
	if got := decodeTo[listq.Page[audit.Entry]](t, rec); got.Total != 1 || *got.Data[0].ActorEmail != e.email(tAcme) {
		t.Fatalf("filtered by actor and route: %+v", got)
	}
	want(t, "bad actor", do(e.h, http.MethodGet, "/audit?actor=nobody", root, ""), http.StatusBadRequest)
	want(t, "bad format", do(e.h, http.MethodGet, "/audit?format=docx", root, ""), http.StatusBadRequest)
	want(t, "member of no platform scope verifies", do(e.h, http.MethodGet, as(e.slug(tAcme), "/audit/verify"), admin, ""), http.StatusForbidden)

	// CSV: heading, then every matching row; the organisation gate still applies.
	rec = do(e.h, http.MethodGet, as(e.slug(tAcme), "/audit?format=csv"), admin, "")
	want(t, "csv", rec, http.StatusOK)
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("csv content type = %q", ct)
	}
	records, err := csv.NewReader(rec.Body).ReadAll()
	if err != nil || len(records) != page.Total+1 || strings.Join(records[0], ",") != strings.Join(audit.AuditColumns, ",") {
		t.Fatalf("csv: %d records for %d rows (%v): %v", len(records), page.Total, err, records[:min(2, len(records))])
	}
	for _, r := range records[1:] {
		if r[4] != acme.ID {
			t.Fatalf("csv row of another organisation: %v", r)
		}
	}

	// The chain: intact, broken by an edit at exactly that row, intact once undone.
	check := func() audit.Check {
		rec := do(e.h, http.MethodGet, "/audit/verify", root, "")
		want(t, "verify chain", rec, http.StatusOK)
		return decodeTo[audit.Check](t, rec)
	}
	if c := check(); !c.OK || c.Checked == 0 {
		t.Fatalf("chain broken before any tampering: %+v", c)
	}
	var seq int64
	e.pool.QueryRow(ctx, `SELECT seq FROM audit_log WHERE route = $1 AND subject_id = $2`, routeDocuments, acme.ID).Scan(&seq)
	restore := func() { e.exec(`UPDATE audit_log SET actor_user_id = $2::uuid WHERE seq = $1`, seq, adminID) }
	t.Cleanup(restore) // never leave the shared database's chain broken
	e.exec(`UPDATE audit_log SET actor_user_id = $2::uuid WHERE seq = $1`, seq, rootID)
	if c := check(); c.OK || c.BrokenAt == nil || *c.BrokenAt != seq {
		t.Fatalf("edited row %d not found: %+v", seq, c)
	}
	restore()
	if c := check(); !c.OK {
		t.Fatalf("chain still broken after undoing the edit: %+v", c)
	}
}

func TestVolumeReport(t *testing.T) {
	e, _, _ := newEnv(t, false)
	ctx := context.Background()
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.verifiedOrg(tAcme, org.KindProducer, root)
	rival := e.verifiedOrg(tRival, org.KindProducer, root)
	admin, rivalAdmin := e.login(tAcme), e.login(tRival)

	facilityOf := func(token, slug string) facility.Facility {
		return e.approvedFacility(token, slug, root, facilityInput(newPlot()))
	}
	plot, rivalPlot := facilityOf(admin, e.slug(tAcme)), facilityOf(rivalAdmin, e.slug(tRival))
	tplIn := coffeeTemplate("coffee_" + e.suffix)
	tplIn.Name, tplIn.Unit = "Coffee "+e.suffix, plot.CapacityUnit
	rec := post(e, root, "/templates", tplIn)
	want(t, "template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)

	// Origin passports written directly: the report reads rows, not the issue rules.
	var issuer string
	e.pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email = $1`, e.email(tAcme)).Scan(&issuer)
	n := 0
	lot := func(orgID, facilityID string, qty int64, at string) {
		n++
		e.exec(`INSERT INTO passports (organisation_id, facility_id, template_id, commodity, unit, quantity, remaining,
		                               period, status, data_hash, issued_by, code, created_at)
		        VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $6, 0, 'active', '0x', $7::uuid, $8, $9::timestamptz)`,
			orgID, facilityID, tpl.ID, tpl.Commodity, tpl.Unit, qty, issuer, "T-"+e.suffix+"-"+string(rune('A'+n)), at)
	}
	lot(acme.ID, plot.ID, 100, "2026-03-02T10:00:00Z")
	lot(acme.ID, plot.ID, 50, "2026-03-30T10:00:00Z")
	lot(acme.ID, plot.ID, 7, "2026-04-01T10:00:00Z")
	lot(rival.ID, rivalPlot.ID, 1000, "2026-03-05T10:00:00Z")

	type report struct{ Data []audit.Volume }
	rec = do(e.h, http.MethodGet, as(e.slug(tAcme), "/reports/volume?start_date=2026-03-01&end_date=2026-04-30"), admin, "")
	want(t, "volume", rec, http.StatusOK)
	got := decodeTo[report](t, rec).Data
	wantRows := []audit.Volume{
		{Month: "2026-03", Commodity: tpl.Commodity, Unit: tpl.Unit, Quantity: 150},
		{Month: "2026-04", Commodity: tpl.Commodity, Unit: tpl.Unit, Quantity: 7},
	}
	if len(got) != len(wantRows) || got[0] != wantRows[0] || got[1] != wantRows[1] {
		t.Fatalf("volume = %+v, want %+v", got, wantRows)
	}

	// The date window cuts; superadmin sees both organisations.
	rec = do(e.h, http.MethodGet, as(e.slug(tAcme), "/reports/volume?start_date=2026-04-01"), admin, "")
	if got := decodeTo[report](t, rec).Data; len(got) != 1 || got[0].Quantity != 7 {
		t.Fatalf("from April: %+v", got)
	}
	var total int64
	for _, v := range decodeTo[report](t, do(e.h, http.MethodGet, "/reports/volume?start_date=2026-03-01&end_date=2026-03-31", root, "")).Data {
		if v.Commodity == tpl.Commodity {
			total += v.Quantity
		}
	}
	if total != 1150 {
		t.Fatalf("superadmin March total = %d, want 1150", total)
	}

	rec = do(e.h, http.MethodGet, as(e.slug(tAcme), "/reports/volume?format=csv&start_date=2026-03-01&end_date=2026-04-30"), admin, "")
	want(t, "volume csv", rec, http.StatusOK)
	wantCSV := strings.Join(audit.VolumeColumns, ",") + "\n2026-03," + tpl.Commodity + "," + tpl.Unit + ",150,0\n2026-04," +
		tpl.Commodity + "," + tpl.Unit + ",7,0\n"
	if rec.Body.String() != wantCSV {
		t.Fatalf("volume csv = %q, want %q", rec.Body.String(), wantCSV)
	}

	// The same reports as PDF; an unknown format is refused.
	for _, path := range []string{"/reports/volume", "/reports/blocked", "/audit"} {
		rec = do(e.h, http.MethodGet, as(e.slug(tAcme), path+"?format="+audit.FormatPDF), admin, "")
		want(t, path+" pdf", rec, http.StatusOK)
		if rec.Header().Get("Content-Type") != "application/pdf" || !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) ||
			!strings.Contains(rec.Header().Get("Content-Disposition"), ".pdf") {
			t.Fatalf("%s: not a PDF download (%s, %q…)", path, rec.Header().Get("Content-Type"), rec.Body.String()[:min(8, rec.Body.Len())])
		}
	}
	want(t, "unknown format", do(e.h, http.MethodGet, as(e.slug(tAcme), "/reports/volume?format=xlsx"), admin, ""), http.StatusBadRequest)
}
