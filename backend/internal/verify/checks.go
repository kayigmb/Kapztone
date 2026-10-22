package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
)

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
		return v.holderDiffers(p.ID)
	}
	return nil
}

// holderDiffers: the holder changes here when the receiver accepts, and
// on-chain a moment later. Any other difference is a record that was altered.
func (v *run) holderDiffers(passportID string) error {
	var pending bool
	err := v.h.Pool.QueryRow(v.ctx,
		`SELECT EXISTS (SELECT 1 FROM handovers WHERE passport_id = $1::uuid AND status = $2
		                AND tx_hash IS NULL AND NOT is_deleted)`, passportID, passport.HandoverCompleted).Scan(&pending)
	if err != nil {
		return err
	}
	if pending {
		v.add(Warning, CheckAnchored, CustodyPending, passportID)
	} else {
		v.add(Failed, CheckAnchored, ChainMismatch, passportID)
	}
	return nil
}

// origins checks that every branch of the tree ends at a facility, and each
// such facility: active, owned by a verified organisation, unchanged, and
// within its capacity. It returns the facilities' ids.
func (v *run) origins(nodes []passport.Passport) ([]string, error) {
	facilityIDs, lots := v.originLots(nodes)
	capacity := map[string]int64{}
	for _, id := range facilityIDs {
		c, err := v.origin(id)
		if err != nil {
			return nil, err
		}
		capacity[id] = c
	}
	for l := range lots {
		if err := v.lotCapacity(l, capacity[l.facility]); err != nil {
			return nil, err
		}
	}
	return facilityIDs, nil
}

type lot struct {
	facility string
	period   int64
}

// originLots lists each origin facility once and every (facility, period) lot
// issued from it. A passport with no facility must have been made from others.
func (v *run) originLots(nodes []passport.Passport) ([]string, map[lot]bool) {
	made := map[string]bool{}
	for _, e := range v.res.Genealogy.Edges {
		made[e.ChildID] = true
	}
	var facilityIDs []string
	lots := map[lot]bool{}
	for _, p := range nodes {
		switch {
		case p.FacilityID != nil && p.Period != nil:
			if !slices.Contains(facilityIDs, *p.FacilityID) {
				facilityIDs = append(facilityIDs, *p.FacilityID)
			}
			lots[lot{*p.FacilityID, *p.Period}] = true
		case !made[p.ID]:
			v.add(Failed, CheckOrigin, OriginMissing, p.ID)
		}
	}
	return facilityIDs, lots
}

// origin adds one origin facility to the result, checks its status and its
// on-chain record, and returns its capacity.
func (v *run) origin(id string) (int64, error) {
	var (
		o                 Origin
		location          string
		status, orgStatus string
		dataHash          *string
		capacity          int64
	)
	err := v.h.Pool.QueryRow(v.ctx,
		`SELECT f.id::text, f.name, f.kind, ST_AsGeoJSON(f.geom), f.status, f.capacity, f.data_hash,
		        o.id::text, o.name, o.status
		 FROM facilities f JOIN organisations o ON o.id = f.organisation_id WHERE f.id = $1::uuid`, id,
	).Scan(&o.Facility.ID, &o.Facility.Name, &o.Facility.Kind, &location, &status, &capacity, &dataHash,
		&o.Organisation.ID, &o.Organisation.Name, &orgStatus)
	if err != nil {
		return 0, err
	}
	o.Facility.Location = json.RawMessage(location)
	v.res.Origins = append(v.res.Origins, o)
	v.res.Route = append(v.res.Route, RoutePoint{FacilityID: o.Facility.ID, Name: o.Facility.Name,
		Kind: o.Facility.Kind, Location: o.Facility.Location, Seq: len(v.res.Route)})

	if status != facility.StatusActive {
		v.add(Warning, CheckOrigin, FacilitySuspended, id)
	}
	if orgStatus != auth.OrgVerified {
		v.add(Warning, CheckOrigin, OwnerNotVerified, id)
	}
	v.facilityOnChain(id, dataHash, capacity)
	return capacity, nil
}

func (v *run) facilityOnChain(id string, dataHash *string, capacity int64) {
	if v.chainDown {
		return
	}
	c, err := v.h.Contracts.Facilities.Facility(v.opts, chain.MustID(id))
	if !v.onChain(err) {
		v.absent(CheckOrigin, id, false)
		return
	}
	if dataHash == nil || chain.Hex(c.DataHash) != *dataHash || c.CapacityPerPeriod.Cmp(big.NewInt(capacity)) != 0 {
		v.add(Failed, CheckOrigin, FacilityAltered, id)
	}
}

func (v *run) lotCapacity(l lot, capacity int64) error {
	var issued int64
	err := v.h.Pool.QueryRow(v.ctx,
		`SELECT COALESCE(sum(quantity), 0) FROM passports WHERE facility_id = $1::uuid AND period = $2 AND NOT is_deleted`,
		l.facility, l.period).Scan(&issued)
	if err != nil {
		return err
	}
	over := issued > capacity
	if !over && !v.chainDown {
		// The contract refuses an issue over capacity, so this holds unless the chain itself is wrong.
		used, err := v.h.Contracts.Facilities.OriginatedInPeriod(v.opts, chain.MustID(l.facility), big.NewInt(l.period))
		over = v.onChain(err) && used.Cmp(big.NewInt(capacity)) > 0
	}
	if over {
		v.add(Failed, CheckCapacity, CapacityExceeded, l.facility)
	}
	return nil
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
		v.certificateOnChain(a)
	}
	return nil
}

func (v *run) certificateOnChain(a attestation.Attestation) {
	if v.chainDown {
		return
	}
	c, err := v.h.Contracts.Attestations.Attestation(v.opts, chain.MustID(a.ID))
	if !v.onChain(err) {
		v.absent(CheckCertificates, a.ID, a.TxHash != nil)
		return
	}
	if chain.Hex(c.DocHash) != a.DocHash || c.Subject != chain.MustID(a.SubjectID) || c.CertType != chain.Key(a.Type) {
		v.add(Failed, CheckCertificates, CertificateAltered, a.ID)
	}
	accredited, err := v.h.Contracts.Orgs.IsAccredited(v.opts, chain.MustID(a.IssuerOrgID), chain.Key(a.Type))
	if v.onChain(err) && !accredited {
		v.add(Warning, CheckCertificates, IssuerNotAccredited, a.ID)
	}
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
