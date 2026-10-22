package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// Two tiny pictures: enough for the content check, which reads the first bytes.
var (
	photoA = []byte("\x89PNG\r\n\x1a\n photo of the bags")
	photoB = []byte("\x89PNG\r\n\x1a\n photo of the seal")
)

// handoverForm sends a handover half (start or accept) as a multipart form with photos.
func handoverForm(e *env, token, path string, fields map[string]string, photos ...[]byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		mw.WriteField(k, v)
	}
	for i, p := range photos {
		fw, _ := mw.CreateFormFile(passport.FormPhotos, fmt.Sprintf("photo-%d.png", i))
		fw.Write(p)
	}
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return send(e.h, req, token)
}

// sameRecord mirrors the passport package's handover record, to check what its data hash covers.
type sameRecord struct {
	ID         string   `json:"id"`
	PassportID string   `json:"passport_id"`
	FromOrg    string   `json:"from_org"`
	ToOrg      string   `json:"to_org"`
	Quantity   int64    `json:"quantity"`
	Note       string   `json:"note"`
	Stage      string   `json:"stage,omitempty"`
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
	Photos     []string `json:"photos,omitempty"`
}

func ruleCount(e *env, rule, orgID string) (n int) {
	e.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM rule_events WHERE rule = $1 AND organisation_id = $2::uuid`, rule, orgID).Scan(&n)
	return n
}

// Handing over part of a lot splits it first; the receiver weighs what
// arrived, within the stage's tolerance; photos and the place go with it;
// a rejection says why and may flag the passport.
func TestPartialHandoverToleranceAndPhotos(t *testing.T) {
	w := newPassportWorld(t)
	e, c := w.e, w.chain
	ctx := context.Background()
	tplIn := withoutOriginEvidence(coffeeTemplate("coffee_" + e.suffix))
	tolerance := 2.0
	tplIn.Stages[1].HandoverTolerance = &tolerance
	tpl := w.template(t, tplIn)

	rec := w.issue(tpl, 1000)
	want(t, "issue", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)
	start := as(w.prodSlug, "/passports/"+lot.ID+"/handovers")
	fields := map[string]string{
		passport.FieldToOrg: w.processor.ID, passport.FieldStage: stageWashing, passport.FieldQuantity: "400",
		passport.FieldNote: "at the gate", passport.FieldLat: "-1.95", passport.FieldLon: "30.06",
	}

	// Refusals first: none of them leaves anything behind.
	tooMany := make([][]byte, passport.MaxPhotos+1)
	for i := range tooMany {
		tooMany[i] = photoA
	}
	want(t, "too many photos", handoverForm(e, w.prodTok, start, fields, tooMany...), http.StatusBadRequest)
	want(t, "a photo that is a PDF", handoverForm(e, w.prodTok, start, fields, pdfBytes), http.StatusBadRequest)
	bad := map[string]string{passport.FieldToOrg: w.processor.ID, passport.FieldStage: stageWashing, passport.FieldQuantity: "lots"}
	want(t, "quantity in words", handoverForm(e, w.prodTok, start, bad), http.StatusBadRequest)
	bad = map[string]string{passport.FieldToOrg: w.processor.ID, passport.FieldStage: stageWashing, passport.FieldLat: "-1.95"}
	want(t, "lat without lon", handoverForm(e, w.prodTok, start, bad), http.StatusBadRequest)
	more := int64(1001)
	want(t, "more than the lot holds", post(e, w.prodTok, start,
		passport.HandoverInput{ToOrg: w.processor.ID, Stage: stageWashing, Quantity: &more}), http.StatusConflict)

	// ---- Part of the lot, with two photos and where it happened.
	rec = handoverForm(e, w.prodTok, start, fields, photoA, photoB)
	want(t, "partial handover", rec, http.StatusCreated)
	h := decodeTo[passport.Handover](t, rec)

	// The boxes each side sees it in (?box=), and a box that doesn't exist.
	boxed := func(tok, slug, box string) int {
		rec := do(e.h, http.MethodGet, as(slug, "/handovers?box="+box), tok, "")
		want(t, "handovers in "+box, rec, http.StatusOK)
		return decodeTo[listq.Page[passport.Handover]](t, rec).Total
	}
	if boxed(w.procTok, w.procSlug, passport.BoxIncoming) != 1 || boxed(w.prodTok, w.prodSlug, passport.BoxOutgoing) != 1 ||
		boxed(w.prodTok, w.prodSlug, passport.BoxIncoming) != 0 || boxed(w.procTok, w.procSlug, passport.BoxDone) != 0 {
		t.Fatal("a waiting handover is in the wrong box")
	}
	want(t, "unknown box", do(e.h, http.MethodGet, as(w.procSlug, "/handovers?box=inbox"), w.procTok, ""), http.StatusBadRequest)

	if h.Quantity != 400 || h.ChildPassportID == nil || *h.ChildPassportID == lot.ID || h.PassportID != *h.ChildPassportID ||
		h.SplitFrom == nil || *h.SplitFrom != lot.ID || h.Location == nil || h.Location.Lat != -1.95 || len(h.Photos) != 2 ||
		h.TemplateVersion == nil || *h.TemplateVersion != 1 {
		t.Fatalf("partial handover: %s", rec.Body)
	}
	shas := []string{}
	for i, p := range [][]byte{photoA, photoB} {
		sum := sha256.Sum256(p)
		if !strings.HasSuffix(h.Photos[i].SHA256, hex.EncodeToString(sum[:])) || h.Photos[i].Side != passport.PhotoBySender {
			t.Fatalf("photo %d recorded wrongly: %+v", i, h.Photos[i])
		}
		shas = append(shas, h.Photos[i].SHA256)
	}
	lat, lon := -1.95, 30.06
	hash, _ := chain.DataHash(sameRecord{h.ID, h.PassportID, w.producer.ID, w.processor.ID, 400, "at the gate", stageWashing, &lat, &lon, shas})
	if h.DataHash != chain.Hex(hash) {
		t.Fatal("the handover's data hash does not cover its place and photos")
	}
	child := *h.ChildPassportID
	parent := decodeTo[passport.Passport](t, do(e.h, http.MethodGet, as(w.prodSlug, "/passports/"+lot.ID), w.prodTok, ""))
	if parent.Remaining != 600 || parent.Status != passport.StatusActive || parent.OrganisationID != w.producer.ID {
		t.Fatalf("the rest of the lot: %+v", parent)
	}

	// ---- Photos: the two parties download them, nobody else.
	photoPath := "/handovers/" + h.ID + "/photos/" + h.Photos[0].ID
	for _, side := range []struct{ token, slug string }{{w.prodTok, w.prodSlug}, {w.procTok, w.procSlug}} {
		rec := do(e.h, http.MethodGet, as(side.slug, photoPath), side.token, "")
		want(t, "party downloads a photo", rec, http.StatusOK)
		if !bytes.Equal(rec.Body.Bytes(), photoA) {
			t.Fatal("downloaded photo differs from the upload")
		}
	}
	want(t, "outsider downloads a photo", do(e.h, http.MethodGet, as(w.certSlug, photoPath), w.certTok, ""), http.StatusNotFound)
	want(t, "photo of another handover", do(e.h, http.MethodGet,
		as(w.prodSlug, "/handovers/"+fakeOrgID+"/photos/"+h.Photos[0].ID), w.prodTok, ""), http.StatusNotFound)
	want(t, "photo id not an id", do(e.h, http.MethodGet, as(w.prodSlug, "/handovers/"+h.ID+"/photos/x"), w.prodTok, ""), http.StatusNotFound)

	// ---- The receiver weighs it: outside the 2% tolerance is refused, and counted.
	acceptPath := as(w.procSlug, "/handovers/"+h.ID+"/accept")
	want(t, "accept without a weight", post(e, w.procTok, acceptPath, passport.AcceptInput{}), http.StatusBadRequest)
	short := int64(380)
	rec = post(e, w.procTok, acceptPath, passport.AcceptInput{ReceivedQuantity: &short})
	want(t, "accept 5% short", rec, http.StatusConflict)
	off := decodeTo[struct {
		Difference, Tolerance float64
	}](t, rec)
	if off.Difference != 5 || off.Tolerance != tolerance {
		t.Fatalf("tolerance refusal: %s", rec.Body)
	}
	if n := ruleCount(e, audit.RuleTolerance, w.processor.ID); n != 1 {
		t.Fatalf("tolerance refusals counted: %d, want 1", n)
	}
	rec = handoverForm(e, w.procTok, acceptPath, map[string]string{
		passport.FieldReceivedQuantity: "399", passport.FieldLat: "-1.96", passport.FieldLon: "30.07"}, photoB)
	want(t, "accept within tolerance", rec, http.StatusOK)
	done := decodeTo[passport.Handover](t, rec)
	if done.Status != passport.HandoverCompleted || done.ReceivedQuantity == nil || *done.ReceivedQuantity != 399 ||
		done.ReceiverLocation == nil || len(done.Photos) != 3 || done.Photos[2].Side != passport.PhotoByReceiver {
		t.Fatalf("accepted handover: %s", rec.Body)
	}
	var flags int
	e.pool.QueryRow(ctx, `SELECT count(*) FROM flags WHERE passport_id = $1::uuid`, child).Scan(&flags)
	if flags != 0 {
		t.Fatal("a difference well inside the tolerance raised a flag")
	}

	// ---- Both passports anchor on the chain: the rest with the producer, the part with the processor.
	c.drain()
	opts := &bind.CallOpts{Context: ctx}
	onParent, err := c.contract.Passports.Passport(opts, chain.MustID(lot.ID))
	if err != nil || onParent.Remaining.Int64() != 600 || onParent.Holder != chain.MustID(w.producer.ID) || onParent.Status != chainActive {
		t.Fatalf("on-chain rest of the lot: %+v %v", onParent, err)
	}
	onChild, err := c.contract.Passports.Passport(opts, chain.MustID(child))
	if err != nil || onChild.Remaining.Int64() != 400 || onChild.Holder != chain.MustID(w.processor.ID) {
		t.Fatalf("on-chain handed part: %+v %v", onChild, err)
	}
	if got, _ := c.contract.Passports.MaxYield(opts, chain.Key(tpl.Commodity), chain.Key(template.StageSplit)); got.Int64() != chain.YieldScale {
		t.Fatalf("on-chain split yield = %s", got)
	}

	// ---- The whole of what is left: no split, as before.
	rec = post(e, w.prodTok, start, passport.HandoverInput{ToOrg: w.processor.ID, Stage: stageWashing})
	want(t, "full handover", rec, http.StatusCreated)
	full := decodeTo[passport.Handover](t, rec)
	if full.Quantity != 600 || full.ChildPassportID != nil || full.SplitFrom != nil || full.PassportID != lot.ID || len(full.Photos) != 0 {
		t.Fatalf("full handover: %s", rec.Body)
	}

	// ---- 1.25% off: accepted, but more than half the tolerance, so the rules raise a flag.
	received := int64(600 - 600*125/10000 - 1) // 592: 1.33% off
	rec = post(e, w.procTok, as(w.procSlug, "/handovers/"+full.ID+"/accept"), passport.AcceptInput{ReceivedQuantity: &received})
	want(t, "accept with a large difference", rec, http.StatusOK)
	var source string
	e.pool.QueryRow(ctx, `SELECT source FROM flags WHERE passport_id = $1::uuid`, lot.ID).Scan(&source)
	if source != flag.SourceRules {
		t.Fatalf("rules flag on a large difference: source %q", source)
	}

	// ---- Reject: from the declared reasons, and with also_flag a flag in the same transaction.
	back := post(e, w.procTok, as(w.procSlug, "/passports/"+child+"/handovers"),
		passport.HandoverInput{ToOrg: w.producer.ID, Stage: stageWashing})
	want(t, "hand back", back, http.StatusCreated)
	r := "/handovers/" + decodeTo[passport.Handover](t, back).ID + "/reject"
	for what, in := range map[string]passport.RejectInput{
		"no reasons":      {Note: "wet"},
		"unknown reason":  {Reasons: []string{"too_heavy"}},
		"reason twice":    {Reasons: []string{passport.RejectDamaged, passport.RejectDamaged}},
		"note too long":   {Reasons: []string{passport.RejectDamaged}, Note: strings.Repeat("x", passport.MaxNoteLen+1)},
		"reasons as text": {Reason: "damaged"},
	} {
		want(t, "reject: "+what, post(e, w.prodTok, as(w.prodSlug, r), in), http.StatusBadRequest)
	}
	rec = post(e, w.prodTok, as(w.prodSlug, r), passport.RejectInput{
		Reasons: []string{passport.RejectWeightShort, passport.RejectSealBroken}, Note: "seal cut", AlsoFlag: true})
	want(t, "reject and flag", rec, http.StatusOK)
	rejected := decodeTo[passport.Handover](t, rec)
	if rejected.Status != passport.HandoverRejected || len(rejected.RejectReasons) != 2 || rejected.RejectReason == nil || *rejected.RejectReason != "seal cut" {
		t.Fatalf("rejected handover: %s", rec.Body)
	}
	var flagID, reason string
	e.pool.QueryRow(ctx, `SELECT id::text, reason, source FROM flags WHERE passport_id = $1::uuid AND organisation_id = $2::uuid`,
		child, w.producer.ID).Scan(&flagID, &reason, &source)
	if flagID == "" || source != flag.SourceUser || !strings.Contains(reason, passport.RejectSealBroken) || !strings.Contains(reason, "seal cut") {
		t.Fatalf("flag raised with the rejection: %q %q", reason, source)
	}
	var notes int
	e.pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE subject_id = $1::uuid AND kind = $2`,
		flagID, notify.KindFlagRaised).Scan(&notes)
	if notes != 2 { // the holder, and the platform
		t.Fatalf("notifications of the flag = %d, want 2", notes)
	}
}

// The receiver sees a lot only while its handover awaits them: once rejected,
// the lot is not theirs to follow.
func TestRejectedReceiverLosesSight(t *testing.T) {
	w := newPassportWorld(t)
	e := w.e
	tpl := w.template(t, withoutOriginEvidence(coffeeTemplate("coffee_"+e.suffix)))
	rec := w.issue(tpl, 1000)
	want(t, "issue", rec, http.StatusCreated)
	lot := decodeTo[passport.Passport](t, rec)

	rec = post(e, w.prodTok, as(w.prodSlug, "/passports/"+lot.ID+"/handovers"), passport.HandoverInput{ToOrg: w.processor.ID, Stage: stageWashing})
	want(t, "handover", rec, http.StatusCreated)
	h := decodeTo[passport.Handover](t, rec)
	view := func() int { return do(e.h, http.MethodGet, as(w.procSlug, "/passports/"+lot.ID), w.procTok, "").Code }
	if got := view(); got != http.StatusOK {
		t.Fatalf("receiver of an awaiting handover = %d, want 200", got)
	}
	want(t, "reject", post(e, w.procTok, as(w.procSlug, "/handovers/"+h.ID+"/reject"),
		passport.RejectInput{Reasons: []string{passport.RejectNotExpected}}), http.StatusOK)
	if got := view(); got != http.StatusNotFound {
		t.Fatalf("receiver of a rejected handover = %d, want 404", got)
	}
}
