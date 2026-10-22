package passport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

func parseStages(stagesJSON string) ([]template.Stage, error) {
	var stages []template.Stage
	err := json.Unmarshal([]byte(stagesJSON), &stages)
	return stages, err
}

// liveStages returns the commodity's live template and its stages.
func liveStages(ctx context.Context, tx pgx.Tx, commodity string) (templateID string, stages []template.Stage, err error) {
	var stagesJSON string
	err = tx.QueryRow(ctx,
		`SELECT id::text, stages::text FROM commodity_templates WHERE commodity = $1 AND NOT is_deleted`, commodity,
	).Scan(&templateID, &stagesJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrTemplateNotFound
	}
	if err != nil {
		return "", nil, err
	}
	stages, err = parseStages(stagesJSON)
	return templateID, stages, err
}

// requireEvidence refuses the action entering stage st unless every piece of
// evidence it asks for is a valid certificate on the subject: issued under
// that evidence key, not revoked, and inside its validity window now.
func requireEvidence(ctx context.Context, tx pgx.Tx, st template.Stage, subjectType, subjectID string) error {
	if len(st.RequiredEvidence) == 0 {
		return nil
	}
	rows, err := tx.Query(ctx,
		`SELECT DISTINCT type FROM attestations
		 WHERE subject_type = $1 AND subject_id = $2::uuid AND type = ANY($3) AND status = $4
		   AND valid_from <= now() AND valid_to >= now() AND NOT is_deleted`,
		subjectType, subjectID, st.RequiredEvidence, attestation.StatusValid)
	if err != nil {
		return err
	}
	have, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	var missing []string
	for _, key := range st.RequiredEvidence {
		if !slices.Contains(have, key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return EvidenceError{Stage: st.Key, Missing: missing}
	}
	return nil
}

// noOpenHandover refuses while a handover of the passport awaits its receiver.
func noOpenHandover(ctx context.Context, tx pgx.Tx, passportID string) error {
	var open bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM handovers WHERE passport_id = $1::uuid AND status = $2
		                AND deadline >= now() AND NOT is_deleted)`, passportID, HandoverAwaiting).Scan(&open)
	if err == nil && open {
		return ErrOpenHandover
	}
	return err
}

// nextCode takes the organisation's next lot code for the year, in the
// caller's transaction. The counter row stays locked until it commits, so
// two passports never get the same number.
//
// The number shows how many lots the organisation made that year; accepted
// for the MVP.
func nextCode(ctx context.Context, tx pgx.Tx, orgID string, now time.Time) (string, error) {
	year := now.UTC().Year()
	var (
		slug string
		n    int
	)
	err := tx.QueryRow(ctx,
		`WITH c AS (
		   INSERT INTO lot_counters (organisation_id, year, next) VALUES ($1::uuid, $2, 2)
		   ON CONFLICT (organisation_id, year) WHERE NOT is_deleted DO UPDATE SET next = lot_counters.next + 1
		   RETURNING next - 1 AS n)
		 SELECT o.slug, c.n FROM c, organisations o WHERE o.id = $1::uuid`, orgID, year).Scan(&slug, &n)
	return fmt.Sprintf(CodeFormat, strings.ToUpper(slug), year, n), err
}

// Resolve turns a passport id or lot code into the id. Anything that is not
// an id is looked up as a code, ignoring case.
func (s Store) Resolve(ctx context.Context, idOrCode string) (string, error) {
	if httpx.IsUUID(idOrCode) {
		return strings.ToLower(idOrCode), nil
	}
	var id string
	err := s.Pool.QueryRow(ctx,
		`SELECT id::text FROM passports WHERE code = upper($1) AND NOT is_deleted`, idOrCode).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return id, err
}

// Retire closes a passport whose lot reached its final buyer or was used up.
// Only the holder may, while it is active and nothing awaits a receiver.
// After that it cannot be handed over, transformed or retired again.
func (s Store) Retire(ctx context.Context, id, orgID string) (Passport, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Passport{}, err
	}
	defer tx.Rollback(ctx)

	if err := lockHeldActive(ctx, tx, id, orgID); err != nil {
		return Passport{}, err
	}

	p, err := onePassport(tx.Query(ctx,
		`UPDATE passports AS p SET status = $2 WHERE p.id = $1::uuid RETURNING `+columns, id, StatusRetired))
	if err != nil {
		return Passport{}, err
	}
	var eventID string
	err = tx.QueryRow(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, from_org, quantity, data_hash)
		 VALUES ($1::uuid, $1::uuid, $2, $3::uuid, $4, $5) RETURNING id::text`,
		id, EventRetire, orgID, p.Remaining, p.DataHash).Scan(&eventID)
	if err != nil {
		return Passport{}, err
	}
	// The queued item's subject is the journey event, which receives the transaction hash.
	if err := chain.Enqueue(ctx, tx, chain.KindPassportRetire, eventID, chain.PassportRetire{PassportID: id}); err != nil {
		return Passport{}, err
	}
	return p, tx.Commit(ctx)
}

// lockHeldActive locks a passport the organisation holds, active and with no handover open.
func lockHeldActive(ctx context.Context, tx pgx.Tx, id, orgID string) error {
	var holder, status string
	err := tx.QueryRow(ctx,
		`SELECT organisation_id::text, status FROM passports WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, id,
	).Scan(&holder, &status)
	switch {
	case errors.Is(err, pgx.ErrNoRows) || (err == nil && holder != orgID):
		return ErrNotHolder
	case err != nil:
		return err
	case status != StatusActive:
		return ErrNotActive
	}
	return noOpenHandover(ctx, tx, id)
}
