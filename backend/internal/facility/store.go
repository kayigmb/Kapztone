package facility

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrBadGeometry = errors.New("invalid geometry")
)

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
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
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
	           f.capacity_unit, f.capacity_period, f.status, f.created_at, f.updated_at`
	gateColumn = "f.organisation_id"
)

func scan(row pgx.CollectableRow) (Facility, error) {
	var (
		f   Facility
		geo string
	)
	err := row.Scan(&f.ID, &f.OrganisationID, &f.Name, &f.Kind, &geo, &f.AreaHa, &f.LicenceNumber,
		&f.Capacity, &f.CapacityUnit, &f.CapacityPeriod, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	f.Location = json.RawMessage(geo)
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
}

// Create stores a facility for the organisation and queues its registration
// on-chain. PostGIS has the last word on the shape: a boundary that crosses
// itself is refused.
func (s Store) Create(ctx context.Context, orgID string, in Input) (Facility, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Facility{}, err
	}
	defer tx.Rollback(ctx)

	var id string
	if err := tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&id); err != nil {
		return Facility{}, err
	}
	hash, err := chain.DataHash(record{id, orgID, in.Name, in.Kind, in.Location, in.LicenceNumber,
		in.Capacity, in.CapacityUnit, in.CapacityPeriod})
	if err != nil {
		return Facility{}, err
	}

	f, err := one(tx.Query(ctx,
		fmt.Sprintf(`WITH g AS (SELECT ST_SetSRID(ST_GeomFromGeoJSON($4), %d) AS geom)
		 INSERT INTO facilities AS f (id, organisation_id, name, kind, geom, licence_number,
		                              capacity, capacity_unit, capacity_period, status, data_hash)
		 SELECT $10::uuid, $1::uuid, $2, $3, g.geom, $5, $6, $7, $8, $9, $11 FROM g WHERE ST_IsValid(g.geom)
		 RETURNING `, SRID)+columns,
		orgID, in.Name, in.Kind, string(in.Location), in.LicenceNumber,
		in.Capacity, in.CapacityUnit, in.CapacityPeriod, StatusActive, id, chain.Hex(hash)))
	if errors.Is(err, ErrNotFound) {
		return Facility{}, ErrBadGeometry // the WHERE filtered the row out
	}
	if err != nil {
		return Facility{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindFacilityRegister, id, chain.FacilityRegister{
		FacilityID: id, OrgID: orgID, DataHash: chain.Hex(hash), Capacity: in.Capacity,
	})
	if err != nil {
		return Facility{}, err
	}
	return f, tx.Commit(ctx)
}

// Find returns facilities visible through the gate, optionally by status and map window.
func (s Store) Find(ctx context.Context, q listq.Query, gate, status string, bbox *BBox) (listq.Page[Facility], error) {
	where := []string{auth.GateSQL(gateColumn, 1), `($2 = '' OR f.status = $2)`}
	args := []any{gate, status}
	if bbox != nil {
		where = append(where, fmt.Sprintf(`f.geom && ST_MakeEnvelope($3, $4, $5, $6, %d)`, SRID))
		args = append(args, bbox[0], bbox[1], bbox[2], bbox[3])
	}
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM facilities f", where, args, scan)
}

func (s Store) Get(ctx context.Context, id, gate string) (Facility, error) {
	return one(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM facilities f
		 WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2), id, gate))
}

// SetStatus suspends or reactivates a facility and queues the change for the chain.
func (s Store) SetStatus(ctx context.Context, id, gate, status string) (Facility, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Facility{}, err
	}
	defer tx.Rollback(ctx)

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
