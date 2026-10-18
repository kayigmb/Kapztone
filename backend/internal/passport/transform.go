package passport

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

var (
	ErrInputNotHeld = errors.New("input not held")
	ErrMixedInputs  = errors.New("mixed inputs")
	ErrStageInvalid = errors.New("stage invalid")
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
	in.Stage = strings.TrimSpace(in.Stage)
	switch {
	case in.Stage == "":
		return "stage is required"
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

// transformStage finds the commodity's live template and one of its stages
// that transforms with a maximum yield.
func transformStage(ctx context.Context, tx pgx.Tx, commodity, stage string) (templateID string, st template.Stage, err error) {
	templateID, stages, err := liveStages(ctx, tx, commodity)
	if err != nil {
		return "", st, err
	}
	for _, st := range stages {
		if st.Key == stage && st.MaxYield != nil && slices.Contains(st.Events, EventTransform) {
			return templateID, st, nil
		}
	}
	return "", st, ErrStageInvalid
}

// Transform uses (parts of) input passports to create output passports,
// after checking every rule the contract will check again: the inputs are
// active, held by the acting organisation and of one commodity, none is
// overdrawn, and the outputs do not exceed inputs × the stage's maximum yield.
//
// Outputs of at least RuleYieldShare of the maximum are allowed, but at a
// stage that loses mass the rules raise a flag about the first output in the
// same transaction.
func (s Store) Transform(ctx context.Context, orgID, userID string, in TransformInput, now time.Time, raise RaiseFlag) (Transform, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Transform{}, err
	}
	defer tx.Rollback(ctx)

	t, err := transformTx(ctx, tx, orgID, userID, in, now, false)
	if err != nil {
		return Transform{}, err
	}
	var totalIn, totalOut int64
	for _, u := range t.Inputs {
		totalIn += u.QuantityUsed
	}
	for _, p := range t.Outputs {
		totalOut += p.Quantity
	}
	// share = out / (in × yield); compared in floats, it only decides a flag.
	share := float64(totalOut) * chain.YieldScale / (float64(totalIn) * float64(t.MaxYieldPPM))
	// A stage that keeps all the mass (sorting, splitting) always yields 100%: never flagged.
	if raise != nil && t.MaxYieldPPM < chain.YieldScale && share >= RuleYieldShare && len(t.Outputs) > 0 {
		reason := fmt.Sprintf(MsgRuleYieldFmt, share*100, t.Stage)
		if err := raise(ctx, tx, orgID, userID, orgID, t.Outputs[0].ID, reason, SourceRules); err != nil {
			return Transform{}, err
		}
	}
	return t, tx.Commit(ctx)
}

// transformTx is Transform in the caller's transaction. With split, the stage
// is the platform's own split stage (yield 1, no evidence): a handover of part
// of a lot cuts that part into its own passport this way first.
func transformTx(ctx context.Context, tx pgx.Tx, orgID, userID string, in TransformInput, now time.Time, split bool) (Transform, error) {
	// Lock the inputs, always in id order so two transforms cannot deadlock.
	ids := make([]string, len(in.Inputs))
	for i, src := range in.Inputs {
		ids[i] = src.PassportID
	}
	type held struct {
		holder, status, commodity, unit string
		remaining                       int64
	}
	rows, err := tx.Query(ctx,
		`SELECT id::text, organisation_id::text, status, commodity, unit, remaining FROM passports
		 WHERE id = ANY($1::uuid[]) AND NOT is_deleted ORDER BY id FOR UPDATE`, ids)
	if err != nil {
		return Transform{}, err
	}
	locked := map[string]held{}
	for rows.Next() {
		var id string
		var h held
		if err := rows.Scan(&id, &h.holder, &h.status, &h.commodity, &h.unit, &h.remaining); err != nil {
			rows.Close()
			return Transform{}, err
		}
		locked[id] = h
	}
	if err := rows.Err(); err != nil {
		return Transform{}, err
	}

	var (
		commodity, unit string
		totalIn         int64
		used            = make([]Used, len(in.Inputs))
	)
	for i, src := range in.Inputs {
		id := strings.ToLower(src.PassportID)
		h, ok := locked[id]
		switch {
		case !ok || h.holder != orgID:
			return Transform{}, ErrInputNotHeld
		case h.status != StatusActive:
			return Transform{}, ErrNotActive
		case i > 0 && (h.commodity != commodity || h.unit != unit):
			return Transform{}, ErrMixedInputs
		case src.Quantity > h.remaining:
			return Transform{}, InsufficientError{PassportID: id, Requested: src.Quantity, Remaining: h.remaining, Unit: h.unit}
		}
		commodity, unit = h.commodity, h.unit
		totalIn += src.Quantity
		used[i] = Used{PassportID: id, QuantityUsed: src.Quantity}
	}

	var (
		templateID string
		stage      template.Stage
	)
	if split {
		if templateID, _, err = liveStages(ctx, tx, commodity); err != nil {
			return Transform{}, err
		}
		whole := 1.0
		stage = template.Stage{Key: template.StageSplit, MaxYield: &whole}
		// Publishing a template queues the split yield; one published before
		// splits existed gets it now, ahead of this transform.
		err = template.QueueSplitYield(ctx, tx, templateID, commodity)
	} else {
		templateID, stage, err = transformStage(ctx, tx, commodity, in.Stage)
	}
	if err != nil {
		return Transform{}, err
	}
	ppm := template.YieldPPM(*stage.MaxYield)
	// The stage's evidence must be on every input.
	for _, u := range used {
		if err := requireEvidence(ctx, tx, stage, attestation.SubjectPassport, u.PassportID); err != nil {
			return Transform{}, err
		}
	}

	// Mass balance, compared without dividing exactly as the contract does,
	// so rounding can never favour the caller.
	var totalOut int64
	for _, out := range in.Outputs {
		totalOut += out.Quantity
	}
	allowed := new(big.Int).Mul(big.NewInt(totalIn), big.NewInt(ppm))
	if new(big.Int).Mul(big.NewInt(totalOut), big.NewInt(chain.YieldScale)).Cmp(allowed) > 0 {
		return Transform{}, YieldError{Outputs: totalOut, Max: allowed.Div(allowed, big.NewInt(chain.YieldScale)).Int64(), Unit: unit}
	}

	id, err := newID(ctx, tx)
	if err != nil {
		return Transform{}, err
	}
	outputs := make([]made, len(in.Outputs))
	queued := make([]chain.TransformOutput, len(in.Outputs))
	for i, out := range in.Outputs {
		outID, err := newID(ctx, tx)
		if err != nil {
			return Transform{}, err
		}
		hash, err := chain.DataHash(outputRecord{outID, orgID, id, templateID, commodity, unit, out.Quantity})
		if err != nil {
			return Transform{}, err
		}
		outputs[i] = made{PassportID: outID, Quantity: out.Quantity}
		queued[i] = chain.TransformOutput{PassportID: outID, Quantity: out.Quantity, DataHash: chain.Hex(hash)}
	}
	hash, err := chain.DataHash(transformRecord{id, orgID, commodity, in.Stage, ppm, used, outputs})
	if err != nil {
		return Transform{}, err
	}
	dataHash := chain.Hex(hash)

	t := Transform{ID: id, OrganisationID: orgID, Commodity: commodity, Stage: in.Stage, MaxYieldPPM: ppm,
		DataHash: dataHash, Inputs: used}
	err = tx.QueryRow(ctx,
		`INSERT INTO transforms (id, organisation_id, commodity, stage, max_yield_ppm, data_hash, created_by)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7::uuid) RETURNING created_at, template_version`,
		id, orgID, commodity, in.Stage, ppm, dataHash, userID).Scan(&t.CreatedAt, &t.TemplateVersion)
	if err != nil {
		return Transform{}, err
	}

	const event = `INSERT INTO trace_events (passport_id, ref_id, kind, from_org, to_org, quantity, data_hash)
	               VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)`
	for _, u := range used {
		// An input used up entirely is consumed; the contract marks it the same way.
		if _, err := tx.Exec(ctx,
			`UPDATE passports SET remaining = remaining - $2,
			        status = CASE WHEN remaining = $2 THEN $3 ELSE status END
			 WHERE id = $1::uuid`, u.PassportID, u.QuantityUsed, StatusConsumed); err != nil {
			return Transform{}, err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO transform_inputs (transform_id, passport_id, quantity_used) VALUES ($1::uuid, $2::uuid, $3)`,
			id, u.PassportID, u.QuantityUsed); err != nil {
			return Transform{}, err
		}
		if _, err := tx.Exec(ctx, event, u.PassportID, id, EventTransform, orgID, nil, u.QuantityUsed, dataHash); err != nil {
			return Transform{}, err
		}
	}
	for _, q := range queued {
		code, err := nextCode(ctx, tx, orgID, now)
		if err != nil {
			return Transform{}, err
		}
		p, err := onePassport(tx.Query(ctx,
			`INSERT INTO passports AS p (id, organisation_id, transform_id, template_id, commodity, unit,
			                             quantity, remaining, status, data_hash, issued_by, code)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $7, $8, $9, $10::uuid, $11)
			 RETURNING `+columns,
			q.PassportID, orgID, id, templateID, commodity, unit, q.Quantity, StatusActive, q.DataHash, userID, code))
		if err != nil {
			return Transform{}, err
		}
		if _, err := tx.Exec(ctx, event, q.PassportID, id, EventTransform, nil, orgID, q.Quantity, q.DataHash); err != nil {
			return Transform{}, err
		}
		t.Outputs = append(t.Outputs, p)
	}

	payload := chain.Transform{TransformID: id, Stage: in.Stage, Outputs: queued, DataHash: dataHash}
	for _, u := range used {
		payload.Inputs = append(payload.Inputs, u.PassportID)
		payload.InputsUsed = append(payload.InputsUsed, u.QuantityUsed)
	}
	if err := chain.Enqueue(ctx, tx, chain.KindTransform, id, payload); err != nil {
		return Transform{}, err
	}
	return t, nil
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
	ids := make([]string, len(g.Nodes))
	for i, n := range g.Nodes {
		ids[i] = n.ID
	}
	rows, err = s.Pool.Query(ctx,
		`SELECT ti.passport_id::text, c.id::text, tr.id::text, tr.stage, ti.quantity_used
		 FROM transform_inputs ti
		 JOIN transforms tr ON tr.id = ti.transform_id
		 JOIN passports c ON c.transform_id = tr.id
		 WHERE ti.passport_id = ANY($1::uuid[]) AND c.id = ANY($1::uuid[]) AND NOT ti.is_deleted
		 ORDER BY tr.created_at, ti.passport_id, c.id`, ids)
	if err != nil {
		return Genealogy{}, err
	}
	edges, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Edge, error) {
		var e Edge
		err := row.Scan(&e.ParentID, &e.ChildID, &e.TransformID, &e.Stage, &e.QuantityUsed)
		return e, err
	})
	if err != nil {
		return Genealogy{}, err
	}
	g.Edges = append(g.Edges, edges...)
	return g, nil
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
		var holder string
		var remaining int64
		err := tx.QueryRow(ctx,
			`UPDATE passports SET status = $2 WHERE id = $1::uuid AND status <> $2 AND NOT is_deleted
			 RETURNING organisation_id::text, remaining`, id, StatusRevoked).Scan(&holder, &remaining)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return 0, err
		}
		var eventID string
		err = tx.QueryRow(ctx,
			`INSERT INTO trace_events (passport_id, ref_id, kind, from_org, quantity, data_hash)
			 VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5, $6) RETURNING id::text`,
			id, refID, EventRevoke, holder, remaining, reasonHash).Scan(&eventID)
		if err != nil {
			return 0, err
		}
		// The queued item's subject is the journey event, which receives the transaction hash.
		err = chain.Enqueue(ctx, tx, chain.KindPassportRevoke, eventID, chain.PassportReason{PassportID: id, ReasonHash: reasonHash})
		if err != nil {
			return 0, err
		}
		if err := notify.Add(ctx, tx, holder, notify.KindPassportRevoked, id, MsgNotifyPassportRevoked); err != nil {
			return 0, err
		}
		revoked++
	}
	return revoked, nil
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
