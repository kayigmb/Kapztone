package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
)

// The organisation's kind caps what its people may do: certifiers and
// regulators inspect and judge, and never trade; traders never certify.
func TestKindsLimitWhatRolesDo(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	regOrg := e.verifiedOrg("reg", org.KindRegulator, w.root)
	reg, regSlug := e.login("reg"), e.slug("reg")
	e.addUser("regmember", auth.RoleMember, regOrg.ID)
	regMember := e.login("regmember")

	tpl := coffee("coffee_" + e.suffix)
	rec := w.issue(tpl, 1000)
	want(t, "producer issues", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)

	// ---- Inspectors never trade.
	issue := passport.IssueInput{FacilityID: w.plot.ID, Commodity: tpl.Commodity, Unit: tpl.Unit, Quantity: 1}
	transform := passport.TransformInput{Stage: stageWashing,
		Inputs: []passport.TransformSource{{PassportID: lot.ID, Quantity: 1}}, Outputs: []passport.TransformOutput{{Quantity: 1}}}
	for _, who := range []struct{ name, tok, slug string }{{"certifier", w.certTok, w.certSlug}, {"regulator", reg, regSlug}} {
		want(t, who.name+" issues", post(e, who.tok, as(who.slug, "/passports"), issue), http.StatusForbidden)
		want(t, who.name+" hands over", post(e, who.tok, as(who.slug, "/passports/"+lot.ID+"/handovers"),
			passport.HandoverInput{ToOrg: w.producer.ID, ToUser: e.personIn(w.producer.ID)}), http.StatusForbidden)
		want(t, who.name+" accepts", post(e, who.tok, as(who.slug, "/handovers/"+fakeOrgID+"/accept"), nil), http.StatusForbidden)
		want(t, who.name+" transforms", post(e, who.tok, as(who.slug, "/transforms"), transform), http.StatusForbidden)
	}
	rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"), passport.HandoverInput{ToOrg: regOrg.ID, ToUser: e.personIn(regOrg.ID), Stage: stageWashing})
	want(t, "hand over to a regulator", rec, http.StatusBadRequest)
	if errorOf(t, rec) != passport.MsgReceiverInvalid {
		t.Fatalf("handover to a regulator: %s", rec.Body)
	}

	// ---- Traders never certify; only producers originate.
	prodFields := certificateFields(attestation.SubjectPassport, lot.ID)
	prodFields[attestation.FormType] = "organic"
	want(t, "producer certifies", e.certify(w.prodTok, w.prodSlug, prodFields, pdfBytes), http.StatusForbidden)
	want(t, "processor issues", post(e, w.procTok, as(w.procSlug, "/passports"), issue), http.StatusForbidden)

	// ---- Only the platform approves facilities; inspectors register their own sites.
	pending := decodeTo[facility.Facility](t, post(e, w.prodTok, as(w.prodSlug, "/facilities"), facilityInput(newPlot())))
	want(t, "certifier approves", post(e, w.certTok, as(w.certSlug, "/facilities/"+pending.ID+"/approve"), nil), http.StatusForbidden)
	want(t, "regulator approves", post(e, reg, as(regSlug, "/facilities/"+pending.ID+"/approve"), nil), http.StatusForbidden)
	want(t, "platform approves", post(e, w.root, "/facilities/"+pending.ID+"/approve", nil), http.StatusOK)
	office := facilityInput(newPlot())
	office.Name = "Regulator office"
	rec = post(e, reg, as(regSlug, "/facilities"), office)
	want(t, "regulator registers its own site", rec, http.StatusCreated)
	regSite := decodeTo[facility.Facility](t, rec)
	want(t, "regulator member registers", post(e, regMember, as(regSlug, "/facilities"), facilityInput(newPlot())), http.StatusForbidden)

	// ---- Inspectors list their own; another organisation's only by lookup or search.
	sees := func(tok, slug, path, id string) bool {
		t.Helper()
		sep := "?"
		if strings.Contains(path, "?") {
			sep = "&"
		}
		rec := do(e.h, http.MethodGet, as(slug, path+sep+"size=100"), tok, "")
		want(t, "list "+path, rec, http.StatusOK)
		for _, x := range decodeTo[listq.Page[struct{ ID string }]](t, rec).Data {
			if x.ID == id {
				return true
			}
		}
		return false
	}
	if sees(reg, regSlug, "/passports", lot.ID) || sees(w.certTok, w.certSlug, "/passports", lot.ID) {
		t.Fatal("an inspector lists another organisation's passports")
	}
	want(t, "regulator looks a lot up by its code", do(e.h, http.MethodGet, as(regSlug, "/passports/"+lot.Code), reg, ""), http.StatusOK)
	if !sees(reg, regSlug, "/facilities", regSite.ID) || sees(reg, regSlug, "/facilities", pending.ID) ||
		sees(reg, regSlug, "/facilities?search="+url.QueryEscape(pending.Name), pending.ID) {
		t.Fatal("a regulator's facilities are not just its own")
	}

	// ---- A regulator sees the flags raised by it or assigned to its people.
	rec = post(e, w.prodTok, as(w.prodSlug, "/flags"), flag.RaiseInput{PassportID: lot.ID, Reason: "Bags weighed twice"})
	want(t, "producer raises a flag", rec, http.StatusCreated)
	raised := decodeTo[flag.Flag](t, rec)
	resolve := `{"outcome":"` + flag.StatusCleared + `","note":"Scales recalibrated"}`
	if sees(reg, regSlug, "/flags", raised.ID) {
		t.Fatal("a regulator sees a flag nobody gave it")
	}
	want(t, "regulator resolves a flag not its own", do(e.h, http.MethodPost, as(regSlug, "/flags/"+raised.ID+"/resolve"), reg, resolve), http.StatusNotFound)
	var regAdmin string
	e.pool.QueryRow(context.Background(), `SELECT id::text FROM users WHERE email = $1`, e.email("reg")).Scan(&regAdmin)
	want(t, "platform assigns the regulator", do(e.h, http.MethodPut, "/flags/"+raised.ID+"/assignee", w.root, `{"user_id":"`+regAdmin+`"}`), http.StatusOK)
	if !sees(reg, regSlug, "/flags", raised.ID) || !sees(regMember, regSlug, "/flags", raised.ID) {
		t.Fatal("the regulator does not see the flag assigned to it")
	}
	want(t, "regulator member resolves", do(e.h, http.MethodPost, as(regSlug, "/flags/"+raised.ID+"/resolve"), regMember, resolve), http.StatusForbidden)
	want(t, "regulator admin resolves", do(e.h, http.MethodPost, as(regSlug, "/flags/"+raised.ID+"/resolve"), reg, resolve), http.StatusOK)

	// ---- A certificate may say which of the issuer's own facilities it was issued at.
	want(t, "accredit", accredit(e, w.root, "/orgs/"+w.certifier.ID+"/accreditations", grant("organic")), http.StatusOK)
	lab := decodeTo[facility.Facility](t, post(e, w.certTok, as(w.certSlug, "/facilities"), facilityInput(newPlot())))
	fields := certificateFields(attestation.SubjectPassport, lot.ID)
	fields[attestation.FormType] = "organic"
	fields[attestation.FormIssuedAt] = w.plot.ID
	want(t, "issued at someone else's facility", e.certify(w.certTok, w.certSlug, fields, pdfBytes), http.StatusBadRequest)
	fields[attestation.FormIssuedAt] = lab.ID
	rec = e.certify(w.certTok, w.certSlug, fields, pdfBytes)
	want(t, "issued at its own lab", rec, http.StatusCreated)
	if c := decodeTo[attestation.Attestation](t, rec); c.IssuedAtFacilityID == nil || *c.IssuedAtFacilityID != lab.ID ||
		c.IssuedAtName == nil || *c.IssuedAtName != lab.Name {
		t.Fatalf("issued-at not recorded: %s", rec.Body)
	}

	// ---- No document uploaded: ChainTrack writes the certificate PDF, and anchors its fingerprint.
	delete(fields, attestation.FormIssuedAt)
	rec = e.certify(w.certTok, w.certSlug, fields, nil)
	want(t, "issue without a document", rec, http.StatusCreated)
	made := decodeTo[attestation.Attestation](t, rec)
	if made.ContentType != "application/pdf" || made.FileName != attestation.DocumentName("organic") {
		t.Fatalf("generated certificate: %s", rec.Body)
	}
	rec = do(e.h, http.MethodGet, as(w.certSlug, "/attestations/"+made.ID+"/file"), w.certTok, "")
	want(t, "download the generated certificate", rec, http.StatusOK)
	sum := sha256.Sum256(rec.Body.Bytes())
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF")) || made.DocHash != "0x"+hex.EncodeToString(sum[:]) {
		t.Fatalf("generated file is not the anchored PDF (hash %s)", made.DocHash)
	}
	if sees(w.procTok, w.procSlug, "/passports", lot.ID) {
		t.Fatal("a processor sees a passport it never touched")
	}

	// ---- A rejected handover: the receiver loses sight, the platform's page keeps it.
	rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"), passport.HandoverInput{ToOrg: w.processor.ID, ToUser: e.personIn(w.processor.ID), Stage: stageWashing})
	want(t, "handover", rec, http.StatusCreated)
	want(t, "reject", post(e, w.procTok, as(w.procSlug, "/handovers/"+decodeTo[passport.Handover](t, rec).ID+"/reject"),
		passport.RejectInput{Reasons: []string{passport.RejectNotExpected}}), http.StatusOK)
	if sees(w.procTok, w.procSlug, "/passports", lot.ID) {
		t.Fatal("the receiver still sees a lot it rejected")
	}
	rec = do(e.h, http.MethodGet, "/passports?size=100&organisation_id="+w.processor.ID, w.root, "")
	want(t, "platform lists the processor's lots", rec, http.StatusOK)
	found := false
	for _, p := range decodeTo[listq.Page[passport.Passport]](t, rec).Data {
		found = found || p.ID == lot.ID
	}
	if !found {
		t.Fatal("the organisation page misses a lot it was offered")
	}

	// ---- The receiver picker lists traders only.
	listed := func(query string) map[string]bool {
		rec := do(e.h, http.MethodGet, as(w.prodSlug, "/orgs/directory?size=100&search="+e.suffix+query), w.prodTok, "")
		want(t, "directory"+query, rec, http.StatusOK)
		out := map[string]bool{}
		for _, l := range decodeTo[listq.Page[org.Listing]](t, rec).Data {
			out[l.ID] = true
		}
		return out
	}
	if all := listed(""); !all[regOrg.ID] || !all[w.certifier.ID] {
		t.Fatal("the directory no longer names inspectors")
	}
	if r := listed("&receivers=true"); r[regOrg.ID] || r[w.certifier.ID] || !r[w.processor.ID] {
		t.Fatalf("receivers: %v", r)
	}
}

// A handover goes to a person of the receiving organisation: they or their
// admins answer it; custody still moves to the organisation. The journey
// names the people.
func TestHandoverToAPerson(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	bob := e.addUser("bob", auth.RoleMember, w.processor.ID)
	e.addUser("eve", auth.RoleMember, w.processor.ID)
	bobTok, eveTok := e.login("bob"), e.login("eve")
	tpl := coffee("coffee_" + e.suffix)
	start := func(toUser string) *httptest.ResponseRecorder {
		t.Helper()
		lot := decodeTo[passport.Passport](t, w.issue(tpl, 100))
		return post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"),
			passport.HandoverInput{ToOrg: w.processor.ID, ToUser: toUser, Stage: stageWashing})
	}

	// The receiver picker: names and roles of a trading organisation's people.
	rec := do(e.h, http.MethodGet, as(w.prodSlug, "/orgs/"+w.processor.ID+"/people"), w.prodTok, "")
	want(t, "people", rec, http.StatusOK)
	people := decodeTo[[]org.Person](t, rec)
	if len(people) != 3 || people[0].Role != auth.RoleOrgAdmin || strings.Contains(rec.Body.String(), "@") {
		t.Fatalf("people: %s", rec.Body)
	}
	want(t, "certifier lists people", do(e.h, http.MethodGet, as(w.certSlug, "/orgs/"+w.processor.ID+"/people"), w.certTok, ""), http.StatusForbidden)

	want(t, "nobody named", start(""), http.StatusBadRequest)
	want(t, "someone of another organisation", start(e.personIn(w.certifier.ID)), http.StatusBadRequest)

	rec = start(bob)
	want(t, "to bob", rec, http.StatusCreated)
	h := decodeTo[passport.Handover](t, rec)
	if h.ToUser == nil || *h.ToUser != bob || h.ToUserName == nil || *h.ToUserName != "bob" || h.FromUserName == nil {
		t.Fatalf("handover names: %s", rec.Body)
	}
	accept := func(tok string) *httptest.ResponseRecorder {
		return post(e, tok, as(w.procSlug, "/handovers/"+h.ID+"/accept"), received(h))
	}
	want(t, "eve accepts bob's handover", accept(eveTok), http.StatusForbidden)
	want(t, "bob accepts", accept(bobTok), http.StatusOK)

	// The passport now names who issued, sent and received it.
	rec = do(e.h, http.MethodGet, as(w.procSlug, "/passports/"+h.PassportID), w.procTok, "")
	want(t, "passport", rec, http.StatusOK)
	var issued, handed bool
	for _, ev := range decodeTo[passport.Passport](t, rec).Events {
		switch ev.Kind {
		case passport.EventIssue:
			issued = ev.By != nil && *ev.By != ""
		case passport.EventHandover:
			handed = ev.By != nil && ev.ReceivedBy != nil && *ev.ReceivedBy == "bob"
		}
	}
	if !issued || !handed {
		t.Fatalf("journey without its people: %s", rec.Body)
	}
	pub := decodeTo[verify.Result](t, do(e.h, http.MethodGet, "/verify/"+h.PassportID, "", ""))
	if pub.Passport.IssuedByName != nil {
		t.Fatal("the public page names who issued the lot")
	}
	for _, ev := range pub.Passport.Events {
		if ev.By != nil || ev.ReceivedBy != nil {
			t.Fatal("the public page names people")
		}
	}

	// The receiving organisation's admin may answer a handover sent to bob.
	h = decodeTo[passport.Handover](t, start(bob))
	want(t, "admin rejects bob's handover", post(e, w.procTok, as(w.procSlug, "/handovers/"+h.ID+"/reject"),
		passport.RejectInput{Reasons: []string{passport.RejectNotExpected}}), http.StatusOK)
}
