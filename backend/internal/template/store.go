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
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrCommodityTaken = errors.New("commodity taken")
	ErrInUse          = errors.New("template in use")
	ErrNoDraft        = errors.New("no draft")
	ErrUnknownKind    = errors.New("unknown facility kind")
)

// Template is a stored commodity template: the published rules plus the base
type Template struct {
	ID string `json:"id"`
	Input
	Version   int       `json:"version"`
	Draft     *Input    `json:"draft"` // null when nothing waits to be published
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
                 t.capacity_rule::text, t.stages::text, t.version, t.draft::text, t.created_at, t.updated_at`

func scan(row pgx.CollectableRow) (Template, error) {
	var (
		t            Template
		rule, stages string
		draft        *string
	)
	err := row.Scan(&t.ID, &t.Commodity, &t.Name, &t.Unit, &t.OriginFacilityKind,
		&rule, &stages, &t.Version, &draft, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	if draft != nil {
		t.Draft = &Input{}
		if err := json.Unmarshal([]byte(*draft), t.Draft); err != nil {
			return t, err
		}
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

// queueYields tells the chain every yield the template sets, and the split
// stage's yield of 1 (see StageSplit). The contract refuses a transform at a
// stage whose yield it has not been told.
func queueYields(ctx context.Context, tx pgx.Tx, t Template) error {
	whole := 1.0
	stages := append(t.Stages, Stage{Key: StageSplit, MaxYield: &whole})
	for _, st := range stages {
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

// QueueSplitYield queues the split stage's yield for a commodity unless it
// was queued before: templates published before splits existed lack it.
func QueueSplitYield(ctx context.Context, tx pgx.Tx, templateID, commodity string) error {
	var queued bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM chain_outbox WHERE kind = $1
		                AND payload->>'commodity' = $2 AND payload->>'stage' = $3)`,
		chain.KindMaxYieldSet, commodity, StageSplit).Scan(&queued)
	if err != nil || queued {
		return err
	}
	return chain.Enqueue(ctx, tx, chain.KindMaxYieldSet, templateID,
		chain.MaxYieldSet{Commodity: commodity, Stage: StageSplit, PartsPerMillion: chain.YieldScale})
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

// ActiveKinds lists the curated, active facility kinds templates may name.
func (s Store) ActiveKinds(ctx context.Context) ([]facility.Kind, error) {
	kinds, err := (&facility.Store{Pool: s.Pool}).Kinds(ctx)
	if err != nil {
		return nil, err
	}
	out := kinds[:0]
	for _, k := range kinds {
		if k.Active {
			out = append(out, k)
		}
	}
	if out == nil {
		out = []facility.Kind{}
	}
	return out, nil
}

// checkKind refuses an origin facility kind the platform has not curated
// (or has since deactivated). Templates name kinds; they never invent them.
func checkKind(ctx context.Context, tx pgx.Tx, kind, existingTemplateID string) error {
	var key string
	err := tx.QueryRow(ctx, `SELECT key FROM facility_kinds WHERE key = $1 AND active AND NOT is_deleted FOR SHARE`, kind).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		if existingTemplateID != "" {
			var priorKind string
			if err := tx.QueryRow(ctx,
				`SELECT origin_facility_kind FROM commodity_templates WHERE id = $1::uuid AND NOT is_deleted`, existingTemplateID).Scan(&priorKind); err == nil && priorKind == kind {
				// A deactivated kind remains valid for an existing template that
				// already references it; it cannot be chosen for new templates.
				return nil
			} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}
		return ErrUnknownKind
	}
	return err
}

func (s Store) Create(ctx context.Context, in Input) (Template, error) {
	rule, stages := jsonArgs(in)
	return s.save(ctx, func(tx pgx.Tx) error { return checkKind(ctx, tx, in.OriginFacilityKind, "") },
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

// inUse refuses a change of commodity or unit once passports use the
// template: passports copy the two they were issued under. It locks the row.
func inUse(ctx context.Context, tx pgx.Tx, id string, in Input) error {
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

// Update saves the template's next version as a draft. The published rules,
// which every check uses, do not change until Publish.
func (s Store) Update(ctx context.Context, id string, in Input) (Template, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Template{}, err
	}
	defer tx.Rollback(ctx)
	if err := inUse(ctx, tx, id, in); err != nil {
		return Template{}, err
	}
	if err := checkKind(ctx, tx, in.OriginFacilityKind, id); err != nil {
		return Template{}, err
	}
	draft, _ := json.Marshal(in)
	t, err := one(tx.Query(ctx,
		`UPDATE commodity_templates AS t SET draft = $2::jsonb
		 WHERE t.id = $1::uuid AND NOT t.is_deleted RETURNING `+columns, id, string(draft)))
	if err != nil {
		return Template{}, err
	}
	return t, tx.Commit(ctx)
}

// Publish makes the draft the template's rules, as the next version, and
// queues its yields for the chain. Transforms and handovers record the
// version they were judged by.
func (s Store) Publish(ctx context.Context, id string) (Template, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Template{}, err
	}
	defer tx.Rollback(ctx)
	in, err := lockDraft(ctx, tx, id)
	if err != nil {
		return Template{}, err
	}
	if err := checkKind(ctx, tx, in.OriginFacilityKind, id); err != nil {
		return Template{}, err
	}
	if err := inUse(ctx, tx, id, in); err != nil {
		return Template{}, err
	}
	rule, stages := jsonArgs(in)
	t, err := one(tx.Query(ctx,
		`UPDATE commodity_templates AS t
		 SET commodity = $2, name = $3, unit = $4, origin_facility_kind = $5,
		     capacity_rule = $6::jsonb, stages = $7::jsonb, version = t.version + 1, draft = NULL
		 WHERE t.id = $1::uuid RETURNING `+columns,
		id, in.Commodity, in.Name, in.Unit, in.OriginFacilityKind, rule, stages))
	if err != nil {
		return Template{}, err
	}
	if err := queueYields(ctx, tx, t); err != nil {
		return Template{}, err
	}
	return t, tx.Commit(ctx)
}

func lockDraft(ctx context.Context, tx pgx.Tx, id string) (Input, error) {
	var in Input
	var draft *string
	err := tx.QueryRow(ctx,
		`SELECT draft::text FROM commodity_templates WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, id).Scan(&draft)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return in, ErrNotFound
	case err != nil:
		return in, err
	case draft == nil:
		return in, ErrNoDraft
	}
	return in, json.Unmarshal([]byte(*draft), &in)
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
