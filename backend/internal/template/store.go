package template

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/dbx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrCommodityTaken = errors.New("commodity taken")
	ErrInUse          = errors.New("template in use")
)

// Template is a stored commodity template: the submitted rules plus the base fields.
type Template struct {
	ID string `json:"id"`
	Input
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// List is what the list endpoint lets callers search and sort by.
var List = listq.Config{
	Alias:  "t",
	Search: []string{"commodity", "name"},
	Sort:   []string{"commodity", "name", listq.ColCreatedAt},
}

const columns = `t.id::text, t.commodity, t.name, t.unit, t.origin_facility_kind,
                 t.capacity_rule::text, t.stages::text, t.created_at, t.updated_at`

func scan(row pgx.CollectableRow) (Template, error) {
	var (
		t            Template
		rule, stages string
	)
	err := row.Scan(&t.ID, &t.Commodity, &t.Name, &t.Unit, &t.OriginFacilityKind,
		&rule, &stages, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal([]byte(rule), &t.CapacityRule); err != nil {
		return t, err
	}
	return t, json.Unmarshal([]byte(stages), &t.Stages)
}

func one(rows pgx.Rows, err error) (Template, error) {
	if err == nil {
		var t Template
		if t, err = pgx.CollectExactlyOneRow(rows, scan); err == nil {
			return t, nil
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Template{}, ErrNotFound
	}
	if _, taken := dbx.UniqueViolation(err); taken {
		return Template{}, ErrCommodityTaken
	}
	return Template{}, err
}

// jsonArgs returns the two JSON columns of an input as text.
func jsonArgs(in Input) (rule, stages string) {
	r, _ := json.Marshal(in.CapacityRule)
	s, _ := json.Marshal(in.Stages)
	return string(r), string(s)
}

type Store struct {
	Pool *pgxpool.Pool
}

// YieldPPM writes a stage's maximum yield the way the contract holds it.
func YieldPPM(maxYield float64) int64 {
	return int64(math.Round(maxYield * chain.YieldScale))
}

// queueYields tells the chain every yield the template sets. The contract
// refuses a transform at a stage whose yield it has not been told.
func queueYields(ctx context.Context, tx pgx.Tx, t Template) error {
	for _, st := range t.Stages {
		if st.MaxYield == nil {
			continue
		}
		err := chain.Enqueue(ctx, tx, chain.KindMaxYieldSet, t.ID,
			chain.MaxYieldSet{Commodity: t.Commodity, Stage: st.Key, PartsPerMillion: YieldPPM(*st.MaxYield)})
		if err != nil {
			return err
		}
	}
	return nil
}

// save runs one insert or update and queues the template's yields with it.
func (s Store) save(ctx context.Context, check func(pgx.Tx) error, sql string, args ...any) (Template, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Template{}, err
	}
	defer tx.Rollback(ctx)

	if check != nil {
		if err := check(tx); err != nil {
			return Template{}, err
		}
	}
	t, err := one(tx.Query(ctx, sql, args...))
	if err != nil {
		return Template{}, err
	}
	if err := queueYields(ctx, tx, t); err != nil {
		return Template{}, err
	}
	return t, tx.Commit(ctx)
}

func (s Store) Create(ctx context.Context, in Input) (Template, error) {
	rule, stages := jsonArgs(in)
	return s.save(ctx, nil,
		`INSERT INTO commodity_templates AS t (commodity, name, unit, origin_facility_kind, capacity_rule, stages)
		 VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb) RETURNING `+columns,
		in.Commodity, in.Name, in.Unit, in.OriginFacilityKind, rule, stages)
}

func (s Store) Find(ctx context.Context, q listq.Query) (listq.Page[Template], error) {
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM commodity_templates t", nil, nil, scan)
}

func (s Store) Get(ctx context.Context, id string) (Template, error) {
	return one(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM commodity_templates t WHERE t.id = $1::uuid AND NOT t.is_deleted`, id))
}

// Update replaces the template. Passports copy the commodity and unit they
// were issued under, so those two cannot change once passports exist.
func (s Store) Update(ctx context.Context, id string, in Input) (Template, error) {
	rule, stages := jsonArgs(in)
	inUse := func(tx pgx.Tx) error {
		var changes, used bool
		err := tx.QueryRow(ctx,
			`SELECT (t.commodity <> $2 OR t.unit <> $3),
			        EXISTS (SELECT 1 FROM passports p WHERE p.template_id = t.id AND NOT p.is_deleted)
			 FROM commodity_templates t WHERE t.id = $1::uuid AND NOT t.is_deleted FOR UPDATE`,
			id, in.Commodity, in.Unit).Scan(&changes, &used)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err == nil && changes && used {
			return ErrInUse
		}
		return err
	}
	return s.save(ctx, inUse,
		`UPDATE commodity_templates AS t
		 SET commodity = $2, name = $3, unit = $4, origin_facility_kind = $5,
		     capacity_rule = $6::jsonb, stages = $7::jsonb
		 WHERE t.id = $1::uuid AND NOT t.is_deleted RETURNING `+columns,
		id, in.Commodity, in.Name, in.Unit, in.OriginFacilityKind, rule, stages)
}

// Delete marks the template deleted; it stays in the database for history.
func (s Store) Delete(ctx context.Context, id string) error {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE commodity_templates SET is_deleted = true WHERE id = $1::uuid AND NOT is_deleted`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}
