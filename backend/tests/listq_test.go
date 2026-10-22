package tests

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var listCfg = listq.Config{Alias: "t", Search: []string{"name", "slug"}, Sort: []string{"name", listq.ColCreatedAt}}

func params(kv ...string) url.Values {
	v := url.Values{}
	for i := 0; i < len(kv); i += 2 {
		v.Set(kv[i], kv[i+1])
	}
	return v
}

func TestListParse(t *testing.T) {
	q, err := listq.Parse(params(), listCfg)
	if err != nil || q.Page != 1 || q.Size != listq.DefaultSize || q.SortCol != listq.ColCreatedAt || !q.Desc {
		t.Fatalf("defaults wrong: %+v %v", q, err)
	}

	q, _ = listq.Parse(params(listq.ParamSize, strconv.Itoa(listq.MaxSize*10)), listCfg)
	if q.Size != listq.MaxSize {
		t.Fatalf("size not capped: %d", q.Size)
	}

	q, _ = listq.Parse(params(listq.ParamSort, "name"), listCfg)
	if q.SortCol != "name" || q.Desc {
		t.Fatalf("ascending sort wrong: %+v", q)
	}
	q, _ = listq.Parse(params(listq.ParamSort, "-name"), listCfg)
	if q.SortCol != "name" || !q.Desc {
		t.Fatalf("descending sort wrong: %+v", q)
	}

	// A bare end date includes that whole day.
	q, err = listq.Parse(params(listq.ParamStartDate, "2026-01-01", listq.ParamEndDate, "2026-01-01"), listCfg)
	if err != nil || q.End.Sub(*q.Start).Hours() != 24 {
		t.Fatalf("same-day range wrong: %+v %v", q, err)
	}

	bad := []url.Values{
		params(listq.ParamPage, "0"),
		params(listq.ParamPage, "abc"),
		params(listq.ParamPage, "999999999999999999"), // OFFSET would overflow
		params(listq.ParamSize, "-1"),
		params(listq.ParamSort, "password_hash"),            // not an allowed column
		params(listq.ParamSort, "name; DROP TABLE users--"), // injection attempt
		params(listq.ParamStartDate, "yesterday"),
		params(listq.ParamStartDate, "2026-02-01", listq.ParamEndDate, "2026-01-01"),
	}
	for _, v := range bad {
		if _, err := listq.Parse(v, listCfg); !errors.Is(err, listq.ErrBadParam) {
			t.Errorf("%v accepted, err = %v", v, err)
		}
	}
}

func TestListSQL(t *testing.T) {
	q, err := listq.Parse(params(
		listq.ParamPage, "3", listq.ParamSize, "10", listq.ParamSearch, "50%_off",
		listq.ParamSort, "-name", listq.ParamStartDate, "2026-01-01",
	), listCfg)
	if err != nil {
		t.Fatal(err)
	}
	list, count, listArgs, countArgs := q.SQL(listCfg, "SELECT t.id", "FROM things t", []string{"t.owner = $1"}, []any{"me"})

	for _, want := range []string{
		"t.owner = $1",
		"NOT t." + listq.ColIsDeleted,
		"(t.name::text ILIKE $2 OR t.slug::text ILIKE $2)",
		"t." + listq.ColCreatedAt + " >= $3",
		"ORDER BY t.name DESC, t." + listq.ColID,
		"LIMIT $4 OFFSET $5",
	} {
		if !strings.Contains(list, want) {
			t.Errorf("list SQL missing %q:\n%s", want, list)
		}
	}
	if strings.Contains(count, "LIMIT") || strings.Contains(count, "ORDER BY") || !strings.HasPrefix(count, "SELECT count(*)") {
		t.Errorf("count SQL wrong:\n%s", count)
	}
	// The search text travels as an argument with its wildcards escaped, never inside the SQL.
	if strings.Contains(list, "50") || listArgs[1] != `%50\%\_off%` {
		t.Errorf("search not parameterised/escaped: %v", listArgs)
	}
	if len(listArgs) != 5 || listArgs[3] != 10 || listArgs[4] != 20 || len(countArgs) != 3 {
		t.Errorf("args wrong: list %v count %v", listArgs, countArgs)
	}
}

// Request text never becomes SQL: the sort column comes from the whitelist and
// the search travels as a bound argument.
func TestListSQLInjection(t *testing.T) {
	if _, err := listq.Parse(params(listq.ParamSort, "name;DROP TABLE users"), listCfg); !errors.Is(err, listq.ErrBadParam) {
		t.Fatalf("sort outside the whitelist accepted: %v", err)
	}
	attack := "'; DROP TABLE users;--"
	q, err := listq.Parse(params(listq.ParamSearch, attack), listCfg)
	if err != nil {
		t.Fatal(err)
	}
	list, count, args, _ := q.SQL(listCfg, "SELECT t.id", "FROM things t", nil, nil)
	if strings.Contains(list, attack) || strings.Contains(count, attack) {
		t.Fatalf("search text inside the SQL: %s", list)
	}
	if args[0] != "%"+attack+"%" {
		t.Fatalf("search not bound as an argument: %v", args)
	}
}
