package tests

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// PassportRegistry.Status values, as the contract numbers them.
const (
	chainActive   = 1
	chainConsumed = 2
	chainRevoked  = 4
)

// Stage keys of the test commodities.
const (
	stageProcess = "processing" // loses mass: the commodity's own yield
	stageSort    = "sorting"    // splits and merges: yield 1
)

// certify sends a certificate as the multipart form the API expects.
func (e *env) certify(token, slug string, fields map[string]string, content []byte) *httptest.ResponseRecorder {
	e.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		mw.WriteField(k, v)
	}
	fw, _ := mw.CreateFormFile(upload.FormFile, "certificate.pdf")
	fw.Write(content)
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, as(slug, "/attestations"), &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return send(e.h, req, token)
}

func certificateFields(subjectType, subjectID string) map[string]string {
	return map[string]string{
		attestation.FormSubjectType: subjectType,
		attestation.FormSubjectID:   subjectID,
		attestation.FormType:        "export_permit",
		attestation.FormValidFrom:   "2026-01-01",
		attestation.FormValidTo:     "2036-01-01T00:00:00Z",
	}
}

func errorOf(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	return envelopeOf(t, rec).Message
}

// TestCommodityChainsEndToEnd is the Sprint 3 goal: a coffee chain and a
// mineral chain run through the API from origin to revocation, with every
// step anchored on the chain. The two differ only in their template.
func TestCommodityChainsEndToEnd(t *testing.T) {
	cases := []struct {
		name         string
		facilityKind string
		rule         template.CapacityRule
		yield        float64
	}{
		{"coffee", facility.KindFarmPlot, template.CapacityRule{Type: template.CapacityYieldPerArea, KgPerHaPerSeason: 1200}, washingYield},
		{"mineral", facility.KindMineSite, template.CapacityRule{Type: template.CapacityFacility}, 0.05},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runChain(t, tc.name, tc.facilityKind, tc.rule, tc.yield)
		})
	}
}

func runChain(t *testing.T, name, facilityKind string, rule template.CapacityRule, yield float64) {
	e, _, _ := newEnv(t, false)
	c := e.newChain()
	ctx := context.Background()
	opts := &bind.CallOpts{Context: ctx}
	started := time.Now().Add(-time.Second)

	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	producer := e.verifiedOrg(tAcme, org.KindProducer, root)
	processor := e.verifiedOrg(tRival, org.KindProcessor, root)
	certifier := e.verifiedOrg(tStaff, org.KindCertifier, root)
	prodTok, procTok, certTok := e.login(tAcme), e.login(tRival), e.login(tStaff)
	prodSlug, procSlug, certSlug := e.slug(tAcme), e.slug(tRival), e.slug(tStaff)

	// at attaches the acting organisation; root acts for none.
	at := func(slug, path string) string {
		if slug == "" {
			return path
		}
		return as(slug, path)
	}
	get := func(token, slug, path string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodGet, at(slug, path), token, "")
	}
	passportOf := func(id string) passport.Passport {
		t.Helper()
		rec := get(root, "", "/passports/"+id)
		want(t, "get passport", rec, http.StatusOK)
		return decodeTo[passport.Passport](t, rec)
	}
	onChain := func(id string) (status uint8, remaining int64) {
		t.Helper()
		p, err := c.contract.Passports.Passport(opts, chain.MustID(id))
		if err != nil {
			t.Fatalf("passport %s is not on-chain: %v", id, err)
		}
		return p.Status, p.Remaining.Int64()
	}

	// ---- Origin: a facility, a template with a lossy stage and a yield-1 stage, two passports.
	fIn := facilityInput(newPlot())
	fIn.Kind = facilityKind
	site := e.approvedFacility(prodTok, prodSlug, root, fIn)

	whole := 1.0
	tplIn := template.Input{
		Commodity: name + "_" + e.suffix, Name: name + " " + e.suffix, Unit: site.CapacityUnit,
		OriginFacilityKind: facilityKind, CapacityRule: rule,
		Stages: []template.Stage{
			{Key: "origin", Events: []string{template.EventIssue}},
			{Key: stageProcess, Events: []string{template.EventHandover, template.EventTransform}, MaxYield: &yield},
			{Key: stageSort, Events: []string{template.EventTransform}, MaxYield: &whole},
			{Key: "export", Events: []string{template.EventHandover, template.EventExport}, RequiredEvidence: []string{"export_permit"}},
		},
	}
	rec := post(e, root, "/templates", tplIn)
	want(t, "create template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)

	issue := func(quantity int64) passport.Passport {
		t.Helper()
		rec := post(e, prodTok, as(prodSlug, "/passports"), passport.IssueInput{FacilityID: site.ID, TemplateID: tpl.ID, Quantity: quantity})
		want(t, "issue", rec, http.StatusCreated)
		return decodeTo[passport.Passport](t, rec)
	}
	handOver := func(id string) {
		t.Helper()
		rec := post(e, prodTok, as(prodSlug, "/passports/"+id+"/handovers"), passport.HandoverInput{ToOrganisationID: processor.ID, Stage: stageProcess})
		want(t, "start handover", rec, http.StatusCreated)
		h := decodeTo[passport.Handover](t, rec)
		want(t, "accept handover", post(e, procTok, as(procSlug, "/handovers/"+h.ID+"/accept"), received(h)), http.StatusOK)
	}
	const qtyA, qtyB, usedB = 600_000, 400_000, 100_000
	a, b := issue(qtyA), issue(qtyB)
	if a.FacilityID == nil || *a.FacilityID != site.ID || a.TransformID != nil {
		t.Fatalf("origin passport: %+v", a)
	}
	handOver(a.ID)
	handOver(b.ID)

	// ---- Transform: every rule has a refused case.
	transform := func(token, slug string, in passport.TransformInput) *httptest.ResponseRecorder {
		return post(e, token, as(slug, "/transforms"), in)
	}
	src := func(id string, q int64) passport.TransformSource {
		return passport.TransformSource{PassportID: id, Quantity: q}
	}
	outs := func(qs ...int64) (o []passport.TransformOutput) {
		for _, q := range qs {
			o = append(o, passport.TransformOutput{Quantity: q})
		}
		return o
	}
	inputs := []passport.TransformSource{src(a.ID, qtyA), src(b.ID, usedB)}
	maxOut := int64(float64(qtyA+usedB) * yield)
	const small = 10_000

	want(t, "former holder transforms", transform(prodTok, prodSlug, passport.TransformInput{Stage: stageProcess, Inputs: inputs, Outputs: outs(1)}), http.StatusBadRequest)
	want(t, "unknown stage", transform(procTok, procSlug, passport.TransformInput{Stage: "smelting", Inputs: inputs, Outputs: outs(1)}), http.StatusBadRequest)
	want(t, "stage that does not transform", transform(procTok, procSlug, passport.TransformInput{Stage: "export", Inputs: inputs, Outputs: outs(1)}), http.StatusBadRequest)
	want(t, "same input twice", transform(procTok, procSlug, passport.TransformInput{
		Stage: stageProcess, Inputs: []passport.TransformSource{src(a.ID, 1), src(a.ID, 1)}, Outputs: outs(1)}), http.StatusBadRequest)

	overdrawn := transform(procTok, procSlug, passport.TransformInput{
		Stage: stageProcess, Inputs: []passport.TransformSource{src(a.ID, qtyA+1)}, Outputs: outs(1)})
	want(t, "more than the input holds", overdrawn, http.StatusConflict)
	if got, wantMsg := errorOf(t, overdrawn), (passport.InsufficientError{PassportID: a.ID, Requested: qtyA + 1, Remaining: qtyA, Unit: site.CapacityUnit}).Error(); got != wantMsg {
		t.Fatalf("overdrawn message = %q, want %q", got, wantMsg)
	}

	// Mass balance: one unit over what the inputs can yield is refused, and the message says the limit.
	breach := transform(procTok, procSlug, passport.TransformInput{Stage: stageProcess, Inputs: inputs, Outputs: outs(maxOut-small, small+1)})
	want(t, "mass balance breach", breach, http.StatusConflict)
	if got, wantMsg := errorOf(t, breach), (passport.YieldError{Outputs: maxOut + 1, Max: maxOut, Unit: site.CapacityUnit}).Error(); got != wantMsg {
		t.Fatalf("yield message = %q, want %q", got, wantMsg)
	}
	if got := passportOf(a.ID); got.Remaining != qtyA || got.Status != passport.StatusActive {
		t.Fatalf("a refused transform changed an input: %+v", got)
	}

	// Merge two inputs (one in part) and process them, at exactly the limit.
	rec = transform(procTok, procSlug, passport.TransformInput{Stage: stageProcess, Inputs: inputs, Outputs: outs(maxOut-small, small)})
	want(t, "transform", rec, http.StatusCreated)
	tr := decodeTo[passport.Transform](t, rec)
	if len(tr.Outputs) != 2 || tr.MaxYieldPPM != template.YieldPPM(yield) || len(tr.Inputs) != 2 {
		t.Fatalf("transform recorded wrongly: %+v", tr)
	}
	child1, child2 := tr.Outputs[0], tr.Outputs[1]
	if child1.OrganisationID != processor.ID || child1.FacilityID != nil || child1.Period != nil ||
		child1.TransformID == nil || *child1.TransformID != tr.ID || child1.Quantity != maxOut-small || child1.Status != passport.StatusActive {
		t.Fatalf("output passport: %+v", child1)
	}
	// At the very limit the transform is allowed, but the rules flag its first output.
	var ruleFlags int
	e.pool.QueryRow(ctx, `SELECT count(*) FROM flags WHERE passport_id = $1::uuid AND source = $2`,
		child1.ID, flag.SourceRules).Scan(&ruleFlags)
	if ruleFlags != 1 {
		t.Fatalf("rules raised %d flags about a transform at its yield limit, want 1", ruleFlags)
	}
	if got := passportOf(a.ID); got.Remaining != 0 || got.Status != passport.StatusConsumed {
		t.Fatalf("an input used up is not consumed: %+v", got)
	}
	if got := passportOf(b.ID); got.Remaining != qtyB-usedB || got.Status != passport.StatusActive {
		t.Fatalf("an input used in part: %+v", got)
	}
	want(t, "consumed input used again", transform(procTok, procSlug, passport.TransformInput{
		Stage: stageProcess, Inputs: []passport.TransformSource{src(a.ID, 1)}, Outputs: outs(1)}), http.StatusConflict)

	// Split one output in two at the yield-1 stage: nothing may appear from nowhere.
	half := child1.Quantity / 2
	split := []passport.TransformSource{src(child1.ID, child1.Quantity)}
	want(t, "split that adds mass", transform(procTok, procSlug, passport.TransformInput{
		Stage: stageSort, Inputs: split, Outputs: outs(half, child1.Quantity-half+1)}), http.StatusConflict)
	rec = transform(procTok, procSlug, passport.TransformInput{Stage: stageSort, Inputs: split, Outputs: outs(half, child1.Quantity-half)})
	want(t, "split", rec, http.StatusCreated)
	grand := decodeTo[passport.Transform](t, rec).Outputs
	g1, g2 := grand[0], grand[1]

	// ---- The chain agrees with all of it.
	c.drain()
	if got, _ := c.contract.Passports.MaxYield(opts, chain.Key(tpl.Commodity), chain.Key(stageProcess)); got.Int64() != template.YieldPPM(yield) {
		t.Fatalf("on-chain yield = %s, want %d", got, template.YieldPPM(yield))
	}
	if status, _ := onChain(a.ID); status != chainConsumed {
		t.Fatalf("on-chain input status = %d, want consumed", status)
	}
	if status, remaining := onChain(b.ID); status != chainActive || remaining != qtyB-usedB {
		t.Fatalf("on-chain partly used input: status %d, remaining %d", status, remaining)
	}
	if status, _ := onChain(child1.ID); status != chainConsumed {
		t.Fatalf("on-chain split input status = %d, want consumed", status)
	}
	if status, remaining := onChain(g1.ID); status != chainActive || remaining != half {
		t.Fatalf("on-chain output: status %d, remaining %d", status, remaining)
	}
	anchored := passportOf(g1.ID)
	if anchored.TxHash == nil || len(anchored.Events) != 1 || anchored.Events[0].Kind != passport.EventTransform || anchored.Events[0].TxHash == nil {
		t.Fatalf("transform hash was not written back to the output: %+v", anchored)
	}

	// ---- Genealogy, both ways.
	tree := func(token, slug, id, direction string) passport.Genealogy {
		t.Helper()
		rec := get(token, slug, "/passports/"+id+"/genealogy?"+passport.ParamDirection+"="+direction)
		want(t, "genealogy", rec, http.StatusOK)
		return decodeTo[passport.Genealogy](t, rec)
	}
	ids := func(g passport.Genealogy) (out []string) {
		for _, n := range g.Nodes {
			out = append(out, n.ID)
		}
		slices.Sort(out)
		return out
	}
	sorted := func(s ...string) []string { slices.Sort(s); return s }

	up := tree(procTok, procSlug, g1.ID, passport.DirectionAncestors)
	if !slices.Equal(ids(up), sorted(g1.ID, child1.ID, a.ID, b.ID)) || len(up.Edges) != 3 {
		t.Fatalf("ancestors of a split output: %d nodes, %d edges", len(up.Nodes), len(up.Edges))
	}
	for _, edge := range up.Edges {
		if edge.ParentID == b.ID && (edge.ChildID != child1.ID || edge.QuantityUsed != usedB || edge.Stage != stageProcess) {
			t.Fatalf("edge from the partly used input: %+v", edge)
		}
	}
	// The default direction is ancestors; an origin passport has none.
	rec = get(procTok, procSlug, "/passports/"+a.ID+"/genealogy")
	want(t, "default genealogy", rec, http.StatusOK)
	if g := decodeTo[passport.Genealogy](t, rec); len(g.Nodes) != 1 || len(g.Edges) != 0 {
		t.Fatalf("an origin passport has ancestors: %+v", g)
	}
	// The producer no longer holds the lot but can still follow what became of it.
	down := tree(prodTok, prodSlug, a.ID, passport.DirectionDescendants)
	if !slices.Equal(ids(down), sorted(a.ID, child1.ID, child2.ID, g1.ID, g2.ID)) {
		t.Fatalf("descendants of an origin passport: %v", ids(down))
	}
	want(t, "bad direction", get(procTok, procSlug, "/passports/"+a.ID+"/genealogy?"+passport.ParamDirection+"=sideways"), http.StatusBadRequest)
	want(t, "genealogy of a passport out of view", get(certTok, certSlug, "/passports/"+a.ID+"/genealogy"), http.StatusNotFound)

	// ---- Certificates: only an accredited organisation issues them.
	want(t, "issue before accreditation", e.certify(certTok, certSlug, certificateFields(attestation.SubjectPassport, g1.ID), pdfBytes), http.StatusForbidden)
	accPath := "/orgs/" + certifier.ID + "/accreditations"
	window := func(certType, from, to string) org.AccreditationInput {
		return org.AccreditationInput{Type: certType, ValidFrom: from, ValidTo: to}
	}
	want(t, "an organisation accredits itself", accredit(e, certTok, as(certSlug, accPath), grant("export_permit")), http.StatusForbidden)
	want(t, "accredit a producer", accredit(e, root, "/orgs/"+producer.ID+"/accreditations", grant("export_permit")), http.StatusConflict)
	for what, grants := range map[string][]org.AccreditationInput{
		"type no template asks for": {grant("moon_rock_" + e.suffix)},
		"type not a key":            {grant("Export Permit")},
		"type twice":                {grant("export_permit"), grant("export_permit")},
		"ends before starts":        {window("export_permit", "2030-01-01", "2029-01-01")},
		"date in words":             {window("export_permit", "today", "2030-01-01")},
	} {
		want(t, what, accredit(e, root, accPath, grants...), http.StatusBadRequest)
	}
	rec = accredit(e, root, accPath, grant("export_permit"))
	want(t, "accredit", rec, http.StatusOK)
	accs := decodeTo[[]org.Accreditation](t, rec)
	if len(accs) != 1 || accs[0].Type != "export_permit" || accs[0].OrganisationID != certifier.ID {
		t.Fatalf("accreditations after granting: %s", rec.Body)
	}
	accID := accs[0].ID
	want(t, "accredit twice", accredit(e, root, accPath, grant("export_permit")), http.StatusOK)
	if statuses, _ := e.outbox(accID); len(statuses) != 1 { // an unchanged window is not queued again
		t.Fatalf("queued %d items for the accreditation, want 1", len(statuses))
	}
	listed := get(certTok, certSlug, "/orgs/"+certifier.ID+"/accreditations")
	want(t, "list own accreditations", listed, http.StatusOK)
	if got := decodeTo[[]org.Accreditation](t, listed); len(got) != 1 || got[0].ID != accID {
		t.Fatalf("listed accreditations: %s", listed.Body)
	}
	// A certificate needs accreditation for its own type, inside the window.
	other := certificateFields(attestation.SubjectPassport, g1.ID)
	other[attestation.FormType] = "phytosanitary_certificate"
	want(t, "issue a type not accredited", e.certify(certTok, certSlug, other, pdfBytes), http.StatusForbidden)
	want(t, "window in the past", accredit(e, root, accPath, window("export_permit", "2020-01-01", "2021-01-01")), http.StatusOK)
	want(t, "issue after the window", e.certify(certTok, certSlug, certificateFields(attestation.SubjectPassport, g1.ID), pdfBytes), http.StatusForbidden)
	want(t, "window back to today", accredit(e, root, accPath, grant("export_permit")), http.StatusOK)
	if statuses, _ := e.outbox(accID); len(statuses) != 3 { // granted, moved, moved back: one row, three changes
		t.Fatalf("queued %d items for the accreditation, want 3", len(statuses))
	}

	bad := func(change func(map[string]string)) map[string]string {
		f := certificateFields(attestation.SubjectPassport, g1.ID)
		change(f)
		return f
	}
	for what, fields := range map[string]map[string]string{
		"unknown subject":    certificateFields(attestation.SubjectPassport, fakeOrgID),
		"unknown kind":       bad(func(f map[string]string) { f[attestation.FormSubjectType] = "shipment" }),
		"type not a key":     bad(func(f map[string]string) { f[attestation.FormType] = "Export Permit" }),
		"ends before starts": bad(func(f map[string]string) { f[attestation.FormValidTo] = "2025-01-01" }),
		"date in words":      bad(func(f map[string]string) { f[attestation.FormValidFrom] = "yesterday" }),
	} {
		want(t, what, e.certify(certTok, certSlug, fields, pdfBytes), http.StatusBadRequest)
	}
	want(t, "file that is not a document", e.certify(certTok, certSlug, certificateFields(attestation.SubjectPassport, g1.ID), exeBytes), http.StatusBadRequest)

	rec = e.certify(certTok, certSlug, certificateFields(attestation.SubjectPassport, g1.ID), pdfBytes)
	want(t, "issue certificate", rec, http.StatusCreated)
	cert := decodeTo[attestation.Attestation](t, rec)
	sum := sha256.Sum256(pdfBytes)
	if cert.DocHash != "0x"+hex.EncodeToString(sum[:]) || cert.Status != attestation.StatusValid || cert.IPFSCID != nil ||
		cert.IssuerOrgID != certifier.ID || strings.Contains(rec.Body.String(), "storage_key") {
		t.Fatalf("certificate recorded wrongly: %s", rec.Body)
	}

	// IPFS: the file is pinned once the certificate is recorded; a failed pin does not undo it.
	var pinned []byte
	e.certificates.Pin = func(_ context.Context, _ string, content io.Reader) (string, error) {
		pinned, _ = io.ReadAll(content)
		return "bafytestcid", nil
	}
	rec = e.certify(certTok, certSlug, certificateFields(attestation.SubjectFacility, site.ID), pdfBytes)
	want(t, "issue with pinning", rec, http.StatusCreated)
	if got := decodeTo[attestation.Attestation](t, rec); got.IPFSCID == nil || *got.IPFSCID != "bafytestcid" || !bytes.Equal(pinned, pdfBytes) {
		t.Fatalf("certificate was not pinned: %+v", got)
	}
	e.certificates.Pin = func(context.Context, string, io.Reader) (string, error) { return "", errors.New("ipfs is down") }
	old := bad(func(f map[string]string) {
		f[attestation.FormValidFrom], f[attestation.FormValidTo] = "2020-01-01", "2021-01-01"
	})
	rec = e.certify(certTok, certSlug, old, pdfBytes)
	want(t, "issue while IPFS is down", rec, http.StatusCreated)
	lapsed := decodeTo[attestation.Attestation](t, rec)
	if lapsed.IPFSCID != nil || lapsed.Status != attestation.StatusExpired {
		t.Fatalf("certificate past its validity, issued with IPFS down: %+v", lapsed)
	}

	// The failed copy is retried once due: still down doubles the wait, then it pins.
	clock := time.Now()
	retrier := &attestation.Retrier{Store: e.certificates.Store, Files: e.certificates.Files, Pin: e.certificates.Pin, Now: func() time.Time { return clock }}
	if n, err := retrier.RetryDue(ctx); err != nil || n != 0 {
		t.Fatalf("retried before it was due: %d, %v", n, err)
	}
	clock = clock.Add(attestation.Backoff(1) + time.Second)
	if n, err := retrier.RetryDue(ctx); err != nil || n != 1 {
		t.Fatalf("due retry: %d, %v", n, err)
	}
	var attempts int
	if err := e.pool.QueryRow(ctx, `SELECT ipfs_attempts FROM attestations WHERE id = $1::uuid`, lapsed.ID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("attempts after a second failure = %d, want 2", attempts)
	}
	retrier.Pin = func(context.Context, string, io.Reader) (string, error) { return "bafy-retried", nil }
	clock = clock.Add(attestation.Backoff(2) + time.Second)
	if n, err := retrier.RetryDue(ctx); err != nil || n != 1 {
		t.Fatalf("retry that pins: %d, %v", n, err)
	}
	if got := decodeTo[attestation.Attestation](t, get(certTok, certSlug, "/attestations/"+lapsed.ID)); got.IPFSCID == nil || *got.IPFSCID != "bafy-retried" {
		t.Fatalf("retried certificate not pinned: %+v", got)
	}
	e.certificates.Pin = nil

	certs := func(query string) listq.Page[attestation.Attestation] {
		t.Helper()
		rec := get(procTok, procSlug, "/attestations?"+query)
		want(t, "list certificates", rec, http.StatusOK)
		return decodeTo[listq.Page[attestation.Attestation]](t, rec)
	}
	aboutG1 := attestation.ParamSubjectID + "=" + g1.ID
	if pg := certs(aboutG1); pg.Total != 2 {
		t.Fatalf("certificates about the passport = %d, want 2", pg.Total)
	}
	if pg := certs(aboutG1 + "&" + attestation.ParamStatus + "=" + attestation.StatusExpired); pg.Total != 1 || pg.Data[0].ID != lapsed.ID {
		t.Fatalf("expired filter returned %d", pg.Total)
	}
	want(t, "bad status filter", get(procTok, procSlug, "/attestations?"+attestation.ParamStatus+"=bogus"), http.StatusBadRequest)
	want(t, "bad subject filter", get(procTok, procSlug, "/attestations?"+attestation.ParamSubjectID+"=123"), http.StatusBadRequest)
	if pg := certs(aboutG1 + "&" + attestation.ParamIssuerID + "=" + certifier.ID); pg.Total != 2 {
		t.Fatalf("certificates issued by the certifier = %d, want 2", pg.Total)
	}
	if pg := certs(attestation.ParamIssuerID + "=" + fakeOrgID); pg.Total != 0 {
		t.Fatalf("certificates issued by nobody = %d", pg.Total)
	}
	want(t, "bad issuer filter", get(procTok, procSlug, "/attestations?"+attestation.ParamIssuerID+"=123"), http.StatusBadRequest)

	// Who sees which certificates: the holder sees those about its lot, the
	// issuer those it issued, an unrelated organisation none, the platform all.
	listAs := func(token, slug, query string) int {
		t.Helper()
		rec := get(token, slug, "/attestations?"+aboutG1+query)
		want(t, "scoped certificates", rec, http.StatusOK)
		return decodeTo[listq.Page[attestation.Attestation]](t, rec).Total
	}
	scope := "&" + attestation.ParamScope + "="
	if listAs(procTok, procSlug, scope+attestation.ScopeAbout) != 2 || listAs(procTok, procSlug, scope+attestation.ScopeIssued) != 0 {
		t.Fatal("the holder's about / issued split is wrong")
	}
	if listAs(certTok, certSlug, scope+attestation.ScopeIssued) != 2 || listAs(certTok, certSlug, scope+attestation.ScopeAbout) != 0 {
		t.Fatal("the issuer's issued / about split is wrong")
	}
	if listAs(prodTok, prodSlug, "") != 0 {
		t.Fatal("an organisation sees certificates it neither issued nor holds")
	}
	if rec := do(e.h, http.MethodGet, "/attestations?"+aboutG1, root, ""); decodeTo[listq.Page[attestation.Attestation]](t, rec).Total != 2 {
		t.Fatal("the platform doesn't see every certificate")
	}
	want(t, "bad scope", get(procTok, procSlug, "/attestations?"+attestation.ParamScope+"=mine"), http.StatusBadRequest)
	want(t, "get certificate", get(procTok, procSlug, "/attestations/"+cert.ID), http.StatusOK)
	want(t, "unknown certificate", get(procTok, procSlug, "/attestations/"+fakeOrgID), http.StatusNotFound)
	want(t, "another organisation's certificate", get(prodTok, prodSlug, "/attestations/"+cert.ID), http.StatusNotFound)
	want(t, "another organisation's certificate file", get(prodTok, prodSlug, "/attestations/"+cert.ID+"/file"), http.StatusNotFound)
	file := get(procTok, procSlug, "/attestations/"+cert.ID+"/file")
	want(t, "download certificate", file, http.StatusOK)
	if !bytes.Equal(file.Body.Bytes(), pdfBytes) {
		t.Fatal("downloaded certificate differs from the uploaded file")
	}

	revoke := func(token, slug, reason string) *httptest.ResponseRecorder {
		return post(e, token, as(slug, "/attestations/"+cert.ID+"/revoke"), attestation.RevokeInput{Reason: reason})
	}
	want(t, "revoke another issuer's certificate", revoke(procTok, procSlug, "not ours"), http.StatusNotFound)
	want(t, "revoke without a reason", revoke(certTok, certSlug, " "), http.StatusBadRequest)
	rec = revoke(certTok, certSlug, "Issued in error")
	want(t, "revoke", rec, http.StatusOK)
	if got := decodeTo[attestation.Attestation](t, rec); got.Status != attestation.StatusRevoked || got.RevokedAt == nil {
		t.Fatalf("revoked certificate: %+v", got)
	}
	want(t, "revoke twice", revoke(certTok, certSlug, "again"), http.StatusConflict)

	c.drain()
	if ok, _ := c.contract.Orgs.IsAccredited(opts, chain.MustID(certifier.ID), chain.Key("export_permit")); !ok {
		t.Fatal("certifier is not accredited for export permits on-chain")
	}
	if ok, _ := c.contract.Orgs.IsAccredited(opts, chain.MustID(certifier.ID), chain.Key("phytosanitary_certificate")); ok {
		t.Fatal("certifier is accredited on-chain for a type it was never given")
	}
	if got := decodeTo[[]org.Accreditation](t, get(certTok, certSlug, "/orgs/"+certifier.ID+"/accreditations")); len(got) != 1 ||
		got[0].TxHash == nil || got[0].AnchoredAt == nil {
		t.Fatalf("accreditation not anchored: %+v", got)
	}
	chainCert, err := c.contract.Attestations.Attestation(opts, chain.MustID(cert.ID))
	if err != nil || !chainCert.Revoked || chain.Hex(chainCert.DocHash) != cert.DocHash ||
		chainCert.Subject != chain.MustID(g1.ID) || chainCert.IssuerOrg != chain.MustID(certifier.ID) {
		t.Fatalf("on-chain certificate: %+v (%v)", chainCert, err)
	}
	if got := decodeTo[attestation.Attestation](t, get(procTok, procSlug, "/attestations/"+cert.ID)); got.TxHash == nil ||
		got.RevokeTxHash == nil || *got.TxHash == *got.RevokeTxHash || got.AnchoredAt == nil {
		t.Fatalf("certificate hashes were not written back: %+v", got)
	}

	// ---- Flags: a cleared one changes nothing; a confirmed one revokes the whole tree.
	raise := func(token, slug, passportID, reason string) *httptest.ResponseRecorder {
		return post(e, token, as(slug, "/flags"), flag.RaiseInput{PassportID: passportID, Reason: reason})
	}
	want(t, "flag a passport out of view", raise(certTok, certSlug, b.ID, "looks wrong"), http.StatusNotFound)
	want(t, "flag without a reason", raise(prodTok, prodSlug, b.ID, ""), http.StatusBadRequest)
	rec = raise(prodTok, prodSlug, b.ID, "Weighed twice at the gate")
	want(t, "raise flag", rec, http.StatusCreated)
	mistaken := decodeTo[flag.Flag](t, rec)
	rec = raise(prodTok, prodSlug, a.ID, "This lot never left our plot")
	want(t, "raise second flag", rec, http.StatusCreated)
	upheld := decodeTo[flag.Flag](t, rec)
	if upheld.Status != flag.StatusOpen || upheld.OrganisationID != producer.ID {
		t.Fatalf("flag recorded wrongly: %+v", upheld)
	}

	flags := func(token, slug, query string) int {
		t.Helper()
		rec := get(token, slug, "/flags?"+listq.ParamSearch+"=gate"+query)
		want(t, "list flags", rec, http.StatusOK)
		return decodeTo[listq.Page[flag.Flag]](t, rec).Total
	}
	if flags(prodTok, prodSlug, "") != 1 || flags(procTok, procSlug, "") != 0 || flags(prodTok, prodSlug, "&"+flag.ParamStatus+"="+flag.StatusCleared) != 0 {
		t.Fatal("flag list is not limited to the organisation that raised them, or ignores the status filter")
	}
	// Open and resolved (cleared or confirmed) together make every flag.
	if open, resolved := flags(prodTok, prodSlug, "&"+flag.ParamStatus+"="+flag.StatusOpen),
		flags(prodTok, prodSlug, "&"+flag.ParamStatus+"="+flag.StatusResolved); open+resolved != 1 {
		t.Fatalf("open %d + resolved %d, want 1", open, resolved)
	}
	want(t, "bad flag filter", get(prodTok, prodSlug, "/flags?"+flag.ParamStatus+"=bogus"), http.StatusBadRequest)
	want(t, "another organisation's flag", get(procTok, procSlug, "/flags/"+upheld.ID), http.StatusNotFound)
	want(t, "own flag", get(prodTok, prodSlug, "/flags/"+upheld.ID), http.StatusOK)

	resolve := func(token, slug, id, outcome string) *httptest.ResponseRecorder {
		return post(e, token, at(slug, "/flags/"+id+"/resolve"), flag.ResolveInput{Outcome: outcome, Note: "Checked against the weighbridge log"})
	}
	want(t, "raiser resolves its own flag", resolve(prodTok, prodSlug, upheld.ID, flag.StatusConfirmed), http.StatusForbidden)
	want(t, "unknown outcome", resolve(root, "", upheld.ID, "ignored"), http.StatusBadRequest)
	want(t, "unknown flag", resolve(root, "", fakeOrgID, flag.StatusCleared), http.StatusNotFound)

	rec = resolve(root, "", mistaken.ID, flag.StatusCleared)
	want(t, "clear", rec, http.StatusOK)
	if got := decodeTo[flag.Flag](t, rec); got.Status != flag.StatusCleared || got.Revoked != 0 || got.ResolutionNote == nil {
		t.Fatalf("cleared flag: %+v", got)
	}
	want(t, "resolve twice", resolve(root, "", mistaken.ID, flag.StatusConfirmed), http.StatusConflict)

	hashBefore := *passportOf(g1.ID).TxHash
	rec = resolve(root, "", upheld.ID, flag.StatusConfirmed)
	want(t, "confirm", rec, http.StatusOK)
	family := []string{a.ID, child1.ID, child2.ID, g1.ID, g2.ID}
	if got := decodeTo[flag.Flag](t, rec); got.Status != flag.StatusConfirmed || got.Revoked != len(family) {
		t.Fatalf("confirmed flag revoked %d passports, want %d", got.Revoked, len(family))
	}
	for _, id := range family {
		if got := passportOf(id); got.Status != passport.StatusRevoked {
			t.Fatalf("descendant %s is %s after its ancestor was revoked", id, got.Status)
		}
	}
	if got := passportOf(b.ID); got.Status != passport.StatusActive {
		t.Fatalf("a passport outside the revoked tree is %s", got.Status)
	}
	want(t, "transform a revoked passport", transform(procTok, procSlug, passport.TransformInput{
		Stage: stageSort, Inputs: []passport.TransformSource{src(g2.ID, 1)}, Outputs: outs(1)}), http.StatusConflict)
	want(t, "hand over a revoked passport", post(e, procTok, as(procSlug, "/passports/"+g2.ID+"/handovers"),
		passport.HandoverInput{ToOrganisationID: producer.ID, Stage: stageProcess}), http.StatusConflict)

	c.drain()
	for _, id := range family {
		if status, _ := onChain(id); status != chainRevoked {
			t.Fatalf("on-chain status of %s = %d, want revoked", id, status)
		}
	}
	if status, _ := onChain(b.ID); status != chainActive {
		t.Fatalf("on-chain status of the passport outside the tree = %d", status)
	}
	// The revocation has its own transaction; the passport keeps the hash of the one that made it.
	revokedG1 := passportOf(g1.ID)
	last := revokedG1.Events[len(revokedG1.Events)-1]
	if *revokedG1.TxHash != hashBefore || last.Kind != passport.EventRevoke || last.TxHash == nil || *last.TxHash == hashBefore {
		t.Fatalf("revocation hashes: passport %s, last event %+v", *revokedG1.TxHash, last)
	}
	if got := decodeTo[flag.Flag](t, get(prodTok, prodSlug, "/flags/"+upheld.ID)); got.TxHash == nil || got.AnchoredAt == nil {
		t.Fatalf("flag was not anchored: %+v", got)
	}

	// Nothing this run queued was refused by a contract or left waiting.
	var stuck int
	var detail *string
	e.pool.QueryRow(ctx,
		`SELECT count(*), min(kind || ': ' || COALESCE(last_error, '')) FROM chain_outbox WHERE created_at >= $1 AND status <> $2`,
		started, chain.StatusConfirmed).Scan(&stuck, &detail)
	if stuck != 0 {
		t.Fatalf("%d queued items were not anchored, e.g. %s", stuck, *detail)
	}

	// ---- Notifications: each organisation was told what concerns it, and only that.
	notes := func(token, slug, query string) listq.Page[notify.Notification] {
		t.Helper()
		rec := get(token, slug, "/notifications?"+listq.ParamSize+"=100"+query)
		want(t, "list notifications", rec, http.StatusOK)
		return decodeTo[listq.Page[notify.Notification]](t, rec)
	}
	kinds := func(pg listq.Page[notify.Notification]) map[string]int {
		out := map[string]int{}
		for _, n := range pg.Data {
			out[n.Kind]++
		}
		return out
	}
	procNotes := notes(procTok, procSlug, "")
	for kind, n := range map[string]int{
		notify.KindHandoverStarted:    2,
		notify.KindCertificateIssued:  2, // the two about its passport; the facility is the producer's
		notify.KindCertificateRevoked: 1,
		notify.KindFlagRaised:         2,
		notify.KindPassportRevoked:    len(family),
	} {
		if got := kinds(procNotes)[kind]; got != n {
			t.Errorf("processor has %d %s notifications, want %d", got, kind, n)
		}
	}
	prodKinds := kinds(notes(prodTok, prodSlug, ""))
	if prodKinds[notify.KindHandoverAccepted] != 2 || prodKinds[notify.KindFlagResolved] != 2 ||
		prodKinds[notify.KindCertificateIssued] != 1 || prodKinds[notify.KindPassportRevoked] != 0 {
		t.Errorf("producer's notifications: %v", prodKinds)
	}

	unread := "&" + notify.ParamUnread + "=true"
	first := procNotes.Data[0]
	want(t, "read another organisation's notification", post(e, prodTok, as(prodSlug, "/notifications/"+first.ID+"/read"), nil), http.StatusNotFound)
	rec = post(e, procTok, as(procSlug, "/notifications/"+first.ID+"/read"), nil)
	want(t, "read", rec, http.StatusOK)
	if decodeTo[notify.Notification](t, rec).ReadAt == nil {
		t.Fatal("notification not marked read")
	}
	want(t, "read twice", post(e, procTok, as(procSlug, "/notifications/"+first.ID+"/read"), nil), http.StatusOK)
	if got := notes(procTok, procSlug, unread).Total; got != procNotes.Total-1 {
		t.Fatalf("unread = %d, want %d", got, procNotes.Total-1)
	}

	// ---- E-mail: each notification is sent once, to the organisation's users.
	sent := map[string]int{} // recipient → messages
	var failing bool
	mailer := &notify.Mailer{Pool: e.pool, From: "no-reply@chaintrack.test",
		FromHeader: `"ChainTrack" <no-reply@chaintrack.test>`,
		Send: func(_ string, _ smtp.Auth, from string, to []string, msg []byte) error {
			if failing {
				return errors.New("mail server is down")
			}
			if from != "no-reply@chaintrack.test" {
				t.Errorf("envelope sender should be the bare address, got %q", from)
			}
			if !strings.Contains(string(msg), `From: "ChainTrack" <no-reply@chaintrack.test>`) || len(to) == 0 {
				t.Errorf("malformed e-mail to %v: %s", to, msg)
			}
			for _, addr := range to {
				sent[addr]++
			}
			return nil
		}}
	failing = true
	if err := mailer.SendPending(ctx); err == nil {
		t.Fatal("a mail server that is down was not reported")
	}
	failing = false
	waiting := func() (n int) {
		e.pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE organisation_id = $1::uuid AND emailed_at IS NULL`, processor.ID).Scan(&n)
		return n
	}
	// The mailer takes the oldest pending notifications of every organisation,
	// a batch at a time, skipping rows another mailer holds (the dev API's,
	// when the tests share its database): keep sending until ours are out.
	for deadline := time.Now().Add(30 * time.Second); waiting() > 0; {
		if time.Now().After(deadline) {
			t.Fatal("notifications are still waiting to be e-mailed")
		}
		if err := mailer.SendPending(ctx); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if sent[e.email(tRival)] != procNotes.Total {
		t.Fatalf("processor's user was sent %d e-mails, want %d", sent[e.email(tRival)], procNotes.Total)
	}
	before := sent[e.email(tRival)]
	if err := mailer.SendPending(ctx); err != nil || sent[e.email(tRival)] != before {
		t.Fatalf("notifications were e-mailed twice (err %v)", err)
	}
}

// A template's commodity and unit are copied onto its passports, so they are
// frozen once passports exist; the rest can still change, and every saved
// yield is queued for the chain.
func TestTemplateInUse(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)
	admin, slug := e.login(tAcme), e.slug(tAcme)

	plot := e.approvedFacility(admin, slug, root, facilityInput(newPlot()))

	in := withoutOriginEvidence(coffeeTemplate("coffee_" + e.suffix))
	in.Name, in.Unit = "Coffee "+e.suffix, plot.CapacityUnit
	rec := post(e, root, "/templates", in)
	want(t, "create template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)
	put := func(in template.Input) *httptest.ResponseRecorder {
		body, _ := json.Marshal(in)
		return do(e.h, http.MethodPut, "/templates/"+tpl.ID, root, string(body))
	}

	publish := func() template.Template {
		t.Helper()
		rec := post(e, root, "/templates/"+tpl.ID+"/publish", nil)
		want(t, "publish", rec, http.StatusOK)
		return decodeTo[template.Template](t, rec)
	}
	if tpl.Version != 1 || tpl.Draft != nil {
		t.Fatalf("new template: version %d, draft %v", tpl.Version, tpl.Draft)
	}
	want(t, "publish without a draft", post(e, root, "/templates/"+tpl.ID+"/publish", nil), http.StatusConflict)
	want(t, "member publishes", post(e, admin, as(slug, "/templates/"+tpl.ID+"/publish"), nil), http.StatusForbidden)

	// Not in use yet: the unit may change. An edit waits as a draft until published.
	changed := in
	changed.Unit = "kg"
	rec = put(changed)
	want(t, "change unit before use", rec, http.StatusOK)
	if got := decodeTo[template.Template](t, rec); got.Unit != in.Unit || got.Draft == nil || got.Draft.Unit != "kg" || got.Version != 1 {
		t.Fatalf("saved draft: %s", rec.Body)
	}
	if got := publish(); got.Unit != "kg" || got.Draft != nil || got.Version != 2 {
		t.Fatalf("published: %+v", got)
	}
	want(t, "change it back", put(in), http.StatusOK)
	publish()

	want(t, "issue", post(e, admin, as(slug, "/passports"), passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: 1}), http.StatusCreated)
	want(t, "change unit once in use", put(changed), http.StatusConflict)
	renamed := in
	renamed.Commodity = "tea_" + e.suffix
	want(t, "change commodity once in use", put(renamed), http.StatusConflict)

	stricter := in
	lower := washingYield / 2
	stricter.Stages = slices.Clone(in.Stages)
	stricter.Stages[1].MaxYield = &lower
	want(t, "change a yield once in use", put(stricter), http.StatusOK)
	// Issuing still works by the published version while the draft waits.
	want(t, "issue with a draft waiting", post(e, admin, as(slug, "/passports"), passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: 1}), http.StatusCreated)
	if got := publish(); got.Version != 4 || *got.Stages[1].MaxYield != lower {
		t.Fatalf("published the stricter yield: %+v", got)
	}

	// Each publish (and the create) queues the stage's yield and the split stage's yield of 1.
	statuses, _ := e.outbox(tpl.ID)
	if len(statuses) != 8 {
		t.Fatalf("queued %d yields for the template, want 8", len(statuses))
	}
	var ppm int64
	e.pool.QueryRow(t.Context(),
		`SELECT (payload->>'parts_per_million')::bigint FROM chain_outbox WHERE subject_id = $1::uuid
		 AND payload->>'stage' <> $2 ORDER BY seq DESC LIMIT 1`, tpl.ID, template.StageSplit).Scan(&ppm)
	if ppm != template.YieldPPM(lower) {
		t.Fatalf("last queued yield = %d, want %d", ppm, template.YieldPPM(lower))
	}
}

func TestProvenanceInputValidation(t *testing.T) {
	id := fakeOrgID
	one := []passport.TransformSource{{PassportID: id, Quantity: 1}}
	out := []passport.TransformOutput{{Quantity: 1}}
	many := make([]passport.TransformOutput, passport.MaxTransformItems+1)
	for name, in := range map[string]passport.TransformInput{
		"no stage":         {Inputs: one, Outputs: out},
		"no inputs":        {Stage: stageSort, Outputs: out},
		"no outputs":       {Stage: stageSort, Inputs: one},
		"too many outputs": {Stage: stageSort, Inputs: one, Outputs: many},
		"input not an id":  {Stage: stageSort, Inputs: []passport.TransformSource{{PassportID: "x", Quantity: 1}}, Outputs: out},
		"input used twice": {Stage: stageSort, Inputs: append(slices.Clone(one), passport.TransformSource{PassportID: strings.ToUpper(id), Quantity: 1}), Outputs: out},
		"zero input":       {Stage: stageSort, Inputs: []passport.TransformSource{{PassportID: id}}, Outputs: out},
		"negative output":  {Stage: stageSort, Inputs: one, Outputs: []passport.TransformOutput{{Quantity: -1}}},
	} {
		if in.Validate() == "" {
			t.Errorf("transform: %s accepted", name)
		}
	}
	ok := passport.TransformInput{Stage: " " + stageSort + " ", Inputs: one, Outputs: out}
	if msg := ok.Validate(); msg != "" || ok.Stage != stageSort {
		t.Errorf("valid transform rejected: %q", msg)
	}

	long := strings.Repeat("a", flag.MaxTextLen+1)
	for name, in := range map[string]flag.RaiseInput{
		"no passport": {Reason: "x"}, "no reason": {PassportID: id, Reason: "  "}, "long reason": {PassportID: id, Reason: long},
	} {
		if in.Validate() == "" {
			t.Errorf("flag: %s accepted", name)
		}
	}
	if (&flag.ResolveInput{Outcome: flag.StatusOpen}).Validate() == "" || (&flag.ResolveInput{Outcome: flag.StatusCleared, Note: long}).Validate() == "" {
		t.Error("flag: a bad resolution was accepted")
	}
	if (&attestation.RevokeInput{Reason: strings.Repeat("a", attestation.MaxReasonLen+1)}).Validate() == "" {
		t.Error("certificate: an overlong revoke reason was accepted")
	}

	fields := certificateFields(attestation.SubjectFacility, id)
	if in, msg := attestation.ParseInput(func(k string) string { return fields[k] }); msg != "" || !in.ValidTo.After(in.ValidFrom) {
		t.Errorf("valid certificate fields rejected: %q", msg)
	}
	fields[attestation.FormValidFrom] = "1960-01-01" // the chain cannot hold a time before 1970
	if _, msg := attestation.ParseInput(func(k string) string { return fields[k] }); msg == "" {
		t.Error("certificate: a validity starting before 1970 was accepted")
	}
}

func TestPinataPin(t *testing.T) {
	const token, cid = "test-jwt", "bafybeigdyrtest"
	var got []byte
	answer := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("no file in the request: %v", err)
			return
		}
		got, _ = io.ReadAll(f)
		w.WriteHeader(answer)
		fmt.Fprintf(w, `{"IpfsHash": %q}`, cid)
	}))
	defer srv.Close()

	p := attestation.Pinata{JWT: token, URL: srv.URL}
	out, err := p.Pin(t.Context(), "certificate.pdf", bytes.NewReader(pdfBytes))
	if err != nil || out != cid || !bytes.Equal(got, pdfBytes) {
		t.Fatalf("Pin = %q, %v", out, err)
	}
	answer = http.StatusUnauthorized
	if _, err := p.Pin(t.Context(), "certificate.pdf", bytes.NewReader(pdfBytes)); err == nil {
		t.Fatal("a refused pin was reported as done")
	}
}

func TestMailerAddress(t *testing.T) {
	m := notify.NewMailer(nil, notify.SMTP{Host: "mail.test", Port: "587", User: "no-reply@chaintrack.test", Password: "secret", TLS: true,
		FromName: "ChainTrack", FromAddr: "no-reply@chaintrack.test"})
	if m.Addr != "mail.test:587" || m.From != "no-reply@chaintrack.test" || m.FromHeader != `"ChainTrack" <no-reply@chaintrack.test>` || m.Auth == nil || m.Send == nil {
		t.Fatalf("mail settings read wrongly: %+v", m)
	}
	if m := notify.NewMailer(nil, notify.SMTP{Host: "localhost", Port: "1025", User: "dev@chaintrack.test"}); m.Auth != nil {
		t.Fatal("a mail server without sign-in got credentials")
	}
}

// With TLS required, a mail server that offers no encryption gets nothing.
func TestMailerRefusesPlainServerWhenTLSRequired(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	got := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		fmt.Fprint(conn, "220 plain.test\r\n")
		var seen []string
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				break
			}
			seen = append(seen, strings.TrimSpace(line))
			fmt.Fprint(conn, "250 plain.test\r\n") // EHLO and everything else: fine, no STARTTLS
		}
		got <- strings.Join(seen, "|")
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	m := notify.NewMailer(nil, notify.SMTP{Host: host, Port: port, User: "a@chaintrack.test", TLS: true})
	if err := m.Send(m.Addr, m.Auth, m.From, []string{"b@chaintrack.test"}, []byte("hello")); err == nil {
		t.Fatal("sent over a plain connection although TLS is required")
	}
	if seen := <-got; strings.Contains(seen, "MAIL FROM") {
		t.Fatalf("server received a message: %s", seen)
	}
}

// Yields reach the contract as whole parts per million.
func TestYieldPPM(t *testing.T) {
	for yield, want := range map[float64]int64{0.2: 200_000, 1: 1_000_000, 0.05: 50_000, 0.333333: 333_333} {
		if got := template.YieldPPM(yield); got != want {
			t.Errorf("YieldPPM(%v) = %d, want %d", yield, got, want)
		}
	}
	if chain.YieldScale != 1_000_000 {
		t.Fatal("YieldScale must equal PassportRegistry.YIELD_SCALE")
	}
}
