package verify

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// Reason is one thing a check found.
type Reason struct {
	Level     string `json:"level"`
	Check     string `json:"check"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	SubjectID string `json:"subject_id"` // the passport, facility or certificate concerned
}

// Check is one line of "what was checked".
type Check struct {
	Name   string `json:"name"`
	Result string `json:"result"` // verified, warning or failed
}

// Named is the public face of an organisation or facility: no licence
// numbers, no people.
type Named struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Origin struct {
	Facility struct {
		Named
		Kind     string          `json:"kind"`
		Location json.RawMessage `json:"location"` // GeoJSON
	} `json:"facility"`
	Organisation Named `json:"organisation"`
}

// Result is everything the public verification page shows.
type Result struct {
	Verdict  string            `json:"verdict"`
	Reasons  []Reason          `json:"reasons"`
	Checks   []Check           `json:"checks"`
	Passport passport.Passport `json:"passport"`
	Origins  []Origin          `json:"origins"`
	// Organisations names every organisation the answer refers to by id.
	Organisations []Named                   `json:"organisations"`
	Genealogy     passport.Genealogy        `json:"genealogy"`
	Certificates  []attestation.Attestation `json:"certificates"`
}

type Handlers struct {
	Pool         *pgxpool.Pool
	Passports    passport.Store
	Certificates attestation.Store
	// Contracts are read, never written, here.
	Contracts chain.Contracts
	// Files holds certificate files, re-hashed on every check.
	Files *upload.Bucket
	// Cache holds each answer briefly; nil means every request is worked out afresh.
	Cache *cache.Cache
}

// Verify is public: no sign-in. It takes the passport's id or its lot code.
// Each answer is cached, so a revocation can take up to the cache's lifetime to show.
func (h *Handlers) Verify(w http.ResponseWriter, r *http.Request) {
	id, err := h.Passports.Resolve(r.Context(), chi.URLParam(r, httpx.ParamID))
	if errors.Is(err, passport.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if err != nil {
		httpx.ServerError(w, "find passport", err)
		return
	}
	key := CacheKeyPrefix + id
	var res Result
	if h.Cache != nil && h.Cache.Get(r.Context(), key, &res) {
		httpx.WriteJSON(w, http.StatusOK, res)
		return
	}
	res, err = h.Run(r.Context(), id)
	switch {
	case errors.Is(err, passport.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	case err != nil:
		httpx.ServerError(w, "verify", err)
		return
	}
	if h.Cache != nil {
		h.Cache.Set(r.Context(), key, res)
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

// run carries the state of one verification.
type run struct {
	h    *Handlers
	ctx  context.Context
	opts *bind.CallOpts
	res  *Result
	// chainDown is set the first time the chain cannot be reached; no
	// further chain reads are tried.
	chainDown bool
}

func (v *run) add(level, check, code, subjectID string) {
	v.res.Reasons = append(v.res.Reasons, Reason{Level: level, Check: check, Code: code, Message: messages[code], SubjectID: subjectID})
}

// onChain interprets the error of a chain read. found is false when the
// chain does not hold the thing, or could not be asked.
func (v *run) onChain(err error) (found bool) {
	switch {
	case err == nil:
		return true
	case chain.Refused(err):
		return false
	}
	if !v.chainDown {
		v.chainDown = true
		v.add(Warning, CheckAnchored, ChainUnreachable, v.res.Passport.ID)
	}
	return false
}

// absent reports a record the chain does not hold: a warning while it waits
// for the relayer, a failure if the database claims it was anchored.
func (v *run) absent(check, subjectID string, anchored bool) {
	if v.chainDown {
		return
	}
	if anchored {
		v.add(Failed, check, ChainMissing, subjectID)
	} else {
		v.add(Warning, check, NotAnchored, subjectID)
	}
}

// Run verifies one passport back to its origins (PLAN.md §4.6).
//
// ponytail: chain reads happen one after another. If a lot with many
// ancestors is too slow on a real network, run them concurrently.
func (h *Handlers) Run(ctx context.Context, id string) (Result, error) {
	res := Result{Reasons: []Reason{}, Origins: []Origin{}}
	v := &run{h: h, ctx: ctx, opts: &bind.CallOpts{Context: ctx}, res: &res}

	var err error
	if res.Passport, err = h.Passports.Get(ctx, id, ""); err != nil {
		return res, err
	}
	if res.Genealogy, err = h.Passports.Genealogy(ctx, id, "", passport.DirectionAncestors); err != nil {
		return res, err
	}
	nodes := res.Genealogy.Nodes
	ids := make([]string, len(nodes))
	for i, p := range nodes {
		ids[i] = p.ID
	}

	for _, p := range nodes {
		if err := v.passport(p); err != nil {
			return res, err
		}
	}
	facilityIDs, err := v.origins(nodes)
	if err != nil {
		return res, err
	}
	if err := v.massBalance(nodes); err != nil {
		return res, err
	}
	if err := v.certificates(append(ids, facilityIDs...)); err != nil {
		return res, err
	}
	if err := v.flags(ids); err != nil {
		return res, err
	}

	if err := v.names(); err != nil {
		return res, err
	}

	// The verdict is the worst level found; each check reports its own worst.
	worst := map[string]string{}
	rank := map[string]int{Verified: 0, Warning: 1, Failed: 2}
	res.Verdict = Verified
	for _, r := range res.Reasons {
		if rank[r.Level] > rank[worst[r.Check]] {
			worst[r.Check] = r.Level
		}
		if rank[r.Level] > rank[res.Verdict] {
			res.Verdict = r.Level
		}
	}
	for _, name := range Checks {
		result := Verified
		if w, ok := worst[name]; ok {
			result = w
		}
		res.Checks = append(res.Checks, Check{Name: name, Result: result})
	}
	return res, nil
}

// names looks up the name of every organisation the answer mentions.
func (v *run) names() error {
	res := v.res
	ids := []string{}
	add := func(id *string) {
		if id != nil {
			ids = append(ids, *id)
		}
	}
	for _, p := range res.Genealogy.Nodes {
		add(&p.OrganisationID)
	}
	for _, e := range res.Passport.Events {
		add(e.FromOrg)
		add(e.ToOrg)
	}
	for _, a := range res.Certificates {
		add(&a.IssuerOrgID)
	}
	for _, o := range res.Origins {
		add(&o.Organisation.ID)
	}
	rows, err := v.h.Pool.Query(v.ctx,
		`SELECT id::text, name FROM organisations WHERE id = ANY($1::uuid[]) ORDER BY name`, ids)
	if err != nil {
		return err
	}
	res.Organisations, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (Named, error) {
		var n Named
		err := row.Scan(&n.ID, &n.Name)
		return n, err
	})
	return err
}

// passport checks one passport of the tree: its record, its copy on-chain, its status.
func (v *run) passport(p passport.Passport) error {
	if p.Status == passport.StatusRevoked {
		v.add(Failed, CheckStatus, PassportRevoked, p.ID)
	}
	want, err := v.h.Passports.RecordHash(v.ctx, p)
	if err != nil {
		return err
	}
	if want != p.DataHash {
		v.add(Failed, CheckRecord, RecordAltered, p.ID)
	}

	if v.chainDown {
		return nil
	}
	c, err := v.h.Contracts.Passports.Passport(v.opts, chain.MustID(p.ID))
	if !v.onChain(err) {
		v.absent(CheckAnchored, p.ID, p.TxHash != nil)
		return nil
	}
	if chain.Hex(c.DataHash) != p.DataHash || c.Quantity.Cmp(big.NewInt(p.Quantity)) != 0 {
		v.add(Failed, CheckAnchored, ChainMismatch, p.ID)
	}
	if c.Holder != chain.MustID(p.OrganisationID) {
		// The holder changes here when the receiver accepts, and on-chain a
		// moment later. Any other difference is a record that was altered.
		var pending bool
		err := v.h.Pool.QueryRow(v.ctx,
			`SELECT EXISTS (SELECT 1 FROM handovers WHERE passport_id = $1::uuid AND status = $2
			                AND tx_hash IS NULL AND NOT is_deleted)`, p.ID, passport.HandoverCompleted).Scan(&pending)
		if err != nil {
			return err
		}
		if pending {
			v.add(Warning, CheckAnchored, CustodyPending, p.ID)
		} else {
			v.add(Failed, CheckAnchored, ChainMismatch, p.ID)
		}
	}
	return nil
}

// origins checks that every branch of the tree ends at a facility, and each
// such facility: active, owned by a verified organisation, unchanged, and
// within its capacity. It returns the facilities' ids.
func (v *run) origins(nodes []passport.Passport) ([]string, error) {
	made := map[string]bool{} // passports the tree shows as made from others
	for _, e := range v.res.Genealogy.Edges {
		made[e.ChildID] = true
	}
	type lot struct {
		facility string
		period   int64
	}
	var (
		facilityIDs []string
		seen        = map[string]bool{}
		lots        = map[lot]bool{}
	)
	for _, p := range nodes {
		switch {
		case p.FacilityID != nil && p.Period != nil:
			lots[lot{*p.FacilityID, *p.Period}] = true
			if !seen[*p.FacilityID] {
				seen[*p.FacilityID] = true
				facilityIDs = append(facilityIDs, *p.FacilityID)
			}
		case !made[p.ID]:
			v.add(Failed, CheckOrigin, OriginMissing, p.ID)
		}
	}

	capacity := map[string]int64{}
	for _, id := range facilityIDs {
		var (
			o                 Origin
			location          string
			status, orgStatus string
			dataHash          *string
			facilityCapacity  int64
		)
		err := v.h.Pool.QueryRow(v.ctx,
			`SELECT f.id::text, f.name, f.kind, ST_AsGeoJSON(f.geom), f.status, f.capacity, f.data_hash,
			        o.id::text, o.name, o.status
			 FROM facilities f JOIN organisations o ON o.id = f.organisation_id WHERE f.id = $1::uuid`, id,
		).Scan(&o.Facility.ID, &o.Facility.Name, &o.Facility.Kind, &location, &status, &facilityCapacity, &dataHash,
			&o.Organisation.ID, &o.Organisation.Name, &orgStatus)
		if err != nil {
			return nil, err
		}
		o.Facility.Location = json.RawMessage(location)
		v.res.Origins = append(v.res.Origins, o)
		capacity[id] = facilityCapacity

		if status != facility.StatusActive {
			v.add(Warning, CheckOrigin, FacilitySuspended, id)
		}
		if orgStatus != auth.OrgVerified {
			v.add(Warning, CheckOrigin, OwnerNotVerified, id)
		}
		if v.chainDown {
			continue
		}
		c, err := v.h.Contracts.Facilities.Facility(v.opts, chain.MustID(id))
		if !v.onChain(err) {
			v.absent(CheckOrigin, id, false)
			continue
		}
		if dataHash == nil || chain.Hex(c.DataHash) != *dataHash || c.CapacityPerPeriod.Cmp(big.NewInt(facilityCapacity)) != 0 {
			v.add(Failed, CheckOrigin, FacilityAltered, id)
		}
	}

	for l := range lots {
		var issued int64
		err := v.h.Pool.QueryRow(v.ctx,
			`SELECT COALESCE(sum(quantity), 0) FROM passports WHERE facility_id = $1::uuid AND period = $2 AND NOT is_deleted`,
			l.facility, l.period).Scan(&issued)
		if err != nil {
			return nil, err
		}
		over := issued > capacity[l.facility]
		if !over && !v.chainDown {
			// The contract refuses an issue over capacity, so this holds unless the chain itself is wrong.
			used, err := v.h.Contracts.Facilities.OriginatedInPeriod(v.opts, chain.MustID(l.facility), big.NewInt(l.period))
			over = v.onChain(err) && used.Cmp(big.NewInt(capacity[l.facility])) > 0
		}
		if over {
			v.add(Failed, CheckCapacity, CapacityExceeded, l.facility)
		}
	}
	return facilityIDs, nil
}

// massBalance rechecks every transform in the tree against the yield it
// recorded: outputs × scale ≤ inputs × yield, compared without dividing.
func (v *run) massBalance(nodes []passport.Passport) error {
	var transforms []string
	seen := map[string]bool{}
	for _, p := range nodes {
		if p.TransformID != nil && !seen[*p.TransformID] {
			seen[*p.TransformID] = true
			transforms = append(transforms, *p.TransformID)
		}
	}
	rows, err := v.h.Pool.Query(v.ctx,
		`SELECT t.id::text, t.max_yield_ppm,
		        (SELECT COALESCE(sum(quantity_used), 0) FROM transform_inputs WHERE transform_id = t.id AND NOT is_deleted),
		        (SELECT COALESCE(sum(quantity), 0) FROM passports WHERE transform_id = t.id AND NOT is_deleted)
		 FROM transforms t WHERE t.id = ANY($1::uuid[])`, transforms)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id           string
			ppm, in, out int64
		)
		if err := rows.Scan(&id, &ppm, &in, &out); err != nil {
			return err
		}
		allowed := new(big.Int).Mul(big.NewInt(in), big.NewInt(ppm))
		if new(big.Int).Mul(big.NewInt(out), big.NewInt(chain.YieldScale)).Cmp(allowed) > 0 {
			v.add(Failed, CheckMassBalance, MassBalanceBroken, id)
		}
	}
	return rows.Err()
}

// fileHash is the SHA-256 of a stored certificate file, as doc_hash writes it.
func (v *run) fileHash(storageKey string) (string, error) {
	f, err := v.h.Files.Open(v.ctx, storageKey)
	if err != nil {
		return "", err
	}
	defer f.Close()
	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(sum.Sum(nil)), nil
}

// certificates checks every certificate about the tree's passports and facilities.
func (v *run) certificates(subjectIDs []string) error {
	certs, err := v.h.Certificates.ForSubjects(v.ctx, subjectIDs)
	if err != nil {
		return err
	}
	v.res.Certificates = certs
	if certs == nil {
		v.res.Certificates = []attestation.Attestation{}
	}
	for _, a := range certs {
		switch a.Status {
		case attestation.StatusExpired:
			v.add(Warning, CheckCertificates, CertificateExpired, a.ID)
		case attestation.StatusRevoked:
			v.add(Warning, CheckCertificates, CertificateRevoked, a.ID)
		}
		// The file itself: altered on disk, it no longer matches its fingerprint.
		switch got, err := v.fileHash(a.StorageKey); {
		case err != nil:
			v.add(Warning, CheckCertificates, CertificateFileGone, a.ID)
		case got != a.DocHash:
			v.add(Failed, CheckCertificates, CertificateAltered, a.ID)
		}

		if v.chainDown {
			continue
		}
		c, err := v.h.Contracts.Attestations.Attestation(v.opts, chain.MustID(a.ID))
		if !v.onChain(err) {
			v.absent(CheckCertificates, a.ID, a.TxHash != nil)
			continue
		}
		if chain.Hex(c.DocHash) != a.DocHash || c.Subject != chain.MustID(a.SubjectID) || c.CertType != chain.Key(a.Type) {
			v.add(Failed, CheckCertificates, CertificateAltered, a.ID)
		}
		accredited, err := v.h.Contracts.Orgs.IsAccredited(v.opts, chain.MustID(a.IssuerOrgID), chain.Key(a.Type))
		if v.onChain(err) && !accredited {
			v.add(Warning, CheckCertificates, IssuerNotAccredited, a.ID)
		}
	}
	return nil
}

// flags reports open concerns anywhere in the tree. What was said stays private.
func (v *run) flags(passportIDs []string) error {
	rows, err := v.h.Pool.Query(v.ctx,
		`SELECT DISTINCT passport_id::text FROM flags WHERE passport_id = ANY($1::uuid[]) AND status = $2 AND NOT is_deleted`,
		passportIDs, flag.StatusOpen)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		v.add(Warning, CheckStatus, FlagOpen, id)
	}
	return rows.Err()
}
