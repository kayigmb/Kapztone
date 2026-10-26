package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
)

// stageWashing is the coffee template's washing station: it takes handovers
// and transforms, and asks for no evidence.
const stageWashing = "washing_station"

// withoutOriginEvidence drops the evidence the template asks for, for tests
// about something else (TestRequiredEvidence covers evidence).
func withoutOriginEvidence(in template.Input) template.Input {
	in.RequiredEvidence = nil
	return in
}

// passportWorld is a producer with a plot, a processor and a certifier, all verified.
type passportWorld struct {
	e                         *env
	root                      string
	producer, processor       org.Organisation
	certifier                 org.Organisation
	prodTok, procTok, certTok string
	prodSlug, procSlug        string
	certSlug                  string
	plot                      facility.Facility
	chain                     *testChain // also sets the handover signing domain
}

func newPassportWorld(t *testing.T) passportWorld {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	w := passportWorld{e: e, root: e.login(tRoot), chain: e.newChain()}
	w.producer = e.verifiedOrg(tAcme, org.KindProducer, w.root)
	w.processor = e.verifiedOrg(tRival, org.KindProcessor, w.root)
	w.certifier = e.verifiedOrg(tStaff, org.KindCertifier, w.root)
	w.prodTok, w.procTok, w.certTok = e.login(tAcme), e.login(tRival), e.login(tStaff)
	w.prodSlug, w.procSlug, w.certSlug = e.slug(tAcme), e.slug(tRival), e.slug(tStaff)

	w.plot = e.approvedFacility(w.prodTok, w.prodSlug, w.root, facilityInput(newPlot()))
	return w
}

func (w passportWorld) template(t *testing.T, in template.Input) template.Template {
	t.Helper()
	in.Name, in.Unit = in.Commodity, w.plot.CapacityUnit
	rec := post(w.e, w.root, "/templates", in)
	want(t, "create template", rec, http.StatusCreated)
	return decodeTo[template.Template](t, rec)
}

func (w passportWorld) issue(tpl template.Template, quantity int64) *httptest.ResponseRecorder {
	return post(w.e, w.prodTok, as(w.prodSlug, "/passports"),
		passport.IssueInput{FacilityID: w.plot.ID, TemplateID: tpl.ID, Quantity: quantity})
}

// Evidence a template asks for must be a valid certificate on the origin
// facility before a passport is issued; a revoked one does not count.
func TestRequiredEvidence(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	tpl := w.template(t, coffeeTemplate("coffee_"+e.suffix)) // plot_geolocation
	want(t, "accredit", accredit(e, w.root, "/orgs/"+w.certifier.ID+"/accreditations", grant("plot_geolocation")), http.StatusOK)

	missing := func(what string) {
		t.Helper()
		rec := w.issue(tpl, 1000)
		want(t, what, rec, http.StatusConflict)
		if msg := errorOf(t, rec); !strings.Contains(msg, "plot_geolocation") {
			t.Fatalf("%s: %q does not name plot_geolocation", what, msg)
		}
	}
	missing("issue without facility evidence")
	cert := w.certify(t, attestation.SubjectFacility, w.plot.ID, "plot_geolocation")
	want(t, "issue with evidence", w.issue(tpl, 1000), http.StatusCreated)
	want(t, "revoke", post(e, w.certTok, as(w.certSlug, "/attestations/"+cert.ID+"/revoke"),
		attestation.RevokeInput{Reason: "issued in error"}), http.StatusOK)
	missing("issue with a revoked certificate")
}

// certify has the certifier attach a certificate of certType to a subject.
func (w passportWorld) certify(t *testing.T, subjectType, subjectID, certType string) attestation.Attestation {
	t.Helper()
	fields := certificateFields(subjectType, subjectID)
	fields[attestation.FormType] = certType
	rec := w.e.certify(w.certTok, w.certSlug, fields, pdfBytes)
	want(t, "certify "+certType, rec, http.StatusCreated)
	return decodeTo[attestation.Attestation](t, rec)
}

// Lot codes count per organisation per year and work wherever an id does;
// a facility reports the capacity it used this period; a retired passport
// can do nothing more and is anchored.
func TestLotCodesCapacityAndRetire(t *testing.T) {
	w := newPassportWorld(t)
	e, c := w.e, w.chain
	tpl := w.template(t, withoutOriginEvidence(coffeeTemplate("coffee_"+e.suffix)))

	var lots []passport.Passport
	for _, q := range []int64{300_000, 200_000} {
		rec := w.issue(tpl, q)
		want(t, "issue", rec, http.StatusCreated)
		lots = append(lots, decodeTo[passport.Passport](t, rec))
	}
	year := time.Now().UTC().Year()
	for i, p := range lots {
		if wantCode := fmt.Sprintf(passport.CodeFormat, strings.ToUpper(w.prodSlug), year, i+1); p.Code != wantCode {
			t.Fatalf("lot %d code = %q, want %q", i, p.Code, wantCode)
		}
	}

	// The code works in place of the id, in any case.
	rec := do(e.h, http.MethodGet, as(w.prodSlug, "/passports/"+strings.ToLower(lots[1].Code)), w.prodTok, "")
	want(t, "get by code", rec, http.StatusOK)
	if got := decodeTo[passport.Passport](t, rec); got.ID != lots[1].ID {
		t.Fatalf("code found %s, want %s", got.ID, lots[1].ID)
	}
	want(t, "unknown code", do(e.h, http.MethodGet, as(w.prodSlug, "/passports/NOPE-2026-0001"), w.prodTok, ""), http.StatusNotFound)
	rec = do(e.h, http.MethodGet, "/verify/"+lots[0].Code, "", "")
	want(t, "verify by code", rec, http.StatusOK)
	if res := decodeTo[verify.Result](t, rec); res.Passport.ID != lots[0].ID {
		t.Fatalf("verify by code found %s", res.Passport.ID)
	}

	// Capacity used this period, on one facility and in the list.
	rec = do(e.h, http.MethodGet, as(w.prodSlug, "/facilities/"+w.plot.ID), w.prodTok, "")
	want(t, "get facility", rec, http.StatusOK)
	if f := decodeTo[facility.Facility](t, rec); f.CapacityUsed != 500_000 {
		t.Fatalf("capacity used = %d, want 500000", f.CapacityUsed)
	}
	rec = do(e.h, http.MethodGet, as(w.prodSlug, "/facilities"), w.prodTok, "")
	if pg := decodeTo[listq.Page[facility.Facility]](t, rec); len(pg.Data) != 1 || pg.Data[0].CapacityUsed != 500_000 {
		t.Fatalf("facility list: %+v", pg.Data)
	}
	byKind := func(kind string) int {
		rec := do(e.h, http.MethodGet, as(w.prodSlug, "/facilities?kind="+kind), w.prodTok, "")
		want(t, "facilities of kind "+kind, rec, http.StatusOK)
		return decodeTo[listq.Page[facility.Facility]](t, rec).Total
	}
	if byKind(w.plot.Kind) != 1 || byKind(facility.KindWarehouse) != 0 {
		t.Fatal("the kind filter doesn't filter")
	}
	want(t, "unknown kind", do(e.h, http.MethodGet, as(w.prodSlug, "/facilities?kind=castle"), w.prodTok, ""), http.StatusBadRequest)

	// A transform's outputs are counted by the transforming organisation.
	handover := func(id string) {
		t.Helper()
		rec := post(e, w.prodTok, as(w.prodSlug, "/passports/"+id+"/handovers"),
			passport.HandoverInput{ToOrganisationID: w.processor.ID, Stage: stageWashing})
		want(t, "start handover", rec, http.StatusCreated)
		h := decodeTo[passport.Handover](t, rec)
		if h.Stage != stageWashing {
			t.Fatalf("handover stage = %q", h.Stage)
		}
		want(t, "accept", post(e, w.procTok, as(w.procSlug, "/handovers/"+h.ID+"/accept"), received(h)), http.StatusOK)
	}
	handover(lots[0].ID)
	rec = post(e, w.procTok, as(w.procSlug, "/transforms"), passport.TransformInput{Stage: stageWashing,
		Inputs:  []passport.TransformSource{{PassportID: lots[0].ID, Quantity: 300_000}},
		Outputs: []passport.TransformOutput{{Quantity: 30_000}, {Quantity: 20_000}}})
	want(t, "transform", rec, http.StatusCreated)
	tr := decodeTo[passport.Transform](t, rec)
	for i, out := range tr.Outputs {
		if wantCode := fmt.Sprintf(passport.CodeFormat, strings.ToUpper(w.procSlug), year, i+1); out.Code != wantCode {
			t.Fatalf("output %d code = %q, want %q", i, out.Code, wantCode)
		}
	}

	// Retire: only the holder, only once, and nothing after it.
	green := tr.Outputs[0]
	retire := func(token, slug, id string) *httptest.ResponseRecorder {
		return post(e, token, as(slug, "/passports/"+id+"/retire"), nil)
	}
	want(t, "retire by a non-holder", retire(w.prodTok, w.prodSlug, green.ID), http.StatusNotFound)
	rec = retire(w.procTok, w.procSlug, green.ID)
	want(t, "retire", rec, http.StatusOK)
	if p := decodeTo[passport.Passport](t, rec); p.Status != passport.StatusRetired {
		t.Fatalf("retired status = %s", p.Status)
	}
	want(t, "retire again", retire(w.procTok, w.procSlug, green.ID), http.StatusConflict)
	want(t, "hand over a retired lot", post(e, w.procTok, as(w.procSlug, "/passports/"+green.ID+"/handovers"),
		passport.HandoverInput{ToOrganisationID: w.producer.ID, Stage: stageWashing}), http.StatusConflict)
	want(t, "transform a retired lot", post(e, w.procTok, as(w.procSlug, "/transforms"), passport.TransformInput{Stage: stageWashing,
		Inputs:  []passport.TransformSource{{PassportID: green.ID, Quantity: 1}},
		Outputs: []passport.TransformOutput{{Quantity: 1}}}), http.StatusConflict)

	// A passport with a handover awaiting its receiver cannot be retired.
	rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lots[1].ID+"/handovers"),
		passport.HandoverInput{ToOrganisationID: w.processor.ID, Stage: stageWashing})
	want(t, "start handover", rec, http.StatusCreated)
	want(t, "retire with a handover open", retire(w.prodTok, w.prodSlug, lots[1].ID), http.StatusConflict)

	// On-chain, and the public page shows the lot as retired without a warning about it.
	c.drain()
	ctx := context.Background()
	p, err := c.contract.Passports.Passport(nil, chain.MustID(green.ID))
	if err != nil || p.Status != chainStatusRetired {
		t.Fatalf("on-chain status = %d (%v), want retired", p.Status, err)
	}
	var anchored bool
	e.pool.QueryRow(ctx, `SELECT tx_hash IS NOT NULL FROM trace_events WHERE passport_id = $1::uuid AND kind = $2`,
		green.ID, passport.EventRetire).Scan(&anchored)
	if !anchored {
		t.Fatal("retirement not anchored")
	}
	rec = do(e.h, http.MethodGet, "/verify/"+green.Code, "", "")
	want(t, "verify retired", rec, http.StatusOK)
	if res := decodeTo[verify.Result](t, rec); res.Passport.Status != passport.StatusRetired || res.Verdict != verify.Verified {
		t.Fatalf("verify retired: status %s verdict %s %+v", res.Passport.Status, res.Verdict, res.Reasons)
	}
}

// chainStatusRetired is PassportRegistry.Status.Retired.
const chainStatusRetired = 3

// Lots move freely: the template only names the product. No certificate is
// needed to issue, hand over or transform, a handover's stage is any label or
// none, and a transform never gives out more than went in.
func TestFreestyleMoves(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	tpl := w.template(t, coffeeTemplate("coffee_"+e.suffix)) // its stages ask for certificates; none are given
	rec := w.issue(tpl, 1000)
	want(t, "issue without evidence", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)

	part := int64(400)
	var held passport.Handover
	for _, stage := range []string{"", "transport"} {
		rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"),
			passport.HandoverInput{ToOrganisationID: w.processor.ID, Stage: stage, Quantity: &part})
		want(t, "handover labelled "+stage, rec, http.StatusCreated)
		held = decodeTo[passport.Handover](t, rec)
		want(t, "accept", post(e, w.procTok, as(w.procSlug, "/handovers/"+held.ID+"/accept"), received(held)), http.StatusOK)
	}

	transform := func(out int64) *httptest.ResponseRecorder {
		return post(e, w.procTok, as(w.procSlug, "/transforms"), passport.TransformInput{
			Inputs:  []passport.TransformSource{{PassportID: held.PassportID, Quantity: 400}},
			Outputs: []passport.TransformOutput{{Quantity: out}}})
	}
	want(t, "transform giving out more than went in", transform(401), http.StatusConflict)
	rec = transform(400)
	want(t, "transform keeping everything", rec, http.StatusCreated)
	if tr := decodeTo[passport.Transform](t, rec); tr.Stage != passport.DefaultTransformStage {
		t.Fatalf("unlabelled transform stage = %q", tr.Stage)
	}
}

// Every signing action needs the signing PIN; reading does not.
func TestSigningPINOnSigningActions(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	tpl := w.template(t, withoutOriginEvidence(coffeeTemplate("coffee_"+e.suffix)))
	rec := w.issue(tpl, 1000)
	want(t, "issue with the PIN", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)

	withPIN := func(method, path, pin, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, as(w.prodSlug, path), strings.NewReader(body))
		req.Header.Set(auth.HeaderSigningPIN, pin)
		return send(e.h, req, w.prodTok)
	}
	issueBody := fmt.Sprintf(`{"facility_id":%q,"template_id":%q,"quantity":1}`, w.plot.ID, tpl.ID)
	handoverBody := fmt.Sprintf(`{"to_organisation_id":%q,"stage":%q}`, w.processor.ID, stageWashing)
	signing := []struct{ what, method, path, body string }{
		{"issue", http.MethodPost, "/passports", issueBody},
		{"handover", http.MethodPost, "/passports/" + lot.ID + "/handovers", handoverBody},
		{"retire", http.MethodPost, "/passports/" + lot.ID + "/retire", ""},
		{"transform", http.MethodPost, "/transforms", `{"stage":"` + stageWashing + `","inputs":[{"passport_id":"` + lot.ID + `","quantity":1}],"outputs":[{"quantity":1}]}`},
	}
	for _, s := range signing {
		want(t, s.what+" with a wrong PIN", withPIN(s.method, s.path, "000000", s.body), http.StatusForbidden)
	}
	want(t, "read without a PIN check", withPIN(http.MethodGet, "/passports/"+lot.ID, "000000", ""), http.StatusOK)

	// The receiver signs too.
	rec = withPIN(http.MethodPost, "/passports/"+lot.ID+"/handovers", testPIN, handoverBody)
	want(t, "handover with the PIN", rec, http.StatusCreated)
	h := decodeTo[passport.Handover](t, rec)
	accept := httptest.NewRequest(http.MethodPost, as(w.procSlug, "/handovers/"+h.ID+"/accept"), nil)
	accept.Header.Set(auth.HeaderSigningPIN, "000000")
	want(t, "accept with a wrong PIN", send(e.h, accept, w.procTok), http.StatusForbidden)

	// A certificate is signed by its issuer.
	fields := certificateFields(attestation.SubjectPassport, lot.ID)
	e.exec(`UPDATE users SET pin_hash = NULL WHERE email = $1`, e.email(tStaff))
	rec = e.certify(w.certTok, w.certSlug, fields, pdfBytes)
	want(t, "certify without a PIN set", rec, http.StatusForbidden)
	if msg := errorOf(t, rec); msg != auth.MsgPINNotSet {
		t.Fatalf("certify without a PIN: %q", msg)
	}

	// No PIN set at all: told to set one.
	e.exec(`UPDATE users SET pin_hash = NULL WHERE email = $1`, e.email(tAcme))
	rec = withPIN(http.MethodPost, "/passports", testPIN, issueBody)
	want(t, "issue without a PIN set", rec, http.StatusForbidden)
	if msg := errorOf(t, rec); msg != auth.MsgPINNotSet {
		t.Fatalf("no-PIN message = %q", msg)
	}
}
