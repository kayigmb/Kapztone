package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
)

// A facility originates goods only once approved and while its licence runs;
// every refusal by a rule is counted for the blocked report; the volume
// report counts exports; transforms record the template version they used.
func TestFacilityRulesAndReports(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	ctx := context.Background()
	tplIn := withoutOriginEvidence(coffeeTemplate("coffee_" + e.suffix))
	tpl := w.template(t, tplIn)

	// ---- Pending and expired facilities originate nothing.
	rec := post(e, w.prodTok, as(w.prodSlug, "/facilities"), facilityInput(newPlot()))
	want(t, "second facility", rec, http.StatusCreated)
	pending := decodeTo[facility.Facility](t, rec)
	issueAt := func(facilityID string, q int64) *httptest.ResponseRecorder {
		return post(e, w.prodTok, as(w.prodSlug, "/passports"), passport.IssueInput{FacilityID: facilityID, TemplateID: tpl.ID, Quantity: q})
	}
	rec = issueAt(pending.ID, 10)
	want(t, "issue at a pending facility", rec, http.StatusConflict)
	if errorOf(t, rec) != passport.MsgFacilityPending {
		t.Fatalf("pending facility: %s", rec.Body)
	}
	e.exec(`UPDATE facilities SET licence_valid_to = $2 WHERE id = $1::uuid`, w.plot.ID, time.Now().UTC().AddDate(0, 0, -1).Format(facility.DateLayout))
	rec = issueAt(w.plot.ID, 10)
	want(t, "issue with an expired licence", rec, http.StatusConflict)
	if errorOf(t, rec) != passport.MsgLicenceExpired {
		t.Fatalf("expired licence: %s", rec.Body)
	}
	e.exec(`UPDATE facilities SET licence_valid_to = $2 WHERE id = $1::uuid`, w.plot.ID, time.Now().Format(facility.DateLayout))
	want(t, "issue on the licence's last day", issueAt(w.plot.ID, 10), http.StatusCreated)

	// ---- Refusals by a rule are counted, though their own transaction rolls back.
	want(t, "over capacity", issueAt(w.plot.ID, w.plot.Capacity*1000), http.StatusConflict)

	rec = issueAt(w.plot.ID, 1000)
	want(t, "issue", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)
	rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"), passport.HandoverInput{ToOrg: w.processor.ID, Stage: stageWashing})
	want(t, "handover", rec, http.StatusCreated)
	want(t, "accept", post(e, w.procTok, as(w.procSlug, "/handovers/"+decodeTo[passport.Handover](t, rec).ID+"/accept"), received(decodeTo[passport.Handover](t, rec))), http.StatusOK)
	transform := func(out int64) *httptest.ResponseRecorder {
		return post(e, w.procTok, as(w.procSlug, "/transforms"), passport.TransformInput{Stage: stageWashing,
			Inputs: []passport.TransformSource{{PassportID: lot.ID, Quantity: 500}}, Outputs: []passport.TransformOutput{{Quantity: out}}})
	}
	want(t, "mass balance breach", transform(501), http.StatusConflict)

	// A new template version: the transform records the one it was judged by.
	edited := tpl.Input // the same rules, a new version all the same
	body, _ := json.Marshal(edited)
	want(t, "save a draft", do(e.h, http.MethodPut, "/templates/"+tpl.ID, w.root, string(body)), http.StatusOK)
	want(t, "publish", post(e, w.root, "/templates/"+tpl.ID+"/publish", nil), http.StatusOK)
	rec = transform(50)
	want(t, "transform", rec, http.StatusCreated)
	if tr := decodeTo[passport.Transform](t, rec); tr.TemplateVersion == nil || *tr.TemplateVersion != 2 {
		t.Fatalf("transform template version: %s", rec.Body)
	}

	// ---- Export: a completed handover at a stage that allows EXPORT.
	rec = post(e, w.procTok, as(w.procSlug, "/passports/"+lot.ID+"/handovers"), passport.HandoverInput{ToOrg: w.producer.ID, Stage: "export"})
	want(t, "export handover", rec, http.StatusCreated)
	exp := decodeTo[passport.Handover](t, rec)
	want(t, "accept export", post(e, w.prodTok, as(w.prodSlug, "/handovers/"+exp.ID+"/accept"), received(exp)), http.StatusOK)
	rec = do(e.h, http.MethodGet, as(w.procSlug, "/reports/volume"), w.procTok, "")
	want(t, "volume", rec, http.StatusOK)
	vols := decodeTo[struct{ Data []audit.Volume }](t, rec).Data
	if len(vols) != 1 || vols[0].Exported != 500 || vols[0].Commodity != tpl.Commodity || vols[0].Quantity != 0 {
		t.Fatalf("processor's volume: %+v", vols)
	}

	// A rejected handover counts too.
	rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"), passport.HandoverInput{ToOrg: w.processor.ID, Stage: "import"})
	want(t, "handover to reject", rec, http.StatusCreated)
	want(t, "reject", post(e, w.procTok, as(w.procSlug, "/handovers/"+decodeTo[passport.Handover](t, rec).ID+"/reject"),
		passport.RejectInput{Reasons: []string{passport.RejectNotExpected}}), http.StatusOK)

	// ---- The blocked report, per organisation, as JSON and CSV.
	blocked := func(token, slug string) map[string]int64 {
		t.Helper()
		rec := do(e.h, http.MethodGet, as(slug, "/reports/blocked"), token, "")
		want(t, "blocked", rec, http.StatusOK)
		out := map[string]int64{}
		rows := decodeTo[struct{ Data []audit.Blocked }](t, rec).Data
		if len(rows) != len(audit.Rules) {
			t.Fatalf("blocked report has %d rules, want %d", len(rows), len(audit.Rules))
		}
		for _, b := range rows {
			out[b.Rule] = b.Count
		}
		return out
	}
	if got := blocked(w.prodTok, w.prodSlug); got[audit.RuleCapacity] != 1 || got[audit.RuleEvidence] != 0 ||
		got[audit.RuleMassBalance] != 0 || got[audit.RuleHandoverRejected] != 1 {
		t.Fatalf("producer's blocked report: %v", got)
	}
	if got := blocked(w.procTok, w.procSlug); got[audit.RuleMassBalance] != 1 || got[audit.RuleCapacity] != 0 || got[audit.RuleHandoverRejected] != 1 {
		t.Fatalf("processor's blocked report: %v", got)
	}
	rec = do(e.h, http.MethodGet, as(w.prodSlug, "/reports/blocked?format=csv"), w.prodTok, "")
	want(t, "blocked csv", rec, http.StatusOK)
	if !strings.HasPrefix(rec.Body.String(), strings.Join(audit.BlockedColumns, ",")+"\n"+audit.RuleCapacity+",1\n") {
		t.Fatalf("blocked csv: %q", rec.Body)
	}
	want(t, "blocked in another format", do(e.h, http.MethodGet, as(w.prodSlug, "/reports/blocked?format=xml"), w.prodTok, ""), http.StatusBadRequest)

	// ---- Flags: who raised them, and who looks into them.
	rec = post(e, w.prodTok, as(w.prodSlug, "/flags"), flag.RaiseInput{PassportID: lot.ID, Reason: "looks off"})
	want(t, "raise", rec, http.StatusCreated)
	f := decodeTo[flag.Flag](t, rec)
	if f.RaisedBy != flag.SourceUser || f.AssigneeUserID != nil {
		t.Fatalf("raised flag: %s", rec.Body)
	}
	var rootID, producerAdmin string
	e.pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email = $1`, e.email(tRoot)).Scan(&rootID)
	e.pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email = $1`, e.email(tAcme)).Scan(&producerAdmin)
	assign := func(token, slug, body string) *httptest.ResponseRecorder {
		return do(e.h, http.MethodPut, at(slug, "/flags/"+f.ID+"/assignee"), token, body)
	}
	want(t, "a business assigns", assign(w.prodTok, w.prodSlug, `{"user_id":"`+rootID+`"}`), http.StatusForbidden)
	want(t, "assign a business user", assign(w.root, "", `{"user_id":"`+producerAdmin+`"}`), http.StatusBadRequest)
	want(t, "assign not an id", assign(w.root, "", `{"user_id":"x"}`), http.StatusBadRequest)
	want(t, "assign an unknown flag", do(e.h, http.MethodPut, "/flags/"+fakeOrgID+"/assignee", w.root, `{"user_id":null}`), http.StatusNotFound)
	rec = assign(w.root, "", `{"user_id":"`+rootID+`"}`)
	want(t, "assign", rec, http.StatusOK)
	if got := decodeTo[flag.Flag](t, rec); got.AssigneeUserID == nil || *got.AssigneeUserID != rootID || got.AssigneeName == nil || *got.AssigneeName != tRoot {
		t.Fatalf("assigned flag: %s", rec.Body)
	}
	rec = assign(w.root, "", `{"user_id":null}`)
	want(t, "unassign", rec, http.StatusOK)
	if got := decodeTo[flag.Flag](t, rec); got.AssigneeUserID != nil || got.AssigneeName != nil {
		t.Fatalf("unassigned flag: %s", rec.Body)
	}
}

// at attaches the acting organisation; "" acts for none (the platform).
func at(slug, path string) string {
	if slug == "" {
		return path
	}
	return as(slug, path)
}
