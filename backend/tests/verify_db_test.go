package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
)

// reason returns the first reason with this code, if any.
func reason(res verify.Result, code string) (verify.Reason, bool) {
	for _, r := range res.Reasons {
		if r.Code == code {
			return r, true
		}
	}
	return verify.Reason{}, false
}

// stmt is one SQL statement with its arguments.
type stmt struct {
	sql  string
	args []any
}

func q(sql string, args ...any) stmt { return stmt{sql, args} }

// TestPublicVerification walks one lot through every verdict the public page
// can show: clean, waiting for the chain, and each kind of tampering.
func TestPublicVerification(t *testing.T) {
	e, _, _ := newEnv(t, false)
	c := e.newChain()
	ctx := context.Background()

	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	producer := e.verifiedOrg(tAcme, org.KindProducer, root)
	processor := e.verifiedOrg(tRival, org.KindProcessor, root)
	certifier := e.verifiedOrg(tStaff, org.KindCertifier, root)
	prodTok, procTok, certTok := e.login(tAcme), e.login(tRival), e.login(tStaff)
	prodSlug, procSlug, certSlug := e.slug(tAcme), e.slug(tRival), e.slug(tStaff)

	var lotID string // the passport the tampering checks verify
	// check verifies a passport with no sign-in and holds the answer to a verdict.
	check := func(what, id, verdict string, codes ...string) verify.Result {
		t.Helper()
		rec := do(e.h, http.MethodGet, "/verify/"+id, "", "")
		want(t, what, rec, http.StatusOK)
		res := decodeTo[verify.Result](t, rec)
		if res.Verdict != verdict {
			t.Fatalf("%s: verdict %s, want %s (%+v)", what, res.Verdict, verdict, res.Reasons)
		}
		for _, code := range codes {
			if r, ok := reason(res, code); !ok || r.Message == "" {
				t.Fatalf("%s: no %s reason in %+v", what, code, res.Reasons)
			}
		}
		// Messages are shown to the public as they are: they say what was checked, never how.
		for _, r := range res.Reasons {
			for _, word := range verify.ForbiddenWords {
				if strings.Contains(strings.ToLower(r.Message), word) {
					t.Fatalf("%s: message for %s says %q: %s", what, r.Code, word, r.Message)
				}
			}
		}
		return res
	}
	// tamper changes a row behind the API's back, checks the verdict, and puts the row back.
	tamper := func(what, verdict, code string, change, restore stmt) {
		t.Helper()
		e.exec(change.sql, change.args...)
		check(what, lotID, verdict, code)
		e.exec(restore.sql, restore.args...)
	}

	rec := post(e, prodTok, as(prodSlug, "/facilities"), facilityInput(plotGeo))
	want(t, "create facility", rec, http.StatusCreated)
	plot := decodeTo[facility.Facility](t, rec)
	tplIn := withoutOriginEvidence(coffeeTemplate("coffee_" + e.suffix))
	tplIn.Name, tplIn.Unit = "Coffee "+e.suffix, plot.CapacityUnit
	rec = post(e, root, "/templates", tplIn)
	want(t, "create template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)
	// Certificates below are export permits, evidence the template asks for.
	want(t, "accredit", accredit(e, root, "/orgs/"+certifier.ID+"/accreditations", grant("export_permit")), http.StatusOK)

	const qty = 500_000
	rec = post(e, prodTok, as(prodSlug, "/passports"), passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: qty})
	want(t, "issue", rec, http.StatusCreated)
	origin := decodeTo[passport.Passport](t, rec)

	// ---- Not on the chain yet: never "verified".
	check("before the relayer runs", origin.ID, verify.Warning, verify.NotAnchored)

	handOver := func(id, fromTok, fromSlug, to, toTok, toSlug string) {
		t.Helper()
		rec := post(e, fromTok, as(fromSlug, "/passports/"+id+"/handovers"), passport.HandoverInput{ToOrganisationID: to, Stage: stageWashing})
		want(t, "start handover", rec, http.StatusCreated)
		h := decodeTo[passport.Handover](t, rec)
		want(t, "accept handover", post(e, toTok, as(toSlug, "/handovers/"+h.ID+"/accept"), nil), http.StatusOK)
	}
	handOver(origin.ID, prodTok, prodSlug, processor.ID, procTok, procSlug)
	rec = post(e, procTok, as(procSlug, "/transforms"), passport.TransformInput{
		Stage:   "washing_station",
		Inputs:  []passport.TransformSource{{PassportID: origin.ID, Quantity: qty}},
		Outputs: []passport.TransformOutput{{Quantity: qty / 5}},
	})
	want(t, "transform", rec, http.StatusCreated)
	tr := decodeTo[passport.Transform](t, rec)
	lot := tr.Outputs[0]
	rec = e.certify(certTok, certSlug, certificateFields(attestation.SubjectPassport, lot.ID), pdfBytes)
	want(t, "certify", rec, http.StatusCreated)
	cert := decodeTo[attestation.Attestation](t, rec)

	// ---- Anchored and untouched: verified, with everything the page shows.
	c.drain()
	rec = do(e.h, http.MethodGet, "/verify/"+lot.ID, "", "")
	clean := check("clean lot", lot.ID, verify.Verified)
	if len(clean.Reasons) != 0 || len(clean.Checks) != len(verify.Checks) || len(clean.Genealogy.Nodes) != 2 ||
		len(clean.Genealogy.Edges) != 1 || len(clean.Certificates) != 1 || len(clean.Passport.Events) != 1 ||
		len(clean.Origins) != 1 || clean.Origins[0].Facility.Name != plot.Name || clean.Origins[0].Organisation.ID != producer.ID ||
		len(clean.Origins[0].Facility.Location) == 0 || len(clean.Organisations) != 3 {
		t.Fatalf("clean verification: %s", rec.Body)
	}
	for _, ch := range clean.Checks {
		if ch.Result != verify.Verified {
			t.Fatalf("check %s = %s on a clean lot", ch.Name, ch.Result)
		}
	}
	// Public means limited: no licence numbers, no people, no storage names.
	for _, private := range []string{plot.LicenceNumber, e.email(tAcme), "storage_key", "issued_by"} {
		if strings.Contains(rec.Body.String(), private) {
			t.Fatalf("public verification exposes %q", private)
		}
	}
	want(t, "unknown passport", do(e.h, http.MethodGet, "/verify/"+fakeOrgID, "", ""), http.StatusNotFound)
	want(t, "not an id", do(e.h, http.MethodGet, "/verify/123", "", ""), http.StatusNotFound)

	// ---- The chain cannot be reached: a warning, never a pass.
	c.backend.down = true
	check("chain down", lot.ID, verify.Warning, verify.ChainUnreachable)
	c.backend.down = false

	// ---- Tampering with the database, one record at a time.
	lotID = lot.ID
	tamper("ancestor's quantity changed", verify.Failed, verify.RecordAltered,
		q(`UPDATE passports SET quantity = quantity + 1 WHERE id = $1::uuid`, origin.ID),
		q(`UPDATE passports SET quantity = quantity - 1 WHERE id = $1::uuid`, origin.ID))
	tamper("holder changed", verify.Failed, verify.ChainMismatch,
		q(`UPDATE passports SET organisation_id = $2::uuid WHERE id = $1::uuid`, lot.ID, certifier.ID),
		q(`UPDATE passports SET organisation_id = $2::uuid WHERE id = $1::uuid`, lot.ID, processor.ID))
	tamper("output inflated", verify.Failed, verify.MassBalanceBroken,
		q(`UPDATE transforms SET max_yield_ppm = 1 WHERE id = $1::uuid`, tr.ID),
		q(`UPDATE transforms SET max_yield_ppm = $2 WHERE id = $1::uuid`, tr.ID, tr.MaxYieldPPM))
	tamper("capacity lowered below what was issued", verify.Failed, verify.CapacityExceeded,
		q(`UPDATE facilities SET capacity = $2 WHERE id = $1::uuid`, plot.ID, int64(qty-1)),
		q(`UPDATE facilities SET capacity = $2 WHERE id = $1::uuid`, plot.ID, plot.Capacity))
	tamper("facility suspended", verify.Warning, verify.FacilitySuspended,
		q(`UPDATE facilities SET status = $2 WHERE id = $1::uuid`, plot.ID, facility.StatusSuspended),
		q(`UPDATE facilities SET status = $2 WHERE id = $1::uuid`, plot.ID, facility.StatusActive))
	tamper("origin link removed", verify.Failed, verify.OriginMissing,
		q(`UPDATE transform_inputs SET is_deleted = true WHERE transform_id = $1::uuid`, tr.ID),
		q(`UPDATE transform_inputs SET is_deleted = false WHERE transform_id = $1::uuid`, tr.ID))

	// ---- Certificates: the file is re-hashed, so an altered file is caught.
	var storageKey string
	e.pool.QueryRow(ctx, `SELECT storage_key FROM attestations WHERE id = $1::uuid`, cert.ID).Scan(&storageKey)
	files := e.certificates.Files
	if err := files.Put(ctx, storageKey, append([]byte("%PDF-1.4 forged "), pdfBytes...)); err != nil {
		t.Fatal(err)
	}
	check("certificate file altered", lot.ID, verify.Failed, verify.CertificateAltered)
	files.Remove(ctx, storageKey)
	check("certificate file missing", lot.ID, verify.Warning, verify.CertificateFileGone)
	if err := files.Put(ctx, storageKey, pdfBytes); err != nil {
		t.Fatal(err)
	}
	tamper("certificate fingerprint changed", verify.Failed, verify.CertificateAltered,
		q(`UPDATE attestations SET doc_hash = '0x' || repeat('ab', 32) WHERE id = $1::uuid`, cert.ID),
		q(`UPDATE attestations SET doc_hash = $2 WHERE id = $1::uuid`, cert.ID, cert.DocHash))
	tamper("certificate type changed", verify.Failed, verify.CertificateAltered,
		q(`UPDATE attestations SET type = 'organic_certificate' WHERE id = $1::uuid`, cert.ID),
		q(`UPDATE attestations SET type = $2 WHERE id = $1::uuid`, cert.ID, cert.Type))
	tamper("certificate expired", verify.Warning, verify.CertificateExpired,
		q(`UPDATE attestations SET valid_to = now() - interval '1 day' WHERE id = $1::uuid`, cert.ID),
		q(`UPDATE attestations SET valid_to = $2 WHERE id = $1::uuid`, cert.ID, cert.ValidTo))

	want(t, "withdraw accreditation", accredit(e, root, "/orgs/"+certifier.ID+"/accreditations"), http.StatusOK)
	c.drain()
	check("issuer no longer accredited", lot.ID, verify.Warning, verify.IssuerNotAccredited)
	want(t, "revoke certificate", post(e, certTok, as(certSlug, "/attestations/"+cert.ID+"/revoke"), attestation.RevokeInput{Reason: "Issued in error"}), http.StatusOK)
	check("certificate revoked", lot.ID, verify.Warning, verify.CertificateRevoked)
	check("lot's ancestor has no certificate of its own", origin.ID, verify.Verified)

	// ---- A handover the chain has not seen yet is a warning, then clears.
	handOver(lot.ID, procTok, procSlug, producer.ID, prodTok, prodSlug)
	check("handover not anchored yet", lot.ID, verify.Warning, verify.CustodyPending)
	c.drain()
	if _, still := reason(check("handover anchored", lot.ID, verify.Warning), verify.CustodyPending); still {
		t.Fatal("custody warning remains after the handover was anchored")
	}

	// ---- A flag on an ancestor shows on the lot; confirmed, the lot fails.
	rec = post(e, prodTok, as(prodSlug, "/flags"), flag.RaiseInput{PassportID: origin.ID, Reason: "This lot never left our plot"})
	want(t, "raise flag", rec, http.StatusCreated)
	raised := decodeTo[flag.Flag](t, rec)
	open := check("open flag on an ancestor", lot.ID, verify.Warning, verify.FlagOpen)
	if r, _ := reason(open, verify.FlagOpen); r.SubjectID != origin.ID || strings.Contains(r.Message, "never left") {
		t.Fatalf("flag reason: %+v", r)
	}
	want(t, "confirm flag", post(e, root, "/flags/"+raised.ID+"/resolve", flag.ResolveInput{Outcome: flag.StatusConfirmed}), http.StatusOK)
	revoked := check("revoked ancestor", lot.ID, verify.Failed, verify.PassportRevoked)
	for _, ch := range revoked.Checks {
		if ch.Name == verify.CheckStatus && ch.Result != verify.Failed || ch.Name == verify.CheckMassBalance && ch.Result != verify.Verified {
			t.Fatalf("check %s = %s after revocation", ch.Name, ch.Result)
		}
	}

	// ---- With Redis: answers are cached, and each address is rate limited.
	redisURL := os.Getenv(envTestRedis)
	if redisURL == "" {
		t.Log(envTestRedis + " not set: cache and rate limit not checked")
		return
	}
	store, err := cache.New(redisURL, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	clear := func() {
		store.DelPattern(ctx, verify.CacheKeyPrefix+"*")
		store.DelPattern(ctx, "auth:rate:"+auth.VerifyRate.Name+":*")
	}
	clear()
	t.Cleanup(func() { clear(); store.Close() })
	e.verify.Cache = store
	defer func() { e.verify.Cache = nil }()

	check("first answer, now cached", lot.ID, verify.Failed)
	e.exec(`UPDATE passports SET status = $2 WHERE id = ANY($1::uuid[])`, []string{origin.ID, lot.ID}, passport.StatusActive)
	check("served from the cache", lot.ID, verify.Failed)
	store.DelPattern(ctx, verify.CacheKeyPrefix+"*")
	check("worked out afresh", lot.ID, verify.Warning)

	e.a.Limiter = &auth.Limiter{Cache: store}
	defer func() { e.a.Limiter = nil }()
	var last *httptest.ResponseRecorder
	for i := int64(0); i <= auth.VerifyRate.Max; i++ {
		last = do(e.h, http.MethodGet, "/verify/"+lot.ID, "", "")
	}
	want(t, "one request over the limit", last, http.StatusTooManyRequests)
}
