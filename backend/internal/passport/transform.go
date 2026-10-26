package passport

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

var (
	ErrInputNotHeld = errors.New("input not held")
	ErrMixedInputs  = errors.New("mixed inputs")
)

// InsufficientError says which input does not have the quantity asked of it.
type InsufficientError struct {
	PassportID           string
	Requested, Remaining int64
	Unit                 string
}

func (e InsufficientError) Error() string {
	return fmt.Sprintf("passport %s has %d %s left; %d %s was asked for",
		e.PassportID, e.Remaining, e.Unit, e.Requested, e.Unit)
}

// YieldError says by how much the outputs exceed what the inputs can yield.
type YieldError struct {
	Outputs, Max int64
	Unit         string
}

func (e YieldError) Error() string {
	return fmt.Sprintf("outputs total %d %s; at most %d %s can be produced from these inputs at this stage",
		e.Outputs, e.Unit, e.Max, e.Unit)
}

// TransformInput is what a holder submits to process, split or merge
// passports. Splitting and merging are transforms at a stage whose maximum
// yield is 1.
type TransformInput struct {
	Stage   string            `json:"stage"`
	Inputs  []TransformSource `json:"inputs"`
	Outputs []TransformOutput `json:"outputs"`
}

// TransformSource is how much of one passport goes into the transform.
type TransformSource struct {
	PassportID string `json:"passport_id"`
	Quantity   int64  `json:"quantity"`
}

type TransformOutput struct {
	Quantity int64 `json:"quantity"`
}

func (in *TransformInput) Validate() string {
	if in.Stage = strings.TrimSpace(in.Stage); in.Stage == "" {
		in.Stage = DefaultTransformStage
	}
	switch {
	case len(in.Stage) > MaxStageLen:
		return fmt.Sprintf("stage may be up to %d characters", MaxStageLen)
	case len(in.Inputs) == 0 || len(in.Inputs) > MaxTransformItems:
		return fmt.Sprintf("inputs: give 1 to %d passports", MaxTransformItems)
	case len(in.Outputs) == 0 || len(in.Outputs) > MaxTransformItems:
		return fmt.Sprintf("outputs: give 1 to %d quantities", MaxTransformItems)
	}
	seen := map[string]bool{}
	for _, src := range in.Inputs {
		id := strings.ToLower(src.PassportID)
		switch {
		case !httpx.IsUUID(id):
			return "inputs: passport_id is required"
		case seen[id]:
			return "inputs: a passport is listed twice"
		case src.Quantity <= 0:
			return "inputs: quantity must be a whole number above 0"
		}
		seen[id] = true
	}
	for _, out := range in.Outputs {
		if out.Quantity <= 0 {
			return "outputs: quantity must be a whole number above 0"
		}
	}
	return ""
}

// Transform is the record of one transform.
type Transform struct {
	ID             string `json:"id"`
	OrganisationID string `json:"organisation_id"`
	Commodity      string `json:"commodity"`
	Stage          string `json:"stage"`
	MaxYieldPPM    int64  `json:"max_yield_ppm"` // the rule as applied, in parts per million
	DataHash       string `json:"data_hash"`
	// TemplateVersion is the version of the commodity's template it was judged by.
	TemplateVersion *int       `json:"template_version"`
	TxHash          *string    `json:"tx_hash"`
	AnchoredAt      *time.Time `json:"anchored_at"`
	CreatedAt       time.Time  `json:"created_at"`
	Inputs          []Used     `json:"inputs"`
	Outputs         []Passport `json:"outputs"`
}

// Used is one input of a transform. It is also what the data hash covers.
type Used struct {
	PassportID   string `json:"passport_id"`
	QuantityUsed int64  `json:"quantity_used"`
}

// made is one output as the data hashes cover it.
type made struct {
	PassportID string `json:"passport_id"`
	Quantity   int64  `json:"quantity"`
}

// transformRecord is what a transform's data hash covers.
type transformRecord struct {
	ID             string `json:"id"`
	OrganisationID string `json:"organisation_id"`
	Commodity      string `json:"commodity"`
	Stage          string `json:"stage"`
	MaxYieldPPM    int64  `json:"max_yield_ppm"`
	Inputs         []Used `json:"inputs"`
	Outputs        []made `json:"outputs"`
}

// outputRecord is what the data hash of a passport made by a transform covers.
type outputRecord struct {
	ID             string `json:"id"`
	OrganisationID string `json:"organisation_id"`
	TransformID    string `json:"transform_id"`
	TemplateID     string `json:"template_id"`
	Commodity      string `json:"commodity"`
	Unit           string `json:"unit"`
	Quantity       int64  `json:"quantity"`
}

// Transform uses (parts of) input passports to create output passports,
// after checking every rule the contract will check again: the inputs are
// active, held by the acting organisation and of one commodity, none is
// overdrawn, and the outputs do not exceed the inputs.
func (s Store) Transform(ctx context.Context, orgID, userID string, in TransformInput, now time.Time) (Transform, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Transform{}, err
	}
	defer tx.Rollback(ctx)

	t, err := transformTx(ctx, tx, orgID, userID, in, now)
	if err != nil {
		return Transform{}, err
	}
	return t, tx.Commit(ctx)
}

// transformTx is Transform in the caller's transaction: a handover of part of
// a lot cuts that part into its own passport this way first.
func transformTx(ctx context.Context, tx pgx.Tx, orgID, userID string, in TransformInput, now time.Time) (Transform, error) {
	j := &transformJob{tx: tx, orgID: orgID, userID: userID, in: in, now: now}
	locked, err := lockInputs(ctx, tx, in.Inputs)
	if err != nil {
		return Transform{}, err
	}
	totalIn, err := j.checkInputs(locked)
	if err != nil {
		return Transform{}, err
	}
	if err := j.rule(ctx); err != nil {
		return Transform{}, err
	}
	if err := j.checkYield(totalIn); err != nil {
		return Transform{}, err
	}
	if err := j.plan(ctx); err != nil {
		return Transform{}, err
	}
	return j.write(ctx)
}

// transformJob carries one transform from its checks to its rows.
type transformJob struct {
	tx                          pgx.Tx
	orgID, userID               string
	in                          TransformInput
	now                         time.Time
	commodity, unit, templateID string
	ppm                         int64
	used                        []Used
	id, dataHash                string
	queued                      []chain.TransformOutput
}

type heldInput struct {
	holder, status, commodity, unit string
	remaining                       int64
}

// lockInputs locks the input passports, always in id order so two transforms cannot deadlock.
func lockInputs(ctx context.Context, tx pgx.Tx, inputs []TransformSource) (map[string]heldInput, error) {
	ids := make([]string, len(inputs))
	for i, src := range inputs {
		ids[i] = src.PassportID
	}
	rows, err := tx.Query(ctx,
		`SELECT id::text, organisation_id::text, status, commodity, unit, remaining FROM passports
		 WHERE id = ANY($1::uuid[]) AND NOT is_deleted ORDER BY id FOR UPDATE`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	locked := map[string]heldInput{}
	for rows.Next() {
		var id string
		var h heldInput
		if err := rows.Scan(&id, &h.holder, &h.status, &h.commodity, &h.unit, &h.remaining); err != nil {
			return nil, err
		}
		locked[id] = h
	}
	return locked, rows.Err()
}

// checkInputs: every input is held and active, all of one commodity and unit,
// with enough left. It returns the total going in.
func (j *transformJob) checkInputs(locked map[string]heldInput) (totalIn int64, err error) {
	j.used = make([]Used, len(j.in.Inputs))
	for i, src := range j.in.Inputs {
		id := strings.ToLower(src.PassportID)
		h, ok := locked[id]
		switch {
		case !ok || h.holder != j.orgID:
			return 0, ErrInputNotHeld
		case h.status != StatusActive:
			return 0, ErrNotActive
		case i > 0 && (h.commodity != j.commodity || h.unit != j.unit):
			return 0, ErrMixedInputs
		case src.Quantity > h.remaining:
			return 0, InsufficientError{PassportID: id, Requested: src.Quantity, Remaining: h.remaining, Unit: h.unit}
		}
		j.commodity, j.unit = h.commodity, h.unit
		totalIn += src.Quantity
		j.used[i] = Used{PassportID: id, QuantityUsed: src.Quantity}
	}
	return totalIn, nil
}

// rule: whatever the label, a transform keeps at most what went in (the
// platform's split yield of 1). A template published before splits existed
// gets that yield queued now, ahead of this transform.
func (j *transformJob) rule(ctx context.Context) error {
	var err error
	if j.templateID, err = liveTemplate(ctx, j.tx, j.commodity); err != nil {
		return err
	}
	j.ppm = chain.YieldScale
	return template.QueueSplitYield(ctx, j.tx, j.templateID, j.commodity)
}

// checkYield is the mass balance, compared without dividing exactly as the
// contract does, so rounding can never favour the caller.
func (j *transformJob) checkYield(totalIn int64) error {
	var totalOut int64
	for _, out := range j.in.Outputs {
		totalOut += out.Quantity
	}
	allowed := new(big.Int).Mul(big.NewInt(totalIn), big.NewInt(j.ppm))
	if new(big.Int).Mul(big.NewInt(totalOut), big.NewInt(chain.YieldScale)).Cmp(allowed) > 0 {
		return YieldError{Outputs: totalOut, Max: allowed.Div(allowed, big.NewInt(chain.YieldScale)).Int64(), Unit: j.unit}
	}
	return nil
}

// plan gives the transform and each output an id and a data hash.
func (j *transformJob) plan(ctx context.Context) error {
	var err error
	if j.id, err = newID(ctx, j.tx); err != nil {
		return err
	}
	outputs := make([]made, len(j.in.Outputs))
	j.queued = make([]chain.TransformOutput, len(j.in.Outputs))
	for i, out := range j.in.Outputs {
		outID, err := newID(ctx, j.tx)
		if err != nil {
			return err
		}
		hash, err := chain.DataHash(outputRecord{outID, j.orgID, j.id, j.templateID, j.commodity, j.unit, out.Quantity})
		if err != nil {
			return err
		}
		outputs[i] = made{PassportID: outID, Quantity: out.Quantity}
		j.queued[i] = chain.TransformOutput{PassportID: outID, Quantity: out.Quantity, DataHash: chain.Hex(hash)}
	}
	hash, err := chain.DataHash(transformRecord{j.id, j.orgID, j.commodity, j.in.Stage, j.ppm, j.used, outputs})
	j.dataHash = chain.Hex(hash)
	return err
}

func (j *transformJob) write(ctx context.Context) (Transform, error) {
	t := Transform{ID: j.id, OrganisationID: j.orgID, Commodity: j.commodity, Stage: j.in.Stage, MaxYieldPPM: j.ppm,
		DataHash: j.dataHash, Inputs: j.used}
	err := j.tx.QueryRow(ctx,
		`INSERT INTO transforms (id, organisation_id, commodity, stage, max_yield_ppm, data_hash, created_by)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7::uuid) RETURNING created_at, template_version`,
		j.id, j.orgID, j.commodity, j.in.Stage, j.ppm, j.dataHash, j.userID).Scan(&t.CreatedAt, &t.TemplateVersion)
	if err != nil {
		return Transform{}, err
	}
	if err := j.consumeInputs(ctx); err != nil {
		return Transform{}, err
	}
	if t.Outputs, err = j.createOutputs(ctx); err != nil {
		return Transform{}, err
	}
	// The contract finds the yield by stage: always the split stage's.
	payload := chain.Transform{TransformID: j.id, Stage: template.StageSplit, Outputs: j.queued, DataHash: j.dataHash}
	for _, u := range j.used {
		payload.Inputs = append(payload.Inputs, u.PassportID)
		payload.InputsUsed = append(payload.InputsUsed, u.QuantityUsed)
	}
	return t, chain.Enqueue(ctx, j.tx, chain.KindTransform, j.id, payload)
}

const transformEvent = `INSERT INTO trace_events (passport_id, ref_id, kind, from_org, to_org, quantity, data_hash)
	VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)`

// consumeInputs takes what was used from each input; one used up entirely is
// consumed, as the contract marks it.
func (j *transformJob) consumeInputs(ctx context.Context) error {
	for _, u := range j.used {
		if _, err := j.tx.Exec(ctx,
			`UPDATE passports SET remaining = remaining - $2,
			        status = CASE WHEN remaining = $2 THEN $3 ELSE status END
			 WHERE id = $1::uuid`, u.PassportID, u.QuantityUsed, StatusConsumed); err != nil {
			return err
		}
		if _, err := j.tx.Exec(ctx,
			`INSERT INTO transform_inputs (transform_id, passport_id, quantity_used) VALUES ($1::uuid, $2::uuid, $3)`,
			j.id, u.PassportID, u.QuantityUsed); err != nil {
			return err
		}
		if _, err := j.tx.Exec(ctx, transformEvent, u.PassportID, j.id, EventTransform, j.orgID, nil, u.QuantityUsed, j.dataHash); err != nil {
			return err
		}
	}
	return nil
}

func (j *transformJob) createOutputs(ctx context.Context) ([]Passport, error) {
	var out []Passport
	for _, q := range j.queued {
		code, err := nextCode(ctx, j.tx, j.orgID, j.now)
		if err != nil {
			return nil, err
		}
		p, err := onePassport(j.tx.Query(ctx,
			`INSERT INTO passports AS p (id, organisation_id, transform_id, template_id, commodity, unit,
			                             quantity, remaining, status, data_hash, issued_by, code)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $7, $8, $9, $10::uuid, $11)
			 RETURNING `+columns,
			q.PassportID, j.orgID, j.id, j.templateID, j.commodity, j.unit, q.Quantity, StatusActive, q.DataHash, j.userID, code))
		if err != nil {
			return nil, err
		}
		if _, err := j.tx.Exec(ctx, transformEvent, q.PassportID, j.id, EventTransform, nil, j.orgID, q.Quantity, q.DataHash); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// Genealogy is a passport's family tree in one direction.
type Genealogy struct {
	Nodes []Passport `json:"nodes"`
	Edges []Edge     `json:"edges"`
}

// Edge says that quantity_used of the parent went into the transform that
// made the child.
type Edge struct {
	ParentID     string `json:"parent_id"`
	ChildID      string `json:"child_id"`
	TransformID  string `json:"transform_id"`
	Stage        string `json:"stage"`
	QuantityUsed int64  `json:"quantity_used"`
}

// The two walks over the family tree. A passport's parents are the inputs of
// the transform that made it; its children are the passports made by any
// transform it was an input of. $1 is the passport to start from; UNION
// drops repeats, so a passport reached twice is walked once.
const (
	ancestorsCTE = `WITH RECURSIVE tree(id) AS (
	    SELECT $1::uuid
	  UNION
	    SELECT ti.passport_id FROM tree t
	    JOIN passports c ON c.id = t.id
	    JOIN transform_inputs ti ON ti.transform_id = c.transform_id AND NOT ti.is_deleted)`
	descendantsCTE = `WITH RECURSIVE tree(id) AS (
	    SELECT $1::uuid
	  UNION
	    SELECT c.id FROM tree t
	    JOIN transform_inputs ti ON ti.passport_id = t.id AND NOT ti.is_deleted
	    JOIN passports c ON c.transform_id = ti.transform_id AND NOT c.is_deleted)`
)

// Genealogy returns the passport with its ancestors back to the origin
// passports, or with everything made from it. Whoever may see the passport
// may see its tree: the point of a passport is that its history can be checked.
func (s Store) Genealogy(ctx context.Context, id, gate, direction string) (Genealogy, error) {
	if _, err := onePassport(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM passports p WHERE p.id = $1::uuid AND NOT p.is_deleted AND `+visible(2), id, gate)); err != nil {
		return Genealogy{}, err
	}
	cte := ancestorsCTE
	if direction == DirectionDescendants {
		cte = descendantsCTE
	}
	rows, err := s.Pool.Query(ctx, cte+`
		SELECT `+columns+` FROM passports p
		WHERE p.id IN (SELECT id FROM tree) AND NOT p.is_deleted ORDER BY p.created_at, p.id`, id)
	if err != nil {
		return Genealogy{}, err
	}
	g := Genealogy{Edges: []Edge{}}
	if g.Nodes, err = pgx.CollectRows(rows, scan); err != nil {
		return Genealogy{}, err
	}
	edges, err := s.edgesBetween(ctx, g.Nodes)
	if err != nil {
		return Genealogy{}, err
	}
	g.Edges = append(g.Edges, edges...)
	return g, nil
}

// edgesBetween lists the transforms that link these passports, parent to child.
func (s Store) edgesBetween(ctx context.Context, nodes []Passport) ([]Edge, error) {
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT ti.passport_id::text, c.id::text, tr.id::text, tr.stage, ti.quantity_used
		 FROM transform_inputs ti
		 JOIN transforms tr ON tr.id = ti.transform_id
		 JOIN passports c ON c.transform_id = tr.id
		 WHERE ti.passport_id = ANY($1::uuid[]) AND c.id = ANY($1::uuid[]) AND NOT ti.is_deleted
		 ORDER BY tr.created_at, ti.passport_id, c.id`, ids)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Edge, error) {
		var e Edge
		err := row.Scan(&e.ParentID, &e.ChildID, &e.TransformID, &e.Stage, &e.QuantityUsed)
		return e, err
	})
}

// RevokeTree revokes a passport and everything made from it, oldest first,
// in the caller's transaction. refID is the row that caused it (the flag);
// reasonHash is what the chain records as the reason. It returns how many
// passports were revoked; those already revoked are left as they are.
//
// A descendant is revoked even if only part of it came from the revoked
// passport: once tainted goods are mixed in, the lot cannot be trusted.
func RevokeTree(ctx context.Context, tx pgx.Tx, rootID, refID, reasonHash string) (int, error) {
	rows, err := tx.Query(ctx, descendantsCTE+`
		SELECT p.id::text FROM passports p WHERE p.id IN (SELECT id FROM tree) ORDER BY p.created_at, p.id`, rootID)
	if err != nil {
		return 0, err
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return 0, err
	}
	revoked := 0
	for _, id := range ids {
		done, err := revokeOne(ctx, tx, id, refID, reasonHash)
		if err != nil {
			return 0, err
		}
		if done {
			revoked++
		}
	}
	return revoked, nil
}

// revokeOne revokes a passport not yet revoked, records it on its journey,
// queues it for the chain and tells its holder.
func revokeOne(ctx context.Context, tx pgx.Tx, id, refID, reasonHash string) (bool, error) {
	var holder string
	var remaining int64
	err := tx.QueryRow(ctx,
		`UPDATE passports SET status = $2 WHERE id = $1::uuid AND status <> $2 AND NOT is_deleted
		 RETURNING organisation_id::text, remaining`, id, StatusRevoked).Scan(&holder, &remaining)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var eventID string
	err = tx.QueryRow(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, from_org, quantity, data_hash)
		 VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5, $6) RETURNING id::text`,
		id, refID, EventRevoke, holder, remaining, reasonHash).Scan(&eventID)
	if err != nil {
		return false, err
	}
	// The queued item's subject is the journey event, which receives the transaction hash.
	err = chain.Enqueue(ctx, tx, chain.KindPassportRevoke, eventID, chain.PassportReason{PassportID: id, ReasonHash: reasonHash})
	if err != nil {
		return false, err
	}
	return true, notify.Add(ctx, tx, holder, notify.KindPassportRevoked, id, MsgNotifyPassportRevoked)
}

// RecordHash rebuilds, from the rows as they are now, the hash a passport's
// data_hash must equal. A difference means the record was changed after it
// was written. The organisation a record names is the one that created it,
// which is the facility's owner or the transform's, not today's holder.
func (s Store) RecordHash(ctx context.Context, p Passport) (string, error) {
	var (
		creator string
		record  any
	)
	switch {
	case p.TransformID != nil:
		err := s.Pool.QueryRow(ctx, `SELECT organisation_id::text FROM transforms WHERE id = $1::uuid`, *p.TransformID).Scan(&creator)
		if err != nil {
			return "", err
		}
		record = outputRecord{p.ID, creator, *p.TransformID, p.TemplateID, p.Commodity, p.Unit, p.Quantity}
	case p.FacilityID != nil && p.Period != nil:
		err := s.Pool.QueryRow(ctx, `SELECT organisation_id::text FROM facilities WHERE id = $1::uuid`, *p.FacilityID).Scan(&creator)
		if err != nil {
			return "", err
		}
		record = issueRecord{p.ID, creator, *p.FacilityID, p.TemplateID, p.Commodity, p.Unit, p.Quantity, *p.Period}
	default:
		return "", fmt.Errorf("passport %s has neither an origin facility nor a transform", p.ID)
	}
	hash, err := chain.DataHash(record)
	if err != nil {
		return "", err
	}
	return chain.Hex(hash), nil
}
