package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/server"
)

// Runs the real SQL against a migrated Postgres. Start one with `make infra`, then:
//
//	TEST_DATABASE_URL=postgres://chaintrack:chaintrack@localhost:5432/chaintrack?sslmode=disable go test ./tests/
func TestAuthAgainstDatabase(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close) // registered first, so it runs after the row cleanup below

	// Unique names so reruns and parallel runs never collide.
	suffix := fmt.Sprint(time.Now().UnixNano())
	slug, rival := "t-acme-"+suffix, "t-rival-"+suffix
	email := "ada-" + suffix + "@acme.test"
	hash, _ := auth.HashPassword("correct-password")

	var userID string
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO organisations (name, slug, kind) VALUES ('Acme', $1, 'producer'), ('Rival', $2, 'exporter')`, slug, rival)
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, 'Ada') RETURNING id::text`,
		email, hash).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1::uuid`, userID)
		pool.Exec(ctx, `DELETE FROM organisations WHERE slug IN ($1, $2)`, slug, rival)
	})
	exec(`INSERT INTO userscope (user_id, role_id, organisation_id)
	      SELECT $1::uuid, r.id, o.id FROM roles r, organisations o WHERE r.name = 'member' AND o.slug = $2`, userID, slug)

	a := &auth.Auth{
		Store:    auth.PgStore{Pool: pool},
		Secret:   []byte("test-secret-test-secret-test-secret"),
		TokenTTL: time.Hour,
		Dev:      true,
	}
	h := server.NewRouter(pool.Ping, a)

	// Sign in; the email check is case-insensitive.
	login := do(h, http.MethodPost, "/auth/login", "",
		fmt.Sprintf(`{"email":"ADA-%s@acme.test","password":"correct-password"}`, suffix))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body %s", login.Code, login.Body)
	}
	var out struct{ Token string }
	json.Unmarshal(login.Body.Bytes(), &out)

	me := do(h, http.MethodGet, "/me?pslug="+slug, out.Token, "")
	if me.Code != http.StatusOK {
		t.Fatalf("/me status = %d, body %s", me.Code, me.Body)
	}
	var got struct {
		Organisation struct{ Slug string }
		Role         string
		Permissions  []string
	}
	json.Unmarshal(me.Body.Bytes(), &got)
	if got.Organisation.Slug != slug || got.Role != "member" {
		t.Fatalf("unexpected scope: %s", me.Body)
	}
	if !slices.Contains(got.Permissions, "passport:issue") || slices.Contains(got.Permissions, "org:verify") {
		t.Fatalf("member permissions wrong: %v", got.Permissions)
	}

	if rec := do(h, http.MethodGet, "/me?pslug="+rival, out.Token, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("other organisation status = %d, want 403", rec.Code)
	}

	// Removal is instant: same token, next request refused.
	exec(`DELETE FROM userscope WHERE user_id = $1::uuid`, userID)
	if rec := do(h, http.MethodGet, "/me?pslug="+slug, out.Token, ""); rec.Code != http.StatusForbidden {
		t.Fatalf("after removal status = %d, want 403", rec.Code)
	}

	exec(`UPDATE users SET status = 'disabled' WHERE id = $1::uuid`, userID)
	if rec := do(h, http.MethodGet, "/me", out.Token, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("disabled user status = %d, want 401", rec.Code)
	}
}
