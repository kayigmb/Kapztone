package audit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
)

// execer is a pool or a transaction.
type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// RecordRule counts one action a rule refused, for the "blocked" report.
//
// A refused action's own transaction is rolled back, so pass the pool (or a
// transaction that commits), never the refused action's transaction: a
// rolled-back row counts nothing. orgID may be "" when no organisation acted.
func RecordRule(ctx context.Context, db execer, rule, orgID string) error {
	_, err := db.Exec(ctx,
		`INSERT INTO rule_events (organisation_id, rule) VALUES (NULLIF($1, '')::uuid, $2)`, orgID, rule)
	return err
}

// Blocked is how many actions one rule refused.
type Blocked struct {
	Rule  string `json:"rule"`
	Count int64  `json:"count"`
}

// Blocked counts refusals per rule between start and end, through the gate.
// Rejected handovers are not refusals of a request but a receiver's answer,
// so they are counted from the handovers themselves.
func (s Store) Blocked(ctx context.Context, gate string, start, end *time.Time) ([]Blocked, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT r.rule, count(*) FROM rule_events r
		 WHERE NOT r.is_deleted AND `+auth.GateSQL("r.organisation_id", 1)+`
		   AND ($2::timestamptz IS NULL OR r.created_at >= $2) AND ($3::timestamptz IS NULL OR r.created_at < $3)
		 GROUP BY r.rule
		 UNION ALL
		 SELECT $4, count(*) FROM handovers h
		 WHERE h.status = $5 AND NOT h.is_deleted
		   AND ($1 = '' OR h.from_org = NULLIF($1, '')::uuid OR h.to_org = NULLIF($1, '')::uuid)
		   AND ($2::timestamptz IS NULL OR h.responded_at >= $2) AND ($3::timestamptz IS NULL OR h.responded_at < $3)`,
		gate, start, end, RuleHandoverRejected, handoverRejected)
	if err != nil {
		return nil, err
	}
	counts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Blocked, error) {
		var b Blocked
		return b, row.Scan(&b.Rule, &b.Count)
	})
	if err != nil {
		return nil, err
	}
	// Every rule appears, in the declared order, zero when nothing was refused.
	byRule := map[string]int64{}
	for _, b := range counts {
		byRule[b.Rule] += b.Count
	}
	out := make([]Blocked, len(Rules))
	for i, r := range Rules {
		out[i] = Blocked{Rule: r, Count: byRule[r]}
	}
	return out, nil
}
