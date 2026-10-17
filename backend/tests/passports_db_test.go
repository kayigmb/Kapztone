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

// withoutOriginEvidence drops the evidence the origin stage asks for, for
// tests about something else (TestRequiredEvidence covers evidence).
func withoutOriginEvidence(in template.Input) template.Input {
	in.Stages = append([]template.Stage(nil), in.Stages...)
	in.Stages[0].RequiredEvidence = nil
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

	rec := post(e, w.prodTok, as(w.prodSlug, "/facilities"), facilityInput(plotGeo))
	want(t, "create facility", rec, http.StatusCreated)
	w.plot = decodeTo[facility.Facility](t, rec)
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
		want(t, "accept", post(e, w.procTok, as(w.procSlug, "/handovers/"+h.ID+"/accept"), nil), http.StatusOK)
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

// Evidence a stage asks for must be a valid certificate before the action
// that enters the stage: at issue on the facility, on the passport before a
// handover, on every input before a transform.
func TestRequiredEvidence(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	in := coffeeTemplate("coffee_" + e.suffix) // origin: plot_geolocation; export: export_permit, certificate_of_origin
	in.Stages[1].RequiredEvidence = []string{"moisture_reading"}
	tpl := w.template(t, in)
	want(t, "accredit", accredit(e, w.root, "/orgs/"+w.certifier.ID+"/accreditations",
		grant("plot_geolocation"), grant("moisture_reading"), grant("export_permit"), grant("certificate_of_origin")), http.StatusOK)

	missing := func(what string, rec *httptest.ResponseRecorder, keys ...string) {
		t.Helper()
		want(t, what, rec, http.StatusConflict)
		msg := errorOf(t, rec)
		for _, k := range keys {
			if !strings.Contains(msg, k) {
				t.Fatalf("%s: %q does not name %s", what, msg, k)
			}
		}
	}

	// Issue: the facility needs its plot geolocation.
	missing("issue without facility evidence", w.issue(tpl, 1000), "plot_geolocation")
	w.certify(t, attestation.SubjectFacility, w.plot.ID, "plot_geolocation")
	rec := w.issue(tpl, 1000)
	want(t, "issue with evidence", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)

	handover := func(stage string) *httptest.ResponseRecorder {
		return post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"),
			passport.HandoverInput{ToOrganisationID: w.processor.ID, Stage: stage})
	}
	want(t, "handover to an unknown stage", handover("moon"), http.StatusBadRequest)
	want(t, "handover to a stage without handovers", handover("origin"), http.StatusBadRequest)
	missing("handover to export without evidence", handover("export"), "export_permit", "certificate_of_origin")

	// A revoked certificate does not count.
	permit := w.certify(t, attestation.SubjectPassport, lot.ID, "export_permit")
	w.certify(t, attestation.SubjectPassport, lot.ID, "certificate_of_origin")
	want(t, "revoke permit", post(e, w.certTok, as(w.certSlug, "/attestations/"+permit.ID+"/revoke"),
		attestation.RevokeInput{Reason: "issued in error"}), http.StatusOK)
	missing("handover with a revoked permit", handover("export"), "export_permit")

	// Handover into the washing station needs the moisture reading on the passport.
	missing("handover to washing without moisture", handover(stageWashing), "moisture_reading")
	w.certify(t, attestation.SubjectPassport, lot.ID, "moisture_reading")
	rec = handover(stageWashing)
	want(t, "handover with evidence", rec, http.StatusCreated)
	h := decodeTo[passport.Handover](t, rec)
	want(t, "accept", post(e, w.procTok, as(w.procSlug, "/handovers/"+h.ID+"/accept"), nil), http.StatusOK)

	// Transform: every input needs the stage's evidence. A second lot without it blocks the whole transform.
	rec = w.issue(tpl, 500)
	want(t, "issue second", rec, http.StatusCreated)
	second := decodeTo[passport.Passport](t, rec)
	e.exec(`UPDATE passports SET organisation_id = $2::uuid WHERE id = $1::uuid`, second.ID, w.processor.ID) // as if handed over
	transform := func(ids ...string) *httptest.ResponseRecorder {
		in := passport.TransformInput{Stage: stageWashing, Outputs: []passport.TransformOutput{{Quantity: 100}}}
		for _, id := range ids {
			in.Inputs = append(in.Inputs, passport.TransformSource{PassportID: id, Quantity: 500})
		}
		return post(e, w.procTok, as(w.procSlug, "/transforms"), in)
	}
	missing("transform with an input lacking evidence", transform(lot.ID, second.ID), "moisture_reading")
	want(t, "transform with evidence", transform(lot.ID), http.StatusCreated)
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
