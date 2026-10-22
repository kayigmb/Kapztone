// Package listq is the one implementation of paging, search, sorting and date
// filtering used by every list endpoint.
//
//	GET /things?page=2&size=20&search=acme&sort=-created_at&start_date=2026-01-01&end_date=2026-03-31
package listq

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// URL parameter names.
const (
	ParamPage      = "page"
	ParamSize      = "size"
	ParamSearch    = "search"
	ParamSort      = "sort"
	ParamStartDate = "start_date"
	ParamEndDate   = "end_date"
)

const (
	DefaultSize = 20
	MaxSize     = 100
	MaxPage     = 100000        // keeps (page-1)*size far from overflowing
	DateLayout  = time.DateOnly // 2026-01-31; RFC 3339 timestamps are accepted too

	// Base columns every table has.
	ColID        = "id"
	ColCreatedAt = "created_at"
	ColIsDeleted = "is_deleted"

	descPrefix = "-"
)

// ErrBadParam marks an error caused by the caller's URL; its message is safe to show.
var ErrBadParam = errors.New("invalid list parameter")

func bad(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrBadParam, fmt.Sprintf(format, a...))
}

// Config is what one list endpoint allows. Column names come from code, never
// from the request, which is what makes it safe to place them in SQL.
type Config struct {
	Alias  string   // table alias in the query, e.g. "o"
	Search []string // text columns matched by ?search=
	Sort   []string // columns allowed in ?sort=
}

// Query is a validated set of list parameters.
type Query struct {
	Page, Size int
	Search     string
	SortCol    string
	Desc       bool
	Start, End *time.Time // End is exclusive
}

// Page is the response shape of every list endpoint.
type Page[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

func Parse(v url.Values, cfg Config) (Query, error) {
	q := Query{Page: 1, Size: DefaultSize, SortCol: ColCreatedAt, Desc: true}

	if s := v.Get(ParamPage); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > MaxPage {
			return q, bad("%s must be a number from 1 to %d", ParamPage, MaxPage)
		}
		q.Page = n
	}
	if s := v.Get(ParamSize); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return q, bad("%s must be a number from 1 to %d", ParamSize, MaxSize)
		}
		q.Size = min(n, MaxSize)
	}

	q.Search = strings.TrimSpace(v.Get(ParamSearch))

	if s := v.Get(ParamSort); s != "" {
		col, desc := strings.CutPrefix(s, descPrefix)
		if !slices.Contains(cfg.Sort, col) {
			return q, bad("%s must be one of: %s (prefix with %s for descending)",
				ParamSort, strings.Join(cfg.Sort, ", "), descPrefix)
		}
		q.SortCol, q.Desc = col, desc
	}

	start, _, err := parseDate(v.Get(ParamStartDate), ParamStartDate)
	if err != nil {
		return q, err
	}
	q.Start = start
	end, dateOnly, err := parseDate(v.Get(ParamEndDate), ParamEndDate)
	if err != nil {
		return q, err
	}
	if end != nil && dateOnly {
		next := end.AddDate(0, 0, 1) // a bare end date includes that whole day
		end = &next
	}
	q.End = end
	if q.Start != nil && q.End != nil && !q.Start.Before(*q.End) {
		return q, bad("%s must be before %s", ParamStartDate, ParamEndDate)
	}
	return q, nil
}

// FromRequest parses the list parameters of a request, or writes 400.
func FromRequest(w http.ResponseWriter, r *http.Request, cfg Config) (Query, bool) {
	q, err := Parse(r.URL.Query(), cfg)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return q, false
	}
	return q, true
}

// BadParam writes the 400 for an endpoint-specific list parameter.
func BadParam(w http.ResponseWriter, param string, allowed []string) {
	httpx.Error(w, http.StatusBadRequest,
		fmt.Sprintf("%s: %s must be one of: %s", ErrBadParam, param, strings.Join(allowed, ", ")))
}

func parseDate(s, param string) (t *time.Time, dateOnly bool, err error) {
	if s == "" {
		return nil, false, nil
	}
	if d, err := time.Parse(DateLayout, s); err == nil {
		return &d, true, nil
	}
	if d, err := time.Parse(time.RFC3339, s); err == nil {
		return &d, false, nil
	}
	return nil, false, bad("%s must look like 2026-01-31", param)
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// SQL builds the list and count statements. from is the FROM clause with joins;
// where and args are the endpoint's own conditions, numbered from $1.
// Deleted rows are always excluded.
func (q Query) SQL(cfg Config, sel, from string, where []string, args []any) (list, count string, listArgs, countArgs []any) {
	col := func(name string) string { return cfg.Alias + "." + name }
	arg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	conds := append(slices.Clone(where), "NOT "+col(ColIsDeleted))
	if q.Search != "" && len(cfg.Search) > 0 {
		p := arg("%" + likeEscaper.Replace(q.Search) + "%")
		var ors []string
		for _, c := range cfg.Search {
			ors = append(ors, col(c)+"::text ILIKE "+p)
		}
		conds = append(conds, "("+strings.Join(ors, " OR ")+")")
	}
	if q.Start != nil {
		conds = append(conds, col(ColCreatedAt)+" >= "+arg(*q.Start))
	}
	if q.End != nil {
		conds = append(conds, col(ColCreatedAt)+" < "+arg(*q.End))
	}
	body := " " + from + " WHERE " + strings.Join(conds, " AND ")

	count = "SELECT count(*)" + body
	countArgs = slices.Clone(args)

	dir := "ASC"
	if q.Desc {
		dir = "DESC"
	}
	// id breaks ties so pages never overlap.
	list = sel + body +
		" ORDER BY " + col(q.SortCol) + " " + dir + ", " + col(ColID) +
		" LIMIT " + arg(q.Size) + " OFFSET " + arg((q.Page-1)*q.Size)
	return list, count, args, countArgs
}

// Querier is satisfied by a pgx pool, connection or transaction.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Fetch runs the list and count statements and returns one page.
func Fetch[T any](ctx context.Context, db Querier, q Query, cfg Config, sel, from string,
	where []string, args []any, scan pgx.RowToFunc[T]) (Page[T], error) {

	list, count, listArgs, countArgs := q.SQL(cfg, sel, from, where, args)
	page := Page[T]{Page: q.Page, Size: q.Size}

	if err := db.QueryRow(ctx, count, countArgs...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := db.Query(ctx, list, listArgs...)
	if err != nil {
		return page, err
	}
	page.Data, err = pgx.CollectRows(rows, scan)
	if page.Data == nil {
		page.Data = []T{} // JSON [] rather than null
	}
	return page, err
}
