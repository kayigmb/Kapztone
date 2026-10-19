package facility

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrBadGeometry = errors.New("invalid geometry")
	ErrNotPending  = errors.New("not pending approval")
	ErrPending     = errors.New("pending approval")
	ErrNoLicence   = errors.New("no licence document")
)

// OverlapError is a new facility's location touching a live one's.
type OverlapError struct {
	FacilityID string // empty when the other facility is not the caller's own
	Name       string
}

func (e OverlapError) Error() string {
	if e.FacilityID == "" {
		return MsgOverlapsOther
	}
	return MsgOverlaps(e.Name)
}

// Licence is the facility's licence document, kept in the bucket.
type Licence struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	StorageKey  string `json:"-"`
}

type Facility struct {
	ID             string          `json:"id"`
	OrganisationID string          `json:"organisation_id"`
	Name           string          `json:"name"`
	Kind           string          `json:"kind"`
	Location       json.RawMessage `json:"location"` // GeoJSON geometry
	AreaHa         float64         `json:"area_ha"`  // computed from the boundary; 0 for a pin
	LicenceNumber  string          `json:"licence_number"`
	Capacity       int64           `json:"capacity"`
	CapacityUnit   string          `json:"capacity_unit"`
	CapacityPeriod string          `json:"capacity_period"`
	// CapacityUsed is what was originated here in the current period.
	CapacityUsed    int64      `json:"capacity_used"`
	Status          string     `json:"status"`
	LicenceValidTo  *string    `json:"licence_valid_to"` // 2027-12-31; null for facilities registered before it was asked
	LicenceDocument *Licence   `json:"licence_document"`
	ApprovedAt      *time.Time `json:"approved_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// List is what the list endpoint lets callers search and sort by.
var List = listq.Config{
	Alias:  "f",
	Search: []string{"name", "licence_number", "kind"},
	Sort:   []string{"name", "kind", "status", "capacity", listq.ColCreatedAt},
}

const (
	// Area is measured on the globe (geography), in hectares.
	columns = `f.id::text, f.organisation_id::text, f.name, f.kind, ST_AsGeoJSON(f.geom),
	           ST_Area(f.geom::geography) / 10000, f.licence_number, f.capacity,
	           f.capacity_unit, f.capacity_period, f.status, to_char(f.licence_valid_to, 'YYYY-MM-DD'),
	           f.licence_file_name, f.licence_content_type, f.licence_size_bytes, f.licence_sha256,
	           f.licence_storage_key, f.approved_at, f.created_at, f.updated_at`
	gateColumn = "f.organisation_id"
)

func scan(row pgx.CollectableRow) (Facility, error) {
	var (
		f                     Facility
		geo                   string
		name, ctype, sum, key *string
		size                  *int64
	)
	err := row.Scan(&f.ID, &f.OrganisationID, &f.Name, &f.Kind, &geo, &f.AreaHa, &f.LicenceNumber,
		&f.Capacity, &f.CapacityUnit, &f.CapacityPeriod, &f.Status, &f.LicenceValidTo,
		&name, &ctype, &size, &sum, &key, &f.ApprovedAt, &f.CreatedAt, &f.UpdatedAt)
	f.Location = json.RawMessage(geo)
	if name != nil {
		f.LicenceDocument = &Licence{FileName: *name, ContentType: *ctype, SizeBytes: *size, SHA256: *sum, StorageKey: *key}
	}
	return f, err
}

func one(rows pgx.Rows, err error) (Facility, error) {
	if err != nil {
		return Facility{}, err
	}
	f, err := pgx.CollectExactlyOneRow(rows, scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Facility{}, ErrNotFound
	}
	return f, err
}

type Store struct {
	Pool *pgxpool.Pool
}

// record is what a facility's data hash covers: the facts the business
// submitted, as submitted. The same record always gives the same hash.
type record struct {
	ID             string          `json:"id"`
	OrganisationID string          `json:"organisation_id"`
	Name           string          `json:"name"`
	Kind           string          `json:"kind"`
	Location       json.RawMessage `json:"location"`
	LicenceNumber  string          `json:"licence_number"`
	Capacity       int64           `json:"capacity"`
	CapacityUnit   string          `json:"capacity_unit"`
	CapacityPeriod string          `json:"capacity_period"`
	LicenceValidTo string          `json:"licence_valid_to,omitempty"` // omitted for facilities from before it was asked
}

// Create stores a facility for the organisation, waiting for approval. It is
// put on the ledger when approved. PostGIS has the last word on the shape: a
// boundary that crosses itself is refused, and so is one that touches any
// other live facility (the same land cannot be registered twice).
func (s Store) Create(ctx context.Context, orgID string, in Input) (Facility, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Facility{}, err
	}
	defer tx.Rollback(ctx)

	// ponytail: one lock for every registration; per-area locks if registrations ever queue up.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, createLock); err != nil {
		return Facility{}, err
	}
	var other, otherName, otherOrg string
	err = tx.QueryRow(ctx,
		fmt.Sprintf(`SELECT f.id::text, f.name, f.organisation_id::text FROM facilities f
		 WHERE NOT f.is_deleted AND ST_Intersects(f.geom, ST_SetSRID(ST_GeomFromGeoJSON($1), %d))
		 ORDER BY f.created_at LIMIT 1`, SRID), string(in.Location)).Scan(&other, &otherName, &otherOrg)
	switch {
	case err == nil && otherOrg == orgID:
		return Facility{}, OverlapError{FacilityID: other, Name: otherName}
	case err == nil:
		return Facility{}, OverlapError{}
	case !errors.Is(err, pgx.ErrNoRows):
		if isGeometryError(err) {
			return Facility{}, ErrBadGeometry
		}
		return Facility{}, err
	}

	var id string
	if err := tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&id); err != nil {
		return Facility{}, err
	}
	hash, err := chain.DataHash(record{id, orgID, in.Name, in.Kind, in.Location, in.LicenceNumber,
		in.Capacity, in.CapacityUnit, in.CapacityPeriod, in.LicenceValidTo})
	if err != nil {
		return Facility{}, err
	}

	f, err := one(tx.Query(ctx,
		fmt.Sprintf(`WITH g AS (SELECT ST_SetSRID(ST_GeomFromGeoJSON($4), %d) AS geom)
		 INSERT INTO facilities AS f (id, organisation_id, name, kind, geom, licence_number,
		                              capacity, capacity_unit, capacity_period, status, data_hash, licence_valid_to)
		 SELECT $10::uuid, $1::uuid, $2, $3, g.geom, $5, $6, $7, $8, $9, $11, $12::date FROM g WHERE ST_IsValid(g.geom)
		 RETURNING `, SRID)+columns,
		orgID, in.Name, in.Kind, string(in.Location), in.LicenceNumber,
		in.Capacity, in.CapacityUnit, in.CapacityPeriod, StatusPending, id, chain.Hex(hash), in.LicenceValidTo))
	if errors.Is(err, ErrNotFound) {
		return Facility{}, ErrBadGeometry // the WHERE filtered the row out
	}
	if err != nil {
		return Facility{}, err
	}
	return f, tx.Commit(ctx)
}

// isGeometryError reports PostGIS refusing to read the GeoJSON at all.
func isGeometryError(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "XX000"
}

// Approve makes a pending facility active and puts it on the ledger. Only
// the platform approves, so there is no organisation gate.
func (s Store) Approve(ctx context.Context, id, userID string) (Facility, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Facility{}, err
	}
	defer tx.Rollback(ctx)

	var status, orgID, dataHash string
	var capacity int64
	err = tx.QueryRow(ctx,
		`SELECT status, organisation_id::text, data_hash, capacity FROM facilities
		 WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, id).Scan(&status, &orgID, &dataHash, &capacity)
	if errors.Is(err, pgx.ErrNoRows) {
		return Facility{}, ErrNotFound
	}
	if err != nil {
		return Facility{}, err
	}
	if status != StatusPending {
		return Facility{}, ErrNotPending
	}
	f, err := one(tx.Query(ctx,
		`UPDATE facilities AS f SET status = $2, approved_by = $3::uuid, approved_at = now()
		 WHERE f.id = $1::uuid RETURNING `+columns, id, StatusActive, userID))
	if err != nil {
		return Facility{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindFacilityRegister, id, chain.FacilityRegister{
		FacilityID: id, OrgID: orgID, DataHash: dataHash, Capacity: capacity,
	})
	if err != nil {
		return Facility{}, err
	}
	return f, tx.Commit(ctx)
}

// SetLicence records the facility's licence document, and its new end date if
// validTo is not "" (a renewed licence). It returns the storage key of the
// document it replaces ("" if none), for the caller to remove. The facility's
// data hash keeps the facts as first registered.
func (s Store) SetLicence(ctx context.Context, id, gate string, l Licence, validTo string) (Facility, string, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Facility{}, "", err
	}
	defer tx.Rollback(ctx)

	var old *string
	err = tx.QueryRow(ctx,
		`SELECT f.licence_storage_key FROM facilities f
		 WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2)+` FOR UPDATE`, id, gate).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		return Facility{}, "", ErrNotFound
	}
	if err != nil {
		return Facility{}, "", err
	}
	f, err := one(tx.Query(ctx,
		`UPDATE facilities AS f SET licence_file_name = $2, licence_content_type = $3, licence_size_bytes = $4,
		        licence_sha256 = $5, licence_storage_key = $6,
		        licence_valid_to = COALESCE(NULLIF($7, '')::date, f.licence_valid_to)
		 WHERE f.id = $1::uuid RETURNING `+columns, id, l.FileName, l.ContentType, l.SizeBytes, l.SHA256, l.StorageKey, validTo))
	if err != nil {
		return Facility{}, "", err
	}
	replaced := ""
	if old != nil {
		replaced = *old
	}
	return f, replaced, tx.Commit(ctx)
}

// Find returns facilities visible through the gate, optionally by status and map window.
func (s Store) Find(ctx context.Context, q listq.Query, gate, status, kind string, bbox *BBox) (listq.Page[Facility], error) {
	where := []string{auth.GateSQL(gateColumn, 1), `($2 = '' OR f.status = $2)`, `($3 = '' OR f.kind = $3)`}
	args := []any{gate, status, kind}
	if bbox != nil {
		where = append(where, fmt.Sprintf(`f.geom && ST_MakeEnvelope($4, $5, $6, $7, %d)`, SRID))
		args = append(args, bbox[0], bbox[1], bbox[2], bbox[3])
	}
	page, err := listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM facilities f", where, args, scan)
	if err != nil {
		return page, err
	}
	return page, s.withUsed(ctx, page.Data)
}

func (s Store) Get(ctx context.Context, id, gate string) (Facility, error) {
	f, err := one(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM facilities f
		 WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2), id, gate))
	if err != nil {
		return f, err
	}
	fs := []Facility{f}
	err = s.withUsed(ctx, fs)
	return fs[0], err
}

// SetStatus suspends or reactivates a facility and queues the change for the
// chain. A pending facility is not on the ledger yet: only approval changes it.
func (s Store) SetStatus(ctx context.Context, id, gate, status string) (Facility, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Facility{}, err
	}
	defer tx.Rollback(ctx)

	var was string
	err = tx.QueryRow(ctx,
		`SELECT f.status FROM facilities f
		 WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2)+` FOR UPDATE`, id, gate).Scan(&was)
	if errors.Is(err, pgx.ErrNoRows) {
		return Facility{}, ErrNotFound
	}
	if err != nil {
		return Facility{}, err
	}
	if was == StatusPending {
		return Facility{}, ErrPending
	}
	f, err := one(tx.Query(ctx,
		`UPDATE facilities AS f SET status = $3
		 WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2)+`
		 RETURNING `+columns, id, gate, status))
	if err != nil {
		return Facility{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindFacilitySuspended, id, chain.FacilitySuspended{
		FacilityID: id, Suspended: status == StatusSuspended,
	})
	if err != nil {
		return Facility{}, err
	}
	return f, tx.Commit(ctx)
}

// Delete marks the facility deleted; it stays in the database for history.
func (s Store) Delete(ctx context.Context, id, gate string) error {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE facilities AS f SET is_deleted = true
		 WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2), id, gate)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
