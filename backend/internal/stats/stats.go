// Package stats answers the platform dashboard: how many of everything there
// are, what happened each day of the last month, and whether the chain is
// keeping up. Platform only; nothing here is per organisation.
package stats

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
)

// Days is how far back the daily series go.
const Days = 30

// TopOrganisations is how many organisations the volume ranking lists.
const TopOrganisations = 5

// MsgPlatformOnly is the refusal for an organisation's user.
const MsgPlatformOnly = "the platform dashboard is for the platform only"

// Count is one bucket of a breakdown (a status, a kind, a commodity…).
type Count struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Day is one day of the activity series.
type Day struct {
	Date       string `json:"date"` // 2026-10-19
	Passports  int64  `json:"passports"`
	Handovers  int64  `json:"handovers"`
	Transforms int64  `json:"transforms"`
	Flags      int64  `json:"flags"`
	Signups    int64  `json:"signups"`
}

// Volume is what was issued at origin for one commodity.
type Volume struct {
	Commodity string `json:"commodity"`
	Unit      string `json:"unit"`
	Quantity  int64  `json:"quantity"`
	Lots      int64  `json:"lots"`
}

// TopOrg is an organisation ranked by what it issued at origin.
type TopOrg struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
	Lots     int64  `json:"lots"`
}

// Outbox is how far the relayer is behind.
type Outbox struct {
	Pending          int64    `json:"pending"`
	Failed           int64    `json:"failed"`
	OldestPendingSec *float64 `json:"oldest_pending_seconds"` // nil when nothing waits
}

type Stats struct {
	Organisations []Count   `json:"organisations"` // by status
	OrgKinds      []Count   `json:"organisation_kinds"`
	Users         int64     `json:"users"`
	Facilities    []Count   `json:"facilities"`   // by status
	Passports     []Count   `json:"passports"`    // by status
	Certificates  []Count   `json:"certificates"` // by status as read (expired included)
	Flags         []Count   `json:"flags"`        // by status
	Activity      []Day     `json:"activity"`     // the last Days days, oldest first
	Volume        []Volume  `json:"volume"`
	TopOrgs       []TopOrg  `json:"top_organisations"`
	Outbox        Outbox    `json:"outbox"`
	GeneratedAt   time.Time `json:"generated_at"`
}

type Store struct {
	Pool *pgxpool.Pool
}

// Get gathers the dashboard. Each block is one statement; they read one snapshot.
func (s Store) Get(ctx context.Context) (Stats, error) {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Stats{}, err
	}
	defer tx.Rollback(ctx)

	st := Stats{GeneratedAt: time.Now().UTC()}
	counts := []struct {
		dst  *[]Count
		sql  string
		args []any
	}{
		{&st.Organisations, `SELECT status, count(*) FROM organisations WHERE NOT is_deleted GROUP BY 1 ORDER BY 2 DESC`, nil},
		{&st.OrgKinds, `SELECT kind, count(*) FROM organisations WHERE NOT is_deleted GROUP BY 1 ORDER BY 2 DESC`, nil},
		{&st.Facilities, `SELECT status, count(*) FROM facilities WHERE NOT is_deleted GROUP BY 1 ORDER BY 2 DESC`, nil},
		{&st.Passports, `SELECT status, count(*) FROM passports WHERE NOT is_deleted GROUP BY 1 ORDER BY 2 DESC`, nil},
		{&st.Certificates, `SELECT CASE WHEN status = $1 AND valid_to < now() THEN $2 ELSE status END, count(*)
		                    FROM attestations WHERE NOT is_deleted GROUP BY 1 ORDER BY 2 DESC`,
			[]any{attestation.StatusValid, attestation.StatusExpired}},
		{&st.Flags, `SELECT status, count(*) FROM flags WHERE NOT is_deleted GROUP BY 1 ORDER BY 2 DESC`, nil},
	}
	for _, c := range counts {
		if *c.dst, err = collect(ctx, tx, c.sql, c.args, pgx.RowToStructByPos[Count]); err != nil {
			return Stats{}, err
		}
	}
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE NOT is_deleted`).Scan(&st.Users); err != nil {
		return Stats{}, err
	}

	// One row per day; each column counts what started (or, for handovers, finished) that day, in UTC.
	st.Activity, err = collect(ctx, tx, `
		WITH days AS (SELECT generate_series(current_date - ($1::int - 1), current_date, interval '1 day')::date AS d)
		SELECT to_char(d, 'YYYY-MM-DD'),
		  (SELECT count(*) FROM passports p WHERE NOT p.is_deleted AND p.facility_id IS NOT NULL AND p.created_at::date = d),
		  (SELECT count(*) FROM handovers h WHERE NOT h.is_deleted AND h.status = $2 AND h.updated_at::date = d),
		  (SELECT count(*) FROM transforms t WHERE NOT t.is_deleted AND t.created_at::date = d),
		  (SELECT count(*) FROM flags f WHERE NOT f.is_deleted AND f.created_at::date = d),
		  (SELECT count(*) FROM organisations o WHERE NOT o.is_deleted AND o.created_at::date = d)
		FROM days ORDER BY d`, []any{Days, passport.HandoverCompleted}, pgx.RowToStructByPos[Day])
	if err != nil {
		return Stats{}, err
	}

	// Origin passports only: what transforms make is counted at its origin already.
	st.Volume, err = collect(ctx, tx, `
		SELECT p.commodity, p.unit, COALESCE(sum(p.quantity), 0)::bigint, count(*)
		FROM passports p
		WHERE NOT p.is_deleted AND p.facility_id IS NOT NULL
		GROUP BY 1, 2 ORDER BY 3 DESC`, nil, pgx.RowToStructByPos[Volume])
	if err != nil {
		return Stats{}, err
	}
	st.TopOrgs, err = collect(ctx, tx, `
		SELECT o.id::text, o.name, COALESCE(sum(p.quantity), 0)::bigint, count(*)
		FROM passports p JOIN facilities f ON f.id = p.facility_id JOIN organisations o ON o.id = f.organisation_id
		WHERE NOT p.is_deleted
		GROUP BY o.id, o.name ORDER BY 3 DESC LIMIT $1`, []any{TopOrganisations}, pgx.RowToStructByPos[TopOrg])
	if err != nil {
		return Stats{}, err
	}

	err = tx.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE status = $1), count(*) FILTER (WHERE status = $2),
		       extract(epoch FROM now() - min(created_at) FILTER (WHERE status = $1))::float8
		FROM chain_outbox WHERE NOT is_deleted`, chain.StatusPending, chain.StatusFailed,
	).Scan(&st.Outbox.Pending, &st.Outbox.Failed, &st.Outbox.OldestPendingSec)
	if err != nil {
		return Stats{}, err
	}
	return st, tx.Commit(ctx)
}

func collect[T any](ctx context.Context, tx pgx.Tx, sql string, args []any, scan pgx.RowToFunc[T]) ([]T, error) {
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(rows, scan)
	if out == nil {
		out = []T{}
	}
	return out, err
}

type Handlers struct {
	Store Store
}

// Get is GET /stats: the platform's dashboard.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	if gate, ok := auth.Gate(w, r); !ok {
		return
	} else if gate != "" {
		httpx.Error(w, http.StatusForbidden, MsgPlatformOnly)
		return
	}
	st, err := h.Store.Get(r.Context())
	if err != nil {
		httpx.ServerError(w, "platform stats", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", st)
}
