package audit

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// Entry is one audit row.
type Entry struct {
	ID             string    `json:"id"`
	Seq            int64     `json:"seq"`
	ActorUserID    *string   `json:"actor_user_id"` // nil for public requests (sign-in, registration)
	ActorEmail     *string   `json:"actor_email"`
	OrganisationID *string   `json:"organisation_id"` // the acting organisation, nil when none
	Method         string    `json:"method"`
	Route          string    `json:"route"`
	SubjectID      *string   `json:"subject_id"` // the {id} in the URL, if any
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	PrevHash       string    `json:"prev_hash"`
	Hash           string    `json:"hash"`
	deleted        bool
}

// genesis is the previous hash of the first row.
var genesis = chain.Hex([32]byte{})

// seal is the row's hash: SHA-256 of the canonical JSON of every recorded
// field, the previous row's hash among them.
func seal(e Entry) string {
	b, _ := chain.CanonicalJSON(struct {
		Seq       int64   `json:"seq"`
		Actor     *string `json:"actor_user_id"`
		Org       *string `json:"organisation_id"`
		Method    string  `json:"method"`
		Route     string  `json:"route"`
		Subject   *string `json:"subject_id"`
		Status    int     `json:"status"`
		CreatedAt string  `json:"created_at"`
		Deleted   bool    `json:"is_deleted"`
		PrevHash  string  `json:"prev_hash"`
	}{e.Seq, e.ActorUserID, e.OrganisationID, e.Method, e.Route, e.SubjectID, e.Status,
		e.CreatedAt.UTC().Format(time.RFC3339Nano), e.deleted, e.PrevHash})
	return chain.Hex(sha256.Sum256(b))
}

var ListConfig = listq.Config{
	Alias:  "a",
	Search: []string{"route", "method"},
	Sort:   []string{"seq", "route", listq.ColCreatedAt},
}

const (
	entryColumns = `a.id::text, a.seq, a.actor_user_id::text, u.email::text, a.organisation_id::text, a.method,
	                a.route, a.subject_id, a.status, a.created_at, a.prev_hash, a.hash, a.is_deleted`
	entryFrom = `FROM audit_log a LEFT JOIN users u ON u.id = a.actor_user_id`
)

func scanEntry(row pgx.CollectableRow) (Entry, error) {
	var e Entry
	err := row.Scan(&e.ID, &e.Seq, &e.ActorUserID, &e.ActorEmail, &e.OrganisationID, &e.Method,
		&e.Route, &e.SubjectID, &e.Status, &e.CreatedAt, &e.PrevHash, &e.Hash, &e.deleted)
	return e, err
}

type Store struct {
	Pool *pgxpool.Pool
}

// Add appends e to the chain. The lock makes appends one at a time, so each
// row sees the true last hash.
func (s Store) Add(ctx context.Context, e Entry) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, chainLock); err != nil {
		return err
	}
	e.PrevHash = genesis
	err = tx.QueryRow(ctx, `SELECT seq, hash FROM audit_log ORDER BY seq DESC LIMIT 1`).Scan(&e.Seq, &e.PrevHash)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	e.Seq++
	// Postgres keeps microseconds; sealing the stored value keeps the hash checkable.
	e.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	e.Hash = seal(e)
	if _, err := tx.Exec(ctx,
		`INSERT INTO audit_log (seq, actor_user_id, organisation_id, method, route, subject_id, status, prev_hash, hash, created_at)
		 VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10)`,
		e.Seq, e.ActorUserID, e.OrganisationID, e.Method, e.Route, e.SubjectID, e.Status, e.PrevHash, e.Hash, e.CreatedAt,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// where is the list filter shared by the JSON page and the CSV download.
func where(gate, actor, route string) ([]string, []any) {
	return []string{auth.GateSQL("a.organisation_id", 1), `($2 = '' OR a.actor_user_id = NULLIF($2, '')::uuid)`, `($3 = '' OR a.route = $3)`},
		[]any{gate, actor, route}
}

func (s Store) List(ctx context.Context, q listq.Query, gate, actor, route string) (listq.Page[Entry], error) {
	w, args := where(gate, actor, route)
	return listq.Fetch(ctx, s.Pool, q, ListConfig, "SELECT "+entryColumns, entryFrom, w, args, scanEntry)
}

// Each streams every matching row, in the list's order, without paging.
func (s Store) Each(ctx context.Context, q listq.Query, gate, actor, route string, fn func(Entry) error) error {
	w, args := where(gate, actor, route)
	q.Page, q.Size = 1, 1<<31-1 // ponytail: the list SQL with an unreachable limit; a download is every row
	list, _, listArgs, _ := q.SQL(ListConfig, "SELECT "+entryColumns, entryFrom, w, args)
	return s.each(ctx, list, listArgs, fn)
}

func (s Store) each(ctx context.Context, sql string, args []any, fn func(Entry) error) error {
	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return err
		}
		if err := fn(e); err != nil {
			return err
		}
	}
	return rows.Err()
}

// Check is the result of walking the chain.
type Check struct {
	OK       bool   `json:"ok"`
	Checked  int64  `json:"checked"`
	BrokenAt *int64 `json:"broken_at"` // seq of the first row that does not match
}

// Verify recomputes every row's hash, deleted rows included, oldest first.
// A removed row shows as the next row's prev_hash not matching. Only the
// newest row can be removed unnoticed.
func (s Store) Verify(ctx context.Context) (Check, error) {
	var c Check
	prev := genesis
	errBroken := errors.New("broken")
	err := s.each(ctx, "SELECT "+entryColumns+" "+entryFrom+" ORDER BY a.seq", nil, func(e Entry) error {
		c.Checked++
		if e.PrevHash != prev || seal(e) != e.Hash {
			c.BrokenAt = &e.Seq
			return errBroken
		}
		prev = e.Hash
		return nil
	})
	if errors.Is(err, errBroken) {
		return c, nil
	}
	c.OK = err == nil
	return c, err
}

// Volume is the quantity issued at origin in one month, per commodity and
// unit, and the quantity exported that month.
type Volume struct {
	Month     string `json:"month"` // 2026-03, in UTC
	Commodity string `json:"commodity"`
	Unit      string `json:"unit"`
	Quantity  int64  `json:"quantity"` // issued at origin
	Exported  int64  `json:"exported"` // handed over, accepted, at a stage that allows EXPORT
}

// Volumes sums origin passports by month. Transform outputs are left out:
// they are made from lots already counted. Every status counts, revoked
// included: the report is what was issued.
//
// Exported is the quantity of completed handovers at a stage of the
// commodity's template that allows EXPORT, by the month they were accepted,
// gated by the sending organisation.
func (s Store) Volumes(ctx context.Context, gate string, start, end *time.Time) ([]Volume, error) {
	rows, err := s.Pool.Query(ctx,
		`WITH issued AS (
		   SELECT to_char(date_trunc('month', p.created_at AT TIME ZONE 'UTC'), 'YYYY-MM') AS month,
		          p.commodity, p.unit, sum(p.quantity)::bigint AS quantity
		   FROM passports p JOIN facilities f ON f.id = p.facility_id
		   WHERE p.transform_id IS NULL AND NOT p.is_deleted AND `+auth.GateSQL("f.organisation_id", 1)+`
		     AND ($2::timestamptz IS NULL OR p.created_at >= $2) AND ($3::timestamptz IS NULL OR p.created_at < $3)
		   GROUP BY 1, 2, 3),
		 exported AS (
		   SELECT to_char(date_trunc('month', h.responded_at AT TIME ZONE 'UTC'), 'YYYY-MM') AS month,
		          p.commodity, p.unit, sum(h.quantity)::bigint AS quantity
		   FROM handovers h
		   JOIN passports p ON p.id = h.passport_id
		   JOIN commodity_templates t ON t.id = p.template_id
		   WHERE h.status = $4 AND NOT h.is_deleted AND `+auth.GateSQL("h.from_org", 1)+`
		     AND ($2::timestamptz IS NULL OR h.responded_at >= $2) AND ($3::timestamptz IS NULL OR h.responded_at < $3)
		     AND EXISTS (SELECT 1 FROM jsonb_array_elements(t.stages) st
		                 WHERE st->>'key' = h.stage AND st->'events' ? $5)
		   GROUP BY 1, 2, 3)
		 SELECT month, commodity, unit, COALESCE(i.quantity, 0), COALESCE(x.quantity, 0)
		 FROM issued i FULL JOIN exported x USING (month, commodity, unit)
		 ORDER BY 1, 2, 3`, gate, start, end, handoverCompleted, template.EventExport)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Volume, error) {
		var v Volume
		return v, row.Scan(&v.Month, &v.Commodity, &v.Unit, &v.Quantity, &v.Exported)
	})
	if out == nil {
		out = []Volume{}
	}
	return out, err
}
