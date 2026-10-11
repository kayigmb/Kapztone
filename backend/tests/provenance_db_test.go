package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	return decodeTo[struct{ Error string }](t, rec).Error
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
	fIn := facilityInput(plotGeo)
	fIn.Kind = facilityKind
	rec := post(e, prodTok, as(prodSlug, "/facilities"), fIn)
	want(t, "create facility", rec, http.StatusCreated)
	site := decodeTo[facility.Facility](t, rec)

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
	rec = post(e, root, "/templates", tplIn)
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
		rec := post(e, prodTok, as(prodSlug, "/passports/"+id+"/handovers"), passport.HandoverInput{ToOrganisationID: processor.ID})
		want(t, "start handover", rec, http.StatusCreated)
		h := decodeTo[passport.Handover](t, rec)
		want(t, "accept handover", post(e, procTok, as(procSlug, "/handovers/"+h.ID+"/accept"), nil), http.StatusOK)
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
	accredit := func(token, id string, on bool) *httptest.ResponseRecorder {
		return post(e, token, "/orgs/"+id+"/accredit", org.Accreditation{Accredited: on})
	}
	want(t, "an organisation accredits itself", post(e, certTok, as(certSlug, "/orgs/"+certifier.ID+"/accredit"), org.Accreditation{Accredited: true}), http.StatusForbidden)
	want(t, "accredit a producer", accredit(root, producer.ID, true), http.StatusConflict)
	rec = accredit(root, certifier.ID, true)
	want(t, "accredit", rec, http.StatusOK)
	if !decodeTo[org.Organisation](t, rec).Accredited {
		t.Fatal("organisation not marked accredited")
	}
	want(t, "accredit twice", accredit(root, certifier.ID, true), http.StatusOK)
	if statuses, _ := e.outbox(certifier.ID); len(statuses) != 3 { // register, verified, accredited once
		t.Fatalf("queued %d items for the certifier, want 3", len(statuses))
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
	want(t, "get certificate", get(procTok, procSlug, "/attestations/"+cert.ID), http.StatusOK)
	want(t, "unknown certificate", get(procTok, procSlug, "/attestations/"+fakeOrgID), http.StatusNotFound)
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
	if ok, _ := c.contract.Orgs.IsAccredited(opts, chain.MustID(certifier.ID)); !ok {
		t.Fatal("certifier is not accredited on-chain")
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
		passport.HandoverInput{ToOrganisationID: producer.ID}), http.StatusConflict)

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
		Send: func(_ string, _ smtp.Auth, from string, to []string, msg []byte) error {
			if failing {
				return errors.New("mail server is down")
			}
			if !strings.Contains(string(msg), "From: "+from) || len(to) == 0 {
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
	for i := 0; waiting() > 0; i++ {
		if i > 50 {
			t.Fatal("notifications are still waiting to be e-mailed")
		}
		if err := mailer.SendPending(ctx); err != nil {
			t.Fatal(err)
		}
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

	rec := post(e, admin, as(slug, "/facilities"), facilityInput(plotGeo))
	want(t, "create facility", rec, http.StatusCreated)
	plot := decodeTo[facility.Facility](t, rec)

	in := coffeeTemplate("coffee_" + e.suffix)
	in.Name, in.Unit = "Coffee "+e.suffix, plot.CapacityUnit
	rec = post(e, root, "/templates", in)
	want(t, "create template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)
	put := func(in template.Input) *httptest.ResponseRecorder {
		body, _ := json.Marshal(in)
		return do(e.h, http.MethodPut, "/templates/"+tpl.ID, root, string(body))
	}

	// Not in use yet: the unit may change.
	changed := in
	changed.Unit = "kg"
	want(t, "change unit before use", put(changed), http.StatusOK)
	want(t, "change it back", put(in), http.StatusOK)

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

	// One yield per save: create, two changes of unit, and the stricter yield.
	statuses, _ := e.outbox(tpl.ID)
	if len(statuses) != 4 {
		t.Fatalf("queued %d yields for the template, want 4", len(statuses))
	}
	var ppm int64
	e.pool.QueryRow(t.Context(),
		`SELECT (payload->>'parts_per_million')::bigint FROM chain_outbox WHERE subject_id = $1::uuid ORDER BY seq DESC LIMIT 1`, tpl.ID).Scan(&ppm)
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
	m, err := notify.NewMailer(nil, "smtp://user:secret@mail.test:587?from=no-reply@chaintrack.test")
	if err != nil || m.Addr != "mail.test:587" || m.From != "no-reply@chaintrack.test" || m.Auth == nil {
		t.Fatalf("mail address read wrongly: %+v (%v)", m, err)
	}
	if m, err := notify.NewMailer(nil, "smtp://mail.test:25?from=a@b.test"); err != nil || m.Auth != nil {
		t.Fatalf("a mail server without sign-in was rejected: %v", err)
	}
	for _, bad := range []string{"mail.test:587", "smtp://mail.test?from=a@b.test", "smtp://mail.test:587", "https://mail.test:587?from=a@b.test"} {
		if _, err := notify.NewMailer(nil, bad); err == nil {
			t.Errorf("%q accepted as a mail server address", bad)
		}
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
