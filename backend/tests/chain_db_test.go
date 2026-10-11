package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/chain/bindings"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// testChain is an in-process chain with the contracts deployed, and a relayer
// pointed at it. A transaction is mined as soon as it is sent.
type testChain struct {
	t        *testing.T
	backend  *minedOnSend
	contract chain.Contracts
	relayer  *chain.Relayer
}

// minedOnSend mines a block for every transaction, and can pretend the chain is unreachable.
type minedOnSend struct {
	simulated.Client
	sim  *simulated.Backend
	down bool
}

var errChainDown = errors.New("dial tcp: connection refused")

func (m *minedOnSend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if m.down {
		return errChainDown
	}
	if err := m.Client.SendTransaction(ctx, tx); err != nil {
		return err
	}
	m.sim.Commit()
	return nil
}

func (m *minedOnSend) CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	if m.down {
		return nil, errChainDown
	}
	return m.Client.CallContract(ctx, call, block)
}

func (m *minedOnSend) HeaderByNumber(ctx context.Context, n *big.Int) (*types.Header, error) {
	if m.down {
		return nil, errChainDown
	}
	return m.Client.HeaderByNumber(ctx, n)
}

// newChain deploys the contracts the way contracts/scripts/deploy.ts does and
// points the test environment's handover signing at them.
func (e *env) newChain() *testChain {
	e.t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		e.t.Fatal(err)
	}
	platform := crypto.PubkeyToAddress(key.PublicKey)
	funds, _ := new(big.Int).SetString("1000000000000000000000", 10)
	sim := simulated.NewBackend(types.GenesisAlloc{platform: {Balance: funds}})
	e.t.Cleanup(func() { sim.Close() })

	backend := &minedOnSend{Client: sim.Client(), sim: sim}
	chainID, err := backend.ChainID(context.Background())
	if err != nil {
		e.t.Fatal(err)
	}
	signer, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		e.t.Fatal(err)
	}
	must := func(_ *types.Transaction, err error) {
		e.t.Helper()
		if err != nil {
			e.t.Fatal(err)
		}
	}
	deployed := func(addr common.Address, tx *types.Transaction, err error) common.Address {
		e.t.Helper()
		must(tx, err)
		return addr
	}

	var a chain.Addresses
	orgAddr, tx, _, err := bindings.DeployOrgRegistry(signer, backend, platform, platform)
	a.Orgs = deployed(orgAddr, tx, err)
	facAddr, tx, facilities, err := bindings.DeployFacilityRegistry(signer, backend, a.Orgs)
	a.Facilities = deployed(facAddr, tx, err)
	passAddr, tx, passports, err := bindings.DeployPassportRegistry(signer, backend, a.Orgs, a.Facilities)
	a.Passports = deployed(passAddr, tx, err)
	custAddr, tx, _, err := bindings.DeployCustodyManager(signer, backend, a.Orgs, a.Passports)
	a.Custody = deployed(custAddr, tx, err)
	attAddr, tx, _, err := bindings.DeployAttestationRegistry(signer, backend, a.Orgs)
	a.Attestations = deployed(attAddr, tx, err)
	must(facilities.SetOriginator(signer, a.Passports))
	must(passports.SetCustodian(signer, a.Custody))

	contracts, err := chain.Bind(a, backend)
	if err != nil {
		e.t.Fatal(err)
	}
	e.passports.Store.Domain = chain.Domain{ChainID: chainID, Contract: a.Custody}
	e.verify.Contracts = contracts
	return &testChain{
		t: e.t, backend: backend, contract: contracts,
		relayer: &chain.Relayer{
			Pool: e.pool, Backend: backend, Signer: signer, Contracts: contracts,
			NotBefore: time.Now().Add(-time.Second), // never touch items queued before this test
		},
	}
}

// drain runs the relayer until the queue is empty and returns how many items it handled.
func (c *testChain) drain() int {
	c.t.Helper()
	for n := 0; ; n++ {
		worked, err := c.relayer.ProcessNext(context.Background())
		if err != nil {
			c.t.Fatalf("relayer: %v", err)
		}
		if !worked {
			return n
		}
	}
}

// outbox returns status and last error of the queued items for a subject, oldest first.
func (e *env) outbox(subjectID string) (statuses, errs []string) {
	e.t.Helper()
	rows, err := e.pool.Query(context.Background(),
		`SELECT status, COALESCE(last_error, '') FROM chain_outbox WHERE subject_id = $1::uuid ORDER BY seq`, subjectID)
	if err != nil {
		e.t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var s, le string
		rows.Scan(&s, &le)
		statuses, errs = append(statuses, s), append(errs, le)
	}
	return statuses, errs
}

// verifiedOrg registers an organisation, uploads a document and has root verify it.
func (e *env) verifiedOrg(name, kind, root string) org.Organisation {
	e.t.Helper()
	o := e.register(name, kind)
	want(e.t, "upload "+name, e.upload(e.login(name), e.slug(name), o.ID, org.DocLicence, pdfBytes), http.StatusCreated)
	body, _ := json.Marshal(org.Review{Decision: org.StatusVerified})
	rec := do(e.h, http.MethodPost, "/orgs/"+o.ID+"/verify", root, string(body))
	want(e.t, "verify "+name, rec, http.StatusOK)
	return decodeTo[org.Organisation](e.t, rec)
}

func post(e *env, token, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	return do(e.h, http.MethodPost, path, token, string(b))
}

func TestSigningKeyIsCreatedAtVerification(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)

	pending := e.register(tAcme, org.KindProducer)
	if pending.SigningAddress != nil {
		t.Fatal("an unverified organisation has a signing key")
	}

	verified := e.verifiedOrg(tRival, org.KindExporter, root)
	if verified.SigningAddress == nil || !common.IsHexAddress(*verified.SigningAddress) {
		t.Fatalf("verified organisation has no signing address: %+v", verified)
	}

	// The stored key is sealed; opened, it signs as the stored address.
	var sealed string
	e.pool.QueryRow(t.Context(), `SELECT signing_key FROM organisations WHERE id = $1::uuid`, verified.ID).Scan(&sealed)
	if _, err := chain.ParseKey(sealed); err == nil {
		t.Fatal("signing key is stored unsealed")
	}
	hexKey, err := testBox(t).Open(sealed)
	if err != nil {
		t.Fatal(err)
	}
	key, err := chain.ParseKey(hexKey)
	if err != nil {
		t.Fatal(err)
	}
	digest := chain.Key("anything")
	sig, _ := chain.Sign(digest, key)
	if signer, _ := chain.Recover(digest, sig); signer != common.HexToAddress(*verified.SigningAddress) {
		t.Fatalf("stored key signs as %s, stored address is %s", signer, *verified.SigningAddress)
	}

	// The key never leaves the API.
	for _, path := range []string{"/orgs/" + verified.ID, "/orgs?" + listq.ParamSearch + "=" + e.suffix} {
		body := do(e.h, http.MethodGet, path, root, "").Body.String()
		if strings.Contains(body, "signing_key") || strings.Contains(body, sealed) || strings.Contains(body, hexKey) {
			t.Fatalf("GET %s exposes the signing key", path)
		}
	}

	// Rejecting and verifying again keeps the same key: the on-chain identity must not change.
	want(t, "reject", post(e, root, "/orgs/"+verified.ID+"/verify", org.Review{Decision: org.StatusRejected, Reason: "licence lapsed"}), http.StatusOK)
	again := post(e, root, "/orgs/"+verified.ID+"/verify", org.Review{Decision: org.StatusVerified})
	want(t, "verify again", again, http.StatusOK)
	if got := decodeTo[org.Organisation](t, again).SigningAddress; got == nil || *got != *verified.SigningAddress {
		t.Fatal("re-verifying changed the signing address")
	}

	// Queued for the chain: register once, then verified → unverified → verified.
	statuses, _ := e.outbox(verified.ID)
	if len(statuses) != 4 {
		t.Fatalf("queued %d items for the organisation, want 4", len(statuses))
	}
}

func TestIssueAndHandoverEndToEnd(t *testing.T) {
	e, _, _ := newEnv(t, false)
	c := e.newChain()
	ctx := context.Background()
	now := time.Now()
	e.passports.Now = func() time.Time { return now }

	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	acme := e.verifiedOrg(tAcme, org.KindProducer, root)
	rival := e.verifiedOrg(tRival, org.KindExporter, root)
	pending := e.register(tStaff, org.KindImporter) // an organisation that never gets verified
	admin, rivalAdmin, pendingAdmin := e.login(tAcme), e.login(tRival), e.login(tStaff)
	acmeSlug, rivalSlug, pendingSlug := e.slug(tAcme), e.slug(tRival), e.slug(tStaff)

	// A facility, and a commodity in the facility's unit.
	rec := post(e, admin, as(acmeSlug, "/facilities"), facilityInput(plotGeo))
	want(t, "create facility", rec, http.StatusCreated)
	plot := decodeTo[facility.Facility](t, rec)

	tplIn := coffeeTemplate("coffee_" + e.suffix)
	tplIn.Name, tplIn.Unit = "Coffee "+e.suffix, plot.CapacityUnit
	rec = post(e, root, "/templates", tplIn)
	want(t, "create template", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)

	issue := func(token, slug string, in passport.IssueInput) *httptest.ResponseRecorder {
		return post(e, token, as(slug, "/passports"), in)
	}
	good := passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: 400_000}

	// ---- Issue: every rule has a refused case.
	want(t, "pending organisation issues", issue(pendingAdmin, pendingSlug, good), http.StatusForbidden)
	want(t, "another organisation's facility", issue(rivalAdmin, rivalSlug, good), http.StatusBadRequest)
	want(t, "zero quantity", issue(admin, acmeSlug, passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID}), http.StatusBadRequest)
	want(t, "unknown template", issue(admin, acmeSlug, passport.IssueInput{FacilityID: plot.ID, TemplateID: fakeOrgID, Quantity: 1}), http.StatusBadRequest)

	wrongUnit := coffeeTemplate("cocoa_" + e.suffix)
	wrongUnit.Name, wrongUnit.Unit = "Cocoa "+e.suffix, "tonne"
	rec = post(e, root, "/templates", wrongUnit)
	want(t, "create second template", rec, http.StatusCreated)
	want(t, "unit mismatch", issue(admin, acmeSlug, passport.IssueInput{
		FacilityID: plot.ID, TemplateID: decodeTo[template.Template](t, rec).ID, Quantity: 1}), http.StatusBadRequest)

	wrongKind := coffeeTemplate("gold_" + e.suffix)
	wrongKind.Name, wrongKind.Unit, wrongKind.OriginFacilityKind = "Gold "+e.suffix, plot.CapacityUnit, facility.KindMineSite
	wrongKind.CapacityRule = template.CapacityRule{Type: template.CapacityFacility}
	rec = post(e, root, "/templates", wrongKind)
	want(t, "create third template", rec, http.StatusCreated)
	want(t, "facility kind mismatch", issue(admin, acmeSlug, passport.IssueInput{
		FacilityID: plot.ID, TemplateID: decodeTo[template.Template](t, rec).ID, Quantity: 1}), http.StatusBadRequest)

	setStatus := func(status string) {
		want(t, "set facility "+status, post(e, root, "/facilities/"+plot.ID+"/status", facility.StatusInput{Status: status}), http.StatusOK)
	}
	setStatus(facility.StatusSuspended)
	want(t, "suspended facility", issue(admin, acmeSlug, good), http.StatusConflict)
	setStatus(facility.StatusActive)

	rec = issue(admin, acmeSlug, good)
	want(t, "issue", rec, http.StatusCreated)
	p := decodeTo[passport.Passport](t, rec)
	wantPeriod := passport.PeriodFor(plot.CapacityPeriod, now)
	if p.OrganisationID != acme.ID || p.Remaining != good.Quantity || p.Status != passport.StatusActive ||
		p.Period == nil || *p.Period != wantPeriod || p.TxHash != nil {
		t.Fatalf("passport recorded wrongly: %+v", p)
	}

	// Capacity: the message says by how much, and the most still allowed.
	left := plot.Capacity - good.Quantity
	over := issue(admin, acmeSlug, passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: left + 5})
	want(t, "over capacity", over, http.StatusConflict)
	wantMsg := passport.CapacityError{Requested: left + 5, Available: left, Unit: plot.CapacityUnit}.Error()
	if got := decodeTo[struct{ Error string }](t, over).Error; got != wantMsg || !strings.Contains(got, fmt.Sprint(left)) {
		t.Fatalf("capacity message = %q, want %q", got, wantMsg)
	}
	want(t, "exactly the rest", issue(admin, acmeSlug, passport.IssueInput{FacilityID: plot.ID, TemplateID: tpl.ID, Quantity: left}), http.StatusCreated)

	// ---- The relayer anchors everything queued so far.
	if n := c.drain(); n == 0 {
		t.Fatal("nothing was queued for the chain")
	}
	opts := &bind.CallOpts{Context: ctx}
	onChain, err := c.contract.Passports.Passport(opts, chain.MustID(p.ID))
	if err != nil {
		t.Fatal(err)
	}
	if onChain.Holder != chain.MustID(acme.ID) || onChain.Quantity.Int64() != good.Quantity ||
		chain.Hex(onChain.DataHash) != p.DataHash || onChain.Commodity != chain.Key(tpl.Commodity) {
		t.Fatalf("on-chain passport differs from the database: %+v", onChain)
	}
	used, _ := c.contract.Facilities.OriginatedInPeriod(opts, chain.MustID(plot.ID), big.NewInt(wantPeriod))
	if used.Int64() != plot.Capacity {
		t.Fatalf("on-chain capacity used = %s, want %d", used, plot.Capacity)
	}
	if verified, _ := c.contract.Orgs.IsVerified(opts, chain.MustID(rival.ID)); !verified {
		t.Fatal("organisation is not verified on-chain")
	}

	one := as(acmeSlug, "/passports/"+p.ID)
	anchored := decodeTo[passport.Passport](t, do(e.h, http.MethodGet, one, admin, ""))
	if anchored.TxHash == nil || anchored.AnchoredAt == nil || len(anchored.Events) != 1 ||
		anchored.Events[0].Kind != passport.EventIssue || anchored.Events[0].TxHash == nil {
		t.Fatalf("transaction hash was not written back: %+v", anchored)
	}

	// ---- Handover: sender's half.
	start := func(token, slug, to string) *httptest.ResponseRecorder {
		return post(e, token, as(slug, "/passports/"+p.ID+"/handovers"), passport.HandoverInput{ToOrganisationID: to, Note: "Loaded at the washing station gate"})
	}
	want(t, "non-holder starts a handover", start(rivalAdmin, rivalSlug, acme.ID), http.StatusNotFound)
	want(t, "handover to itself", start(admin, acmeSlug, acme.ID), http.StatusBadRequest)
	want(t, "handover to an unverified organisation", start(admin, acmeSlug, pending.ID), http.StatusBadRequest)
	want(t, "rival cannot see the passport yet", do(e.h, http.MethodGet, as(rivalSlug, "/passports/"+p.ID), rivalAdmin, ""), http.StatusNotFound)

	rec = start(admin, acmeSlug, rival.ID)
	want(t, "start handover", rec, http.StatusCreated)
	h := decodeTo[passport.Handover](t, rec)
	if h.Status != passport.HandoverAwaiting || !h.SignedByFrom || h.SignedByTo || h.Quantity != good.Quantity {
		t.Fatalf("handover recorded wrongly: %+v", h)
	}
	if strings.Contains(rec.Body.String(), "sig_") {
		t.Fatalf("handover response exposes a signature: %s", rec.Body)
	}
	want(t, "second handover while one is open", start(admin, acmeSlug, rival.ID), http.StatusConflict)

	// Holder unchanged; the receiver can now inspect what it is asked to accept.
	if got := decodeTo[passport.Passport](t, do(e.h, http.MethodGet, one, admin, "")); got.OrganisationID != acme.ID {
		t.Fatal("holder changed before the receiver signed")
	}
	want(t, "receiver inspects the passport", do(e.h, http.MethodGet, as(rivalSlug, "/passports/"+p.ID), rivalAdmin, ""), http.StatusOK)

	hPath := "/handovers/" + h.ID
	want(t, "organisation not party to it", do(e.h, http.MethodGet, as(pendingSlug, hPath), pendingAdmin, ""), http.StatusNotFound)
	want(t, "sender accepts its own handover", post(e, admin, as(acmeSlug, hPath+"/accept"), nil), http.StatusNotFound)
	want(t, "receiver cancels", post(e, rivalAdmin, as(rivalSlug, hPath+"/cancel"), nil), http.StatusNotFound)
	incoming := do(e.h, http.MethodGet, as(rivalSlug, "/handovers?"+passport.ParamStatus+"="+passport.HandoverAwaiting), rivalAdmin, "")
	want(t, "receiver lists handovers", incoming, http.StatusOK)
	if pg := decodeTo[listq.Page[passport.Handover]](t, incoming); pg.Total != 1 || pg.Data[0].ID != h.ID {
		t.Fatalf("receiver's list: %s", incoming.Body)
	}

	// ---- Receiver's half: both signatures, holder changes at once.
	rec = post(e, rivalAdmin, as(rivalSlug, hPath+"/accept"), nil)
	want(t, "accept", rec, http.StatusOK)
	if done := decodeTo[passport.Handover](t, rec); done.Status != passport.HandoverCompleted || !done.SignedByTo || done.TxHash != nil {
		t.Fatalf("accepted handover: %+v", done)
	}
	want(t, "accept twice", post(e, rivalAdmin, as(rivalSlug, hPath+"/accept"), nil), http.StatusConflict)
	held := decodeTo[passport.Passport](t, do(e.h, http.MethodGet, as(rivalSlug, "/passports/"+p.ID), rivalAdmin, ""))
	if held.OrganisationID != rival.ID || held.HandoverCount != 1 || len(held.Events) != 2 {
		t.Fatalf("after acceptance: %+v", held)
	}
	want(t, "former holder still sees its history", do(e.h, http.MethodGet, one, admin, ""), http.StatusOK)
	want(t, "former holder cannot hand it over", start(admin, acmeSlug, rival.ID), http.StatusNotFound)

	// ---- Chain unreachable: the handover waits, then is anchored.
	c.backend.down = true
	if worked, err := c.relayer.ProcessNext(ctx); !worked || err == nil {
		t.Fatalf("with the chain down: worked=%v err=%v", worked, err)
	}
	if statuses, _ := e.outbox(h.ID); len(statuses) != 1 || statuses[0] != chain.StatusPending {
		t.Fatalf("with the chain down the item is %v, want pending", statuses)
	}
	c.backend.down = false
	if n := c.drain(); n != 1 {
		t.Fatalf("after the chain came back the relayer handled %d items, want 1", n)
	}
	onChain, _ = c.contract.Passports.Passport(opts, chain.MustID(p.ID))
	count, _ := c.contract.Custody.Nonces(opts, chain.MustID(p.ID))
	if onChain.Holder != chain.MustID(rival.ID) || count.Int64() != 1 {
		t.Fatalf("on-chain holder %x, handover count %s", onChain.Holder, count)
	}
	final := decodeTo[passport.Handover](t, do(e.h, http.MethodGet, as(rivalSlug, hPath), rivalAdmin, ""))
	if final.TxHash == nil || final.AnchoredAt == nil {
		t.Fatalf("handover not marked anchored: %+v", final)
	}

	// ---- Reject, cancel and expiry leave the holder unchanged.
	back := func() passport.Handover {
		t.Helper()
		rec := post(e, rivalAdmin, as(rivalSlug, "/passports/"+p.ID+"/handovers"), passport.HandoverInput{ToOrganisationID: acme.ID})
		want(t, "start return handover", rec, http.StatusCreated)
		return decodeTo[passport.Handover](t, rec)
	}
	holder := func() string {
		return decodeTo[passport.Passport](t, do(e.h, http.MethodGet, as(rivalSlug, "/passports/"+p.ID), rivalAdmin, "")).OrganisationID
	}

	r1 := "/handovers/" + back().ID
	want(t, "reject without a reason", post(e, admin, as(acmeSlug, r1+"/reject"), passport.RejectInput{}), http.StatusBadRequest)
	rec = post(e, admin, as(acmeSlug, r1+"/reject"), passport.RejectInput{Reason: "Bags were wet"})
	want(t, "reject", rec, http.StatusOK)
	if got := decodeTo[passport.Handover](t, rec); got.Status != passport.HandoverRejected || got.RejectReason == nil || got.SignedByTo {
		t.Fatalf("rejected handover: %+v", got)
	}
	want(t, "accept after rejecting", post(e, admin, as(acmeSlug, r1+"/accept"), nil), http.StatusConflict)

	r2 := "/handovers/" + back().ID
	rec = post(e, rivalAdmin, as(rivalSlug, r2+"/cancel"), nil)
	want(t, "cancel", rec, http.StatusOK)
	if got := decodeTo[passport.Handover](t, rec); got.Status != passport.HandoverCancelled {
		t.Fatalf("cancelled handover: %+v", got)
	}

	r3 := "/handovers/" + back().ID
	// The receiver answers too late: move the deadline into the past.
	e.exec(`UPDATE handovers SET deadline = now() - interval '1 minute' WHERE id = $1::uuid`, strings.TrimPrefix(r3, "/handovers/"))
	want(t, "accept after the deadline", post(e, admin, as(acmeSlug, r3+"/accept"), nil), http.StatusConflict)
	if got := decodeTo[passport.Handover](t, do(e.h, http.MethodGet, as(acmeSlug, r3), admin, "")); got.Status != passport.HandoverExpired {
		t.Fatalf("lapsed handover reads as %s", got.Status)
	}
	expired := do(e.h, http.MethodGet, as(rivalSlug, "/handovers?"+passport.ParamStatus+"="+passport.HandoverExpired), rivalAdmin, "")
	if pg := decodeTo[listq.Page[passport.Handover]](t, expired); pg.Total != 1 {
		t.Fatalf("expired filter returned %d", pg.Total)
	}
	if holder() != rival.ID {
		t.Fatal("holder changed without both signatures")
	}
	// An expired handover no longer blocks a new one.
	back()
	if c.drain() != 0 {
		t.Fatal("rejected, cancelled or expired handovers were queued for the chain")
	}

	// ---- Lists and filters.
	for path, token := range map[string]string{
		as(acmeSlug, "/passports?"+passport.ParamStatus+"=bogus"): admin,
		as(acmeSlug, "/handovers?"+passport.ParamStatus+"=bogus"): admin,
	} {
		want(t, "bad filter "+path, do(e.h, http.MethodGet, path, token, ""), http.StatusBadRequest)
	}
	mine := do(e.h, http.MethodGet, as(acmeSlug, "/passports?"+passport.ParamStatus+"="+passport.StatusActive), admin, "")
	if pg := decodeTo[listq.Page[passport.Passport]](t, mine); pg.Total != 2 {
		t.Fatalf("acme sees %d passports, want 2 (one it holds, one it handed over)", pg.Total)
	}
	if pg := decodeTo[listq.Page[passport.Passport]](t, do(e.h, http.MethodGet, as(pendingSlug, "/passports"), pendingAdmin, "")); pg.Total != 0 {
		t.Fatalf("an unrelated organisation sees %d passports", pg.Total)
	}
}

func TestRelayerMarksRefusedItemsAndMovesOn(t *testing.T) {
	e, _, _ := newEnv(t, false)
	c := e.newChain()
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)

	// An item the contract will refuse: verifying an organisation it has never heard of.
	ghost := e.register(tStaff, org.KindImporter)
	tx, err := e.pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := chain.Enqueue(t.Context(), tx, chain.KindOrgVerified, ghost.ID, chain.OrgVerified{OrgID: ghost.ID, Verified: true}); err != nil {
		t.Fatal(err)
	}
	if err := chain.Enqueue(t.Context(), tx, "no_such_kind", ghost.ID, struct{}{}); err != nil {
		t.Fatal(err)
	}
	tx.Commit(t.Context())

	// A good item queued after it.
	acme := e.verifiedOrg(tAcme, org.KindProducer, root)

	if n := c.drain(); n != 4 {
		t.Fatalf("relayer handled %d items, want 4", n)
	}
	statuses, errs := e.outbox(ghost.ID)
	if len(statuses) != 2 || statuses[0] != chain.StatusFailed || statuses[1] != chain.StatusFailed {
		t.Fatalf("refused items are %v, want both failed", statuses)
	}
	// The contract's own error name is recorded, so a person can see why.
	if errs[0] != "OrgUnknown" {
		t.Fatalf("recorded reason = %q, want the contract error name", errs[0])
	}
	if statuses, _ := e.outbox(acme.ID); len(statuses) != 2 || statuses[0] != chain.StatusConfirmed || statuses[1] != chain.StatusConfirmed {
		t.Fatalf("items after the refused ones are %v, want confirmed", statuses)
	}
	verified, _ := c.contract.Orgs.IsVerified(&bind.CallOpts{}, chain.MustID(acme.ID))
	if !verified {
		t.Fatal("the good item did not reach the chain")
	}

	// A crashed relayer that already sent a transaction finds it instead of sending again.
	rival := e.verifiedOrg(tRival, org.KindExporter, root)
	if worked, err := c.relayer.ProcessNext(t.Context()); !worked || err != nil {
		t.Fatalf("first item: %v %v", worked, err)
	}
	e.exec(`UPDATE chain_outbox SET status = $2 WHERE subject_id = $1::uuid AND kind = $3`, rival.ID, chain.StatusPending, chain.KindOrgRegister)
	if n := c.drain(); n != 2 {
		t.Fatalf("after the simulated crash the relayer handled %d items, want 2", n)
	}
	if statuses, _ := e.outbox(rival.ID); statuses[0] != chain.StatusConfirmed || statuses[1] != chain.StatusConfirmed {
		t.Fatalf("after the simulated crash items are %v", statuses)
	}
}

func TestPeriodFor(t *testing.T) {
	at := func(month time.Month) time.Time { return time.Date(2026, month, 15, 12, 0, 0, 0, time.UTC) }
	cases := []struct {
		period string
		when   time.Time
		want   int64
	}{
		{facility.PeriodYear, at(time.October), 2026},
		{facility.PeriodMonth, at(time.October), 202610},
		{facility.PeriodMonth, at(time.March), 202603},
		{facility.PeriodSeason, at(time.June), 20261},
		{facility.PeriodSeason, at(time.July), 20262},
	}
	for _, tc := range cases {
		if got := passport.PeriodFor(tc.period, tc.when); got != tc.want {
			t.Errorf("%s in %s = %d, want %d", tc.period, tc.when.Month(), got, tc.want)
		}
	}
}

func TestPassportInputValidation(t *testing.T) {
	bad := []passport.IssueInput{
		{TemplateID: fakeOrgID, Quantity: 1},
		{FacilityID: fakeOrgID, Quantity: 1},
		{FacilityID: fakeOrgID, TemplateID: fakeOrgID},
		{FacilityID: fakeOrgID, TemplateID: fakeOrgID, Quantity: -1},
	}
	for _, in := range bad {
		if in.Validate() == "" {
			t.Errorf("%+v accepted", in)
		}
	}
	if msg := (passport.IssueInput{FacilityID: fakeOrgID, TemplateID: fakeOrgID, Quantity: 1}).Validate(); msg != "" {
		t.Errorf("valid input rejected: %s", msg)
	}

	h := passport.HandoverInput{ToOrganisationID: fakeOrgID, Note: "  at the gate  "}
	if msg := h.Validate(); msg != "" || h.Note != "at the gate" {
		t.Errorf("handover input: %q %q", msg, h.Note)
	}
	for _, in := range []passport.HandoverInput{{}, {ToOrganisationID: fakeOrgID, Note: strings.Repeat("x", passport.MaxNoteLen+1)}} {
		if in.Validate() == "" {
			t.Errorf("%+v accepted", in)
		}
	}
	for _, in := range []passport.RejectInput{{}, {Reason: " "}, {Reason: strings.Repeat("x", passport.MaxNoteLen+1)}} {
		if in.Validate() == "" {
			t.Errorf("reject %+v accepted", in)
		}
	}
}
