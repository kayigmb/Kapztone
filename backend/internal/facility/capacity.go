package facility

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// PeriodFor returns the capacity period a date falls in, as one number:
// 2026 for a year, 202610 for a month, 20261 or 20262 for a season.
func PeriodFor(capacityPeriod string, t time.Time) int64 {
	t = t.UTC()
	year := int64(t.Year())
	switch capacityPeriod {
	case PeriodMonth:
		return year*100 + int64(t.Month())
	case PeriodSeason:
		if t.Month() <= firstSeasonLastMonth {
			return year*10 + 1
		}
		return year*10 + 2
	}
	return year
}

// querier is a pool or a transaction.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// CapacityUsed is how much each facility has originated in the given period:
// the sum of its origin passports. It is the figure issuing checks against
// capacity, and the one the contract counts.
func CapacityUsed(ctx context.Context, db querier, ids []string, periods []int64) (map[string]int64, error) {
	rows, err := db.Query(ctx,
		`SELECT t.id::text, COALESCE(sum(p.quantity), 0)::bigint
		 FROM unnest($1::uuid[], $2::bigint[]) AS t(id, period)
		 LEFT JOIN passports p ON p.facility_id = t.id AND p.period = t.period AND NOT p.is_deleted
		 GROUP BY t.id`, ids, periods)
	if err != nil {
		return nil, err
	}
	used := map[string]int64{}
	var (
		id string
		n  int64
	)
	_, err = pgx.ForEachRow(rows, []any{&id, &n}, func() error {
		used[id] = n
		return nil
	})
	return used, err
}

// withUsed fills in each facility's capacity used in the period now falls in.
func (s Store) withUsed(ctx context.Context, fs []Facility) error {
	if len(fs) == 0 {
		return nil
	}
	now := time.Now()
	ids, periods := make([]string, len(fs)), make([]int64, len(fs))
	for i, f := range fs {
		ids[i], periods[i] = f.ID, PeriodFor(f.CapacityPeriod, now)
	}
	used, err := CapacityUsed(ctx, s.Pool, ids, periods)
	if err != nil {
		return err
	}
	for i := range fs {
		fs[i].CapacityUsed = used[fs[i].ID]
	}
	return nil
}
