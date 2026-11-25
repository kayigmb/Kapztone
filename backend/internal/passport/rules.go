package passport

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// queueSplitYield queues the split stage's yield of 1 for a commodity the
// first time it is seen. The contract refuses a transform at a stage whose
// yield it has not been told, and every transform runs at the split stage.
func queueSplitYield(ctx context.Context, tx pgx.Tx, commodity string) error {
	// Two first issues of one commodity at once would both queue it; the
	// second is harmless (it sets the same yield again).
	var queued bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM chain_outbox WHERE kind = $1
		                AND payload->>'commodity' = $2 AND payload->>'stage' = $3)`,
		chain.KindMaxYieldSet, commodity, StageSplit).Scan(&queued)
	if err != nil || queued {
		return err
	}
	id, err := newID(ctx, tx) // the outbox wants an id; a yield has none of its own
	if err != nil {
		return err
	}
	return chain.Enqueue(ctx, tx, chain.KindMaxYieldSet, id,
		chain.MaxYieldSet{Commodity: commodity, Stage: StageSplit, PartsPerMillion: chain.YieldScale})
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
	var slug string
	if err := tx.QueryRow(ctx, `SELECT slug FROM organisations WHERE id = $1::uuid`, orgID).Scan(&slug); err != nil {
		return "", err
	}
	// Counted per prefix, not per organisation: two that share one never collide.
	var n int
	err := tx.QueryRow(ctx,
		`INSERT INTO code_counters (prefix, year, next) VALUES ($1, $2, 2)
		 ON CONFLICT (prefix, year) WHERE NOT is_deleted DO UPDATE SET next = code_counters.next + 1
		 RETURNING next - 1`, ShortCode(slug), now.UTC().Year()).Scan(&n)
	return LotCode(slug, now, n), err
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
