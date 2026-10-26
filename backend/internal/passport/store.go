package passport

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrFacilityNotFound  = errors.New("facility not found")
	ErrFacilitySuspended = errors.New("facility suspended")
	ErrFacilityPending   = errors.New("facility pending approval")
	ErrLicenceExpired    = errors.New("facility licence expired")
	ErrTemplateNotFound  = errors.New("template not found")
	ErrKindMismatch      = errors.New("facility kind mismatch")
	ErrUnitMismatch      = errors.New("unit mismatch")
	ErrNotHolder         = errors.New("not holder")
	ErrNotActive         = errors.New("passport not active")
	ErrOpenHandover      = errors.New("open handover exists")
	ErrReceiverInvalid   = errors.New("receiver invalid")
	ErrHandoverClosed    = errors.New("handover closed")
	ErrHandoverStale     = errors.New("handover stale")
)

// EvidenceError lists the evidence the template asks for that has no valid
// certificate on the origin facility yet.
type EvidenceError struct {
	Missing []string
}

func (e EvidenceError) Error() string {
	return fmt.Sprintf("this product needs evidence that is missing: %s. Attach a valid certificate of each first",
		strings.Join(e.Missing, ", "))
}

// CapacityError says by how much an issue would exceed the facility's
// capacity for the period, and the most that can still be issued.
type CapacityError struct {
	Requested, Available int64
	Unit                 string
}

func (e CapacityError) Error() string {
	return fmt.Sprintf("capacity exceeded by %d %s; at most %d %s can still be issued from this facility in this period",
		e.Requested-e.Available, e.Unit, e.Available, e.Unit)
}

type Passport struct {
	ID string `json:"id"`
	// Code is the readable lot code (ACME-2026-0412); it works wherever the id does.
	Code           string `json:"code"`
	OrganisationID string `json:"organisation_id"` // current holder
	// FacilityID and Period are set on a passport issued at an origin
	// facility; TransformID on one made by a transform. Never both.
	FacilityID    *string    `json:"facility_id"`
	TransformID   *string    `json:"transform_id"`
	TemplateID    string     `json:"template_id"`
	Commodity     string     `json:"commodity"`
	Unit          string     `json:"unit"`
	Quantity      int64      `json:"quantity"`
	Remaining     int64      `json:"remaining"`
	Period        *int64     `json:"period"`
	Status        string     `json:"status"`
	DataHash      string     `json:"data_hash"`
	HandoverCount int64      `json:"handover_count"`
	TxHash        *string    `json:"tx_hash"`
	AnchoredAt    *time.Time `json:"anchored_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// Events is the journey so far; filled when one passport is fetched.
	Events []Event `json:"events,omitempty"`
}

type Event struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	FromOrg   *string   `json:"from_org"`
	ToOrg     *string   `json:"to_org"`
	Quantity  int64     `json:"quantity"`
	DataHash  string    `json:"data_hash"`
	TxHash    *string   `json:"tx_hash"`
	CreatedAt time.Time `json:"created_at"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func location(lat, lon *float64) *Location {
	if lat == nil || lon == nil {
		return nil
	}
	return &Location{*lat, *lon}
}

// What each list endpoint lets callers search and sort by.
var (
	List = listq.Config{
		Alias:  "p",
		Search: []string{"commodity", "status"},
		Sort:   []string{"commodity", "quantity", "status", listq.ColCreatedAt},
	}
	HandoverList = listq.Config{
		Alias:  "h",
		Search: []string{"note", "status"},
		Sort:   []string{"status", "deadline", listq.ColCreatedAt},
	}
)

const (
	columns = `p.id::text, p.code, p.organisation_id::text, p.facility_id::text, p.transform_id::text, p.template_id::text, p.commodity,
	           p.unit, p.quantity, p.remaining, p.period, p.status, p.data_hash, p.handover_count,
	           p.tx_hash, p.anchored_at, p.created_at, p.updated_at`

	// An awaiting handover past its deadline reads as expired ($%[1]d = awaiting, $%[2]d = expired).
	handoverColumnsFmt = `h.id::text, h.passport_id::text, h.from_org::text, h.to_org::text, h.stage, h.quantity, h.note,
	           h.data_hash, h.deadline,
	           CASE WHEN h.status = $%[1]d AND h.deadline < now() THEN $%[2]d ELSE h.status END,
	           true, h.sig_to IS NOT NULL, h.reject_reason, h.reject_reasons, h.responded_at, h.tx_hash, h.anchored_at,
	           h.created_at, h.updated_at, h.split_from::text, h.lat, h.lon, h.received_quantity,
	           h.receiver_lat, h.receiver_lon, h.template_version, h.seal_number, h.checklist`

	// visibleFmt: who may see a passport. Its holder, anyone it passed through,
	// and the receiver of a handover still awaiting an answer (they must be
	// able to inspect what they are asked to accept). $%[1]d is the gate.
	visibleFmt = `($%[1]d = '' OR p.organisation_id = NULLIF($%[1]d, '')::uuid
	  OR EXISTS (SELECT 1 FROM trace_events e WHERE e.passport_id = p.id AND NOT e.is_deleted
	             AND (e.from_org = NULLIF($%[1]d, '')::uuid OR e.to_org = NULLIF($%[1]d, '')::uuid))
	  OR EXISTS (SELECT 1 FROM handovers h WHERE h.passport_id = p.id AND NOT h.is_deleted
	             AND h.to_org = NULLIF($%[1]d, '')::uuid
	             AND h.status = '` + HandoverAwaiting + `' AND h.deadline >= now()))`

	// partyFmt: who may see a handover: its sender and its receiver.
	partyFmt = `($%[1]d = '' OR h.from_org = NULLIF($%[1]d, '')::uuid OR h.to_org = NULLIF($%[1]d, '')::uuid)`
)

func scan(row pgx.CollectableRow) (Passport, error) {
	var p Passport
	err := row.Scan(&p.ID, &p.Code, &p.OrganisationID, &p.FacilityID, &p.TransformID, &p.TemplateID, &p.Commodity, &p.Unit,
		&p.Quantity, &p.Remaining, &p.Period, &p.Status, &p.DataHash, &p.HandoverCount,
		&p.TxHash, &p.AnchoredAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func scanEvent(row pgx.CollectableRow) (Event, error) {
	var e Event
	err := row.Scan(&e.ID, &e.Kind, &e.FromOrg, &e.ToOrg, &e.Quantity, &e.DataHash, &e.TxHash, &e.CreatedAt)
	return e, err
}

func one[T any](rows pgx.Rows, err error, scan pgx.RowToFunc[T]) (T, error) {
	var zero T
	if err != nil {
		return zero, err
	}
	v, err := pgx.CollectExactlyOneRow(rows, scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return zero, ErrNotFound
	}
	return v, err
}

func onePassport(rows pgx.Rows, err error) (Passport, error) { return one(rows, err, scan) }

func newID(ctx context.Context, tx pgx.Tx) (id string, err error) {
	err = tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&id)
	return id, err
}

type Store struct {
	Pool *pgxpool.Pool
	// Secrets opens organisations' signing keys.
	Secrets *secret.Box
	// Domain binds handover signatures to the deployed CustodyManager.
	Domain chain.Domain
}

// issueRecord is what a passport's data hash covers.
type issueRecord struct {
	ID             string `json:"id"`
	OrganisationID string `json:"organisation_id"`
	FacilityID     string `json:"facility_id"`
	TemplateID     string `json:"template_id"`
	Commodity      string `json:"commodity"`
	Unit           string `json:"unit"`
	Quantity       int64  `json:"quantity"`
	Period         int64  `json:"period"`
}

// Issue creates an origin passport after checking every rule the contract
// will check again: the facility is the organisation's and active, it suits
// the commodity, and its capacity for the period is not exceeded.
func (s Store) Issue(ctx context.Context, orgID, userID string, in IssueInput, now time.Time) (Passport, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Passport{}, err
	}
	defer tx.Rollback(ctx)

	f, err := lockOrigin(ctx, tx, orgID, in.FacilityID, now)
	if err != nil {
		return Passport{}, err
	}
	commodity, unit, err := originTemplate(ctx, tx, in, f)
	if err != nil {
		return Passport{}, err
	}
	period := facility.PeriodFor(f.period, now)
	used, err := facility.CapacityUsed(ctx, tx, []string{in.FacilityID}, []int64{period})
	if err != nil {
		return Passport{}, err
	}
	if available := f.capacity - used[in.FacilityID]; in.Quantity > available {
		return Passport{}, CapacityError{Requested: in.Quantity, Available: available, Unit: unit}
	}
	rec := issueRecord{OrganisationID: orgID, FacilityID: in.FacilityID, TemplateID: in.TemplateID,
		Commodity: commodity, Unit: unit, Quantity: in.Quantity, Period: period}
	p, err := insertIssued(ctx, tx, rec, userID, now)
	if err != nil {
		return Passport{}, err
	}
	return p, tx.Commit(ctx)
}

type originFacility struct {
	kind, unit, period string
	capacity           int64
}

// lockOrigin locks the facility, so concurrent issues queue up and two of them
// cannot both fit into the last of the capacity. It must be active and licensed
// (a licence is valid through its last day).
func lockOrigin(ctx context.Context, tx pgx.Tx, orgID, facilityID string, now time.Time) (originFacility, error) {
	var f originFacility
	var status string
	var licenceValidTo *time.Time
	err := tx.QueryRow(ctx,
		`SELECT kind, status, capacity, capacity_unit, capacity_period, licence_valid_to FROM facilities
		 WHERE id = $1::uuid AND organisation_id = $2::uuid AND NOT is_deleted FOR UPDATE`,
		facilityID, orgID).Scan(&f.kind, &status, &f.capacity, &f.unit, &f.period, &licenceValidTo)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return f, ErrFacilityNotFound
	case err != nil:
		return f, err
	case status == facility.StatusPending:
		return f, ErrFacilityPending
	case status != facility.StatusActive:
		return f, ErrFacilitySuspended
	case licenceValidTo != nil && now.UTC().Format(facility.DateLayout) > licenceValidTo.Format(facility.DateLayout):
		return f, ErrLicenceExpired
	}
	return f, nil
}

// originTemplate loads the template (the product), checks it fits the
// facility, and checks the facility holds the evidence the template asks for.
func originTemplate(ctx context.Context, tx pgx.Tx, in IssueInput, f originFacility) (commodity, unit string, err error) {
	var (
		originKind string
		evidence   []string
	)
	err = tx.QueryRow(ctx,
		`SELECT commodity, unit, origin_facility_kind, required_evidence FROM commodity_templates
		 WHERE id = $1::uuid AND NOT is_deleted`, in.TemplateID).Scan(&commodity, &unit, &originKind, &evidence)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", "", ErrTemplateNotFound
	case err != nil:
		return "", "", err
	case originKind != f.kind:
		return "", "", ErrKindMismatch
	case unit != f.unit:
		return "", "", ErrUnitMismatch
	}
	if err := requireEvidence(ctx, tx, evidence, in.FacilityID); err != nil {
		return "", "", err
	}
	return commodity, unit, nil
}

// insertIssued writes the passport, its ISSUE trace event and its chain item.
func insertIssued(ctx context.Context, tx pgx.Tx, rec issueRecord, userID string, now time.Time) (Passport, error) {
	var err error
	if rec.ID, err = newID(ctx, tx); err != nil {
		return Passport{}, err
	}
	hash, err := chain.DataHash(rec)
	if err != nil {
		return Passport{}, err
	}
	dataHash := chain.Hex(hash)
	code, err := nextCode(ctx, tx, rec.OrganisationID, now)
	if err != nil {
		return Passport{}, err
	}
	p, err := onePassport(tx.Query(ctx,
		`INSERT INTO passports AS p (id, organisation_id, facility_id, template_id, commodity, unit,
		                             quantity, remaining, period, status, data_hash, issued_by, code)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $7, $8, $9, $10, $11::uuid, $12)
		 RETURNING `+columns,
		rec.ID, rec.OrganisationID, rec.FacilityID, rec.TemplateID, rec.Commodity, rec.Unit, rec.Quantity, rec.Period,
		StatusActive, dataHash, userID, code))
	if err != nil {
		return Passport{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, to_org, quantity, data_hash)
		 VALUES ($1::uuid, $1::uuid, $2, $3::uuid, $4, $5)`, rec.ID, EventIssue, rec.OrganisationID, rec.Quantity, dataHash); err != nil {
		return Passport{}, err
	}
	return p, chain.Enqueue(ctx, tx, chain.KindPassportIssue, rec.ID, chain.PassportIssue{
		PassportID: rec.ID, Commodity: rec.Commodity, FacilityID: rec.FacilityID,
		Period: rec.Period, Quantity: rec.Quantity, DataHash: dataHash,
	})
}

func visible(n int) string { return fmt.Sprintf(visibleFmt, n) }

func (s Store) Find(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Passport], error) {
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM passports p",
		[]string{visible(1), `($2 = '' OR p.status = $2)`}, []any{gate, status}, scan)
}

// Get returns one passport with its journey.
func (s Store) Get(ctx context.Context, id, gate string) (Passport, error) {
	p, err := onePassport(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM passports p WHERE p.id = $1::uuid AND NOT p.is_deleted AND `+visible(2), id, gate))
	if err != nil {
		return Passport{}, err
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT id::text, kind, from_org::text, to_org::text, quantity, data_hash, tx_hash, created_at
		 FROM trace_events WHERE passport_id = $1::uuid AND NOT is_deleted ORDER BY created_at, id`, id)
	if err != nil {
		return Passport{}, err
	}
	p.Events, err = pgx.CollectRows(rows, scanEvent)
	return p, err
}
