package tests

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/server"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// These tests run the real SQL against a migrated Postgres, and the cache
// against a real Redis. `make test-db` starts both and sets the variables:
//
//	TEST_DATABASE_URL=postgres://chaintrack:chaintrack@localhost:5478/chaintrack?sslmode=disable
//	TEST_REDIS_URL=redis://localhost:6379/1
const (
	envTestDB    = "TEST_DATABASE_URL"
	envTestRedis = "TEST_REDIS_URL"
)

// Names of the organisations and users each test creates.
const (
	tAcme  = "acme"
	tRival = "rival"
	tStaff = "staff"
	tRoot  = "root"
)

var pdfBytes = []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")
var exeBytes = []byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xff\xff\x00\x00 this is not a document")

// env is one test's isolated world: its own organisations and users, removed afterwards.
type env struct {
	t      *testing.T
	pool   *pgxpool.Pool
	a      *auth.Auth
	h      http.Handler
	suffix string
	// passports is kept so tests can set its clock and signing domain.
	passports *passport.Handlers
}

// countingStore counts lookups that reach the database.
type countingStore struct {
	auth.Store
	lookups atomic.Int64
}

func (c *countingStore) ActiveUser(ctx context.Context, id string) (auth.User, error) {
	c.lookups.Add(1)
	return c.Store.ActiveUser(ctx, id)
}

func (c *countingStore) Scope(ctx context.Context, userID, slug string) (auth.Scope, error) {
	c.lookups.Add(1)
	return c.Store.Scope(ctx, userID, slug)
}

// newEnv wires the real router. With useCache, lookups go through Redis
// exactly as in production, and the test is skipped if Redis is not configured.
func newEnv(t *testing.T, useCache bool) (*env, *countingStore, *auth.CachedStore) {
	t.Helper()
	dbURL := os.Getenv(envTestDB)
	if dbURL == "" {
		t.Skip(envTestDB + " not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close) // registered first, so it runs after the row cleanup below
	if err := auth.SyncRoles(ctx, pool); err != nil {
		t.Fatal(err)
	}

	e := &env{t: t, pool: pool, suffix: fmt.Sprint(time.Now().UnixNano())}
	t.Cleanup(func() {
		like := "%" + e.suffix + "%"
		// Everything a test created hangs off its organisations.
		const orgs = `(SELECT id FROM organisations WHERE slug LIKE $1)`
		for _, sql := range []string{
			`DELETE FROM chain_outbox WHERE subject_id IN ` + orgs + `
			    OR subject_id IN (SELECT id FROM facilities WHERE organisation_id IN ` + orgs + `)
			    OR subject_id IN (SELECT id FROM handovers WHERE from_org IN ` + orgs + `)
			    OR subject_id IN (SELECT p.id FROM passports p JOIN facilities f ON f.id = p.facility_id WHERE f.organisation_id IN ` + orgs + `)`,
			`DELETE FROM trace_events WHERE to_org IN ` + orgs + ` OR from_org IN ` + orgs,
			`DELETE FROM handovers WHERE from_org IN ` + orgs,
			`DELETE FROM passports WHERE facility_id IN (SELECT id FROM facilities WHERE organisation_id IN ` + orgs + `)`,
			`DELETE FROM documents WHERE organisation_id IN ` + orgs,
			`DELETE FROM facilities WHERE organisation_id IN ` + orgs,
			`DELETE FROM commodity_templates WHERE name LIKE $1`,
			`DELETE FROM userscope WHERE user_id IN (SELECT id FROM users WHERE email::text LIKE $1)`,
			`DELETE FROM organisations WHERE slug LIKE $1`,
			`DELETE FROM users WHERE email::text LIKE $1`,
		} {
			if _, err := pool.Exec(ctx, sql, like); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})

	counting := &countingStore{Store: auth.PgStore{Pool: pool}}
	var store auth.Store = counting
	invalidate := func(context.Context, string) {}
	var cached *auth.CachedStore
	if useCache {
		redisURL := os.Getenv(envTestRedis)
		if redisURL == "" {
			t.Skip(envTestRedis + " not set")
		}
		c, err := cache.New(redisURL, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		cached = &auth.CachedStore{Store: counting, Cache: c}
		store, invalidate = cached, cached.InvalidateOrg
	}

	// The signing domain is filled in by tests that deploy a chain (see newChain).
	e.passports = &passport.Handlers{Store: passport.Store{Pool: pool, Secrets: testBox(t)}}
	e.a = &auth.Auth{Store: store, Secret: []byte(testSecret), TokenTTL: time.Hour, Dev: true, Secrets: testBox(t)}
	e.h = server.NewRouter(server.Deps{
		Ping: pool.Ping,
		Auth: e.a,
		Org: &org.Handlers{
			Store: org.Store{Pool: pool}, UploadDir: t.TempDir(), InvalidateOrg: invalidate, Secrets: testBox(t),
		},
		Facility: &facility.Handlers{Store: facility.Store{Pool: pool}},
		Template: &template.Handlers{Store: template.Store{Pool: pool}},
		Passport: e.passports,
	})
	return e, counting, cached
}

func (e *env) exec(sql string, args ...any) {
	e.t.Helper()
	if _, err := e.pool.Exec(context.Background(), sql, args...); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) slug(name string) string  { return "t-" + name + "-" + e.suffix }
func (e *env) email(name string) string { return name + "-" + e.suffix + "@chaintrack.test" }

// register signs a business up through the API and returns it.
func (e *env) register(name, kind string) org.Organisation {
	e.t.Helper()
	body, _ := json.Marshal(org.Registration{
		Name: name, Slug: e.slug(name), Kind: kind,
		UserName: name + " admin", Email: e.email(name), Password: testPassword,
	})
	rec := do(e.h, http.MethodPost, "/orgs", "", string(body))
	if rec.Code != http.StatusCreated {
		e.t.Fatalf("register %s = %d %s", name, rec.Code, rec.Body)
	}
	return decodeTo[org.Organisation](e.t, rec)
}

// addUser inserts a user directly with the given role; orgID "" means platform-wide.
func (e *env) addUser(name, role, orgID string) (userID string) {
	e.t.Helper()
	hash, _ := auth.HashPassword(testPassword)
	err := e.pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, name, status) VALUES ($1, $2, $3, $4) RETURNING id::text`,
		e.email(name), hash, name, auth.UserActive).Scan(&userID)
	if err != nil {
		e.t.Fatal(err)
	}
	e.exec(`INSERT INTO userscope (user_id, role_id, organisation_id)
	        SELECT $1::uuid, id, NULLIF($2, '')::uuid FROM roles WHERE name = $3`, userID, orgID, role)
	return userID
}

func (e *env) login(name string) string {
	e.t.Helper()
	rec := do(e.h, http.MethodPost, "/auth/login", "", loginBody(e.email(name), testPassword))
	if rec.Code != http.StatusOK {
		e.t.Fatalf("login %s = %d %s", name, rec.Code, rec.Body)
	}
	return decodeTo[struct{ Token string }](e.t, rec).Token
}

// as returns path with the acting organisation attached.
func as(slug, path string) string {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + auth.ParamOrgSlug + "=" + slug
}

func (e *env) upload(token, slug, orgID, docType string, content []byte) *httptest.ResponseRecorder {
	e.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	mw.WriteField(org.FormType, docType)
	fw, _ := mw.CreateFormFile(org.FormFile, "../../licence.pdf") // hostile name: must be reduced to its base
	fw.Write(content)
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, as(slug, "/orgs/"+orgID+"/documents"), &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return send(e.h, req, token)
}

func decodeTo[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return v
}

func want(t *testing.T, what string, rec *httptest.ResponseRecorder, code int) {
	t.Helper()
	if rec.Code != code {
		t.Fatalf("%s = %d, want %d (%s)", what, rec.Code, code, rec.Body)
	}
}

type meResponse struct {
	Organisation *auth.Organisation
	Role         string
	Permissions  []string
}

func TestConventionsInDatabase(t *testing.T) {
	e, _, _ := newEnv(t, false)
	ctx := context.Background()

	// Every table carries the four base columns, with a uuid id.
	rows, err := e.pool.Query(ctx,
		`SELECT t.table_name FROM information_schema.tables t
		 WHERE t.table_schema = 'public' AND t.table_type = 'BASE TABLE'
		   AND t.table_name NOT IN ('schema_migrations', 'spatial_ref_sys')
		   AND (SELECT count(*) FROM information_schema.columns c
		        WHERE c.table_schema = t.table_schema AND c.table_name = t.table_name
		          AND (c.column_name, c.data_type) IN (('id', 'uuid'), ('created_at', 'timestamp with time zone'),
		               ('updated_at', 'timestamp with time zone'), ('is_deleted', 'boolean'))) <> 4`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		t.Errorf("table %s is missing a base column", name)
	}

	// updated_at moves on update without the application setting it.
	acme := e.register(tAcme, org.KindProducer)
	var moved bool
	err = e.pool.QueryRow(ctx,
		`UPDATE organisations SET name = name || '!' WHERE id = $1::uuid RETURNING updated_at > created_at`,
		acme.ID).Scan(&moved)
	if err != nil || !moved {
		t.Fatalf("updated_at did not move: %v %v", moved, err)
	}

	// The database holds exactly the permissions the code declares.
	var names []string
	err = e.pool.QueryRow(ctx,
		`SELECT array_agg(name ORDER BY name) FROM permissions WHERE NOT is_deleted`).Scan(&names)
	wantNames := slices.Sorted(slices.Values(auth.AllPermissions))
	if err != nil || !slices.Equal(names, wantNames) {
		t.Fatalf("permissions in database %v, in code %v (%v)", names, wantNames, err)
	}
}

func TestAccessAgainstDatabase(t *testing.T) {
	// No cache here: this test changes rows behind the API's back to prove the
	// SQL itself ignores deleted rows. TestCache covers the cached path.
	e, _, _ := newEnv(t, false)
	acme := e.register(tAcme, org.KindProducer)
	e.register(tRival, org.KindExporter)
	memberID := e.addUser(tStaff, auth.RoleMember, acme.ID)
	slug := e.slug(tAcme)

	// The email check is case-insensitive.
	rec := do(e.h, http.MethodPost, "/auth/login", "", loginBody(strings.ToUpper(e.email(tStaff)), testPassword))
	want(t, "login", rec, http.StatusOK)
	token := decodeTo[struct{ Token string }](t, rec).Token

	me := do(e.h, http.MethodGet, as(slug, "/me"), token, "")
	want(t, "/me", me, http.StatusOK)
	got := decodeTo[meResponse](t, me)
	if got.Organisation == nil || got.Organisation.Slug != slug || got.Role != auth.RoleMember {
		t.Fatalf("unexpected scope: %s", me.Body)
	}
	if !slices.Equal(got.Permissions, slices.Sorted(slices.Values(auth.DefaultRoles[auth.RoleMember]))) {
		t.Fatalf("member permissions = %v", got.Permissions)
	}
	// Another organisation's address gives no organisation and no permissions.
	other := do(e.h, http.MethodGet, as(e.slug(tRival), "/me"), token, "")
	want(t, "other organisation /me", other, http.StatusOK)
	if got := decodeTo[meResponse](t, other); got.Organisation != nil || len(got.Permissions) != 0 {
		t.Fatalf("scope leaked into another organisation: %s", other.Body)
	}
	// A route the member's role allows: the probe for whether access is live.
	allowed := as(slug, "/facilities")
	want(t, "other organisation data", do(e.h, http.MethodGet, as(e.slug(tRival), "/facilities"), token, ""), http.StatusForbidden)
	want(t, "own organisation data", do(e.h, http.MethodGet, allowed, token, ""), http.StatusOK)

	// Each soft delete must cut access on the very next request, and restoring must return it.
	type cut struct {
		name, table, id string
		code            int
	}
	var scopeID string
	e.pool.QueryRow(context.Background(), `SELECT id::text FROM userscope WHERE user_id = $1::uuid`, memberID).Scan(&scopeID)
	for _, c := range []cut{
		{"deleted scope", "userscope", scopeID, http.StatusForbidden},
		{"deleted organisation", "organisations", acme.ID, http.StatusForbidden},
		{"deleted user", "users", memberID, http.StatusUnauthorized},
	} {
		e.exec(`UPDATE `+c.table+` SET is_deleted = true WHERE id = $1::uuid`, c.id)
		want(t, c.name, do(e.h, http.MethodGet, allowed, token, ""), c.code)
		e.exec(`UPDATE `+c.table+` SET is_deleted = false WHERE id = $1::uuid`, c.id)
		want(t, c.name+" restored", do(e.h, http.MethodGet, allowed, token, ""), http.StatusOK)
	}

	e.exec(`UPDATE users SET status = $2 WHERE id = $1::uuid`, memberID, auth.UserDisabled)
	want(t, "disabled user", do(e.h, http.MethodGet, as(slug, "/me"), token, ""), http.StatusUnauthorized)

	// A deleted account frees its email for a new registration.
	e.exec(`UPDATE users SET is_deleted = true WHERE id = $1::uuid`, memberID)
	e.addUser(tStaff, auth.RoleMember, acme.ID)
}

func TestKYBFlow(t *testing.T) {
	e, _, _ := newEnv(t, true)

	acme := e.register(tAcme, org.KindProducer)
	rival := e.register(tRival, org.KindExporter)
	if acme.Status != org.StatusPending {
		t.Fatalf("new organisation status = %s", acme.Status)
	}
	acmeSlug, rivalSlug := e.slug(tAcme), e.slug(tRival)

	// Registration refuses a taken address or email.
	dup := org.Registration{Name: "Copy", Slug: acmeSlug, Kind: org.KindProducer,
		UserName: "x", Email: e.email("copy"), Password: testPassword}
	body, _ := json.Marshal(dup)
	rec := do(e.h, http.MethodPost, "/orgs", "", string(body))
	want(t, "duplicate slug", rec, http.StatusConflict)
	dup.Slug, dup.Email = e.slug("copy"), e.email(tAcme)
	body, _ = json.Marshal(dup)
	want(t, "duplicate email", do(e.h, http.MethodPost, "/orgs", "", string(body)), http.StatusConflict)
	var leftovers int
	e.pool.QueryRow(context.Background(), `SELECT count(*) FROM organisations WHERE slug = $1`, e.slug("copy")).Scan(&leftovers)
	if leftovers != 0 {
		t.Fatal("failed registration left an organisation behind")
	}

	e.addUser(tStaff, auth.RoleMember, acme.ID)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	admin, rivalAdmin, member, root := e.login(tAcme), e.login(tRival), e.login(tStaff), e.login(tRoot)

	// Data gate: an organisation's admin sees their own organisation and nothing else.
	list := do(e.h, http.MethodGet, as(acmeSlug, "/orgs"), admin, "")
	want(t, "own list", list, http.StatusOK)
	page := decodeTo[listq.Page[org.Organisation]](t, list)
	if page.Total != 1 || len(page.Data) != 1 || page.Data[0].ID != acme.ID {
		t.Fatalf("org admin sees %d organisations: %s", page.Total, list.Body)
	}
	want(t, "other organisation by id", do(e.h, http.MethodGet, as(acmeSlug, "/orgs/"+rival.ID), admin, ""), http.StatusNotFound)
	want(t, "own organisation by id", do(e.h, http.MethodGet, as(acmeSlug, "/orgs/"+acme.ID), admin, ""), http.StatusOK)
	want(t, "no organisation named", do(e.h, http.MethodGet, "/orgs", admin, ""), http.StatusForbidden)
	want(t, "malformed id", do(e.h, http.MethodGet, as(acmeSlug, "/orgs/not-a-uuid"), admin, ""), http.StatusNotFound)

	// Upload: content decides the type, permissions decide who.
	up := e.upload(admin, acmeSlug, acme.ID, org.DocLicence, pdfBytes)
	want(t, "upload pdf", up, http.StatusCreated)
	doc := decodeTo[org.Document](t, up)
	sum := sha256.Sum256(pdfBytes)
	if doc.SHA256 != hex.EncodeToString(sum[:]) || doc.FileName != "licence.pdf" ||
		doc.ContentType != org.AllowedContentTypes[0] || doc.SizeBytes != int64(len(pdfBytes)) {
		t.Fatalf("document recorded wrongly: %+v", doc)
	}
	want(t, "executable named .pdf", e.upload(admin, acmeSlug, acme.ID, org.DocLicence, exeBytes), http.StatusBadRequest)
	want(t, "unknown document type", e.upload(admin, acmeSlug, acme.ID, "selfie", pdfBytes), http.StatusBadRequest)
	want(t, "too large", e.upload(admin, acmeSlug, acme.ID, org.DocLicence,
		append(slices.Clone(pdfBytes), make([]byte, org.MaxUploadBytes)...)), http.StatusRequestEntityTooLarge)
	want(t, "member uploads", e.upload(member, acmeSlug, acme.ID, org.DocLicence, pdfBytes), http.StatusForbidden)
	want(t, "member lists documents", do(e.h, http.MethodGet, as(acmeSlug, "/orgs/"+acme.ID+"/documents"), member, ""), http.StatusForbidden)
	want(t, "upload to another organisation", e.upload(admin, acmeSlug, rival.ID, org.DocLicence, pdfBytes), http.StatusNotFound)

	// Download: same bytes for those allowed, invisible to another organisation.
	file := "/documents/" + doc.ID + "/file"
	dl := do(e.h, http.MethodGet, file, root, "")
	want(t, "superadmin download", dl, http.StatusOK)
	if !bytes.Equal(dl.Body.Bytes(), pdfBytes) {
		t.Fatal("downloaded bytes differ from uploaded bytes")
	}
	want(t, "other organisation download", do(e.h, http.MethodGet, as(rivalSlug, file), rivalAdmin, ""), http.StatusNotFound)

	// Superadmin sees every organisation; list parameters work together.
	all := do(e.h, http.MethodGet, fmt.Sprintf("/orgs?%s=%s&%s=%s&%s=name&%s=1",
		listq.ParamSearch, e.suffix, org.ParamStatus, org.StatusPending, listq.ParamSort, listq.ParamSize), root, "")
	want(t, "superadmin list", all, http.StatusOK)
	page = decodeTo[listq.Page[org.Organisation]](t, all)
	if page.Total != 2 || len(page.Data) != 1 || page.Data[0].ID != acme.ID || page.Size != 1 {
		t.Fatalf("superadmin list wrong: %s", all.Body)
	}
	want(t, "unsortable column", do(e.h, http.MethodGet, "/orgs?"+listq.ParamSort+"=password_hash", root, ""), http.StatusBadRequest)
	want(t, "unknown status filter", do(e.h, http.MethodGet, "/orgs?"+org.ParamStatus+"=maybe", root, ""), http.StatusBadRequest)

	// Review rules.
	verify := func(token, slug, id, decision, reason string) *httptest.ResponseRecorder {
		b, _ := json.Marshal(org.Review{Decision: decision, Reason: reason})
		path := "/orgs/" + id + "/verify"
		if slug != "" {
			path = as(slug, path)
		}
		return do(e.h, http.MethodPost, path, token, string(b))
	}
	status := func() string {
		me := do(e.h, http.MethodGet, as(acmeSlug, "/me"), admin, "")
		return decodeTo[meResponse](t, me).Organisation.Status
	}
	want(t, "org admin verifies own organisation", verify(admin, acmeSlug, acme.ID, org.StatusVerified, ""), http.StatusForbidden)
	want(t, "verify with no documents", verify(root, "", rival.ID, org.StatusVerified, ""), http.StatusConflict)
	want(t, "reject without reason", verify(root, "", acme.ID, org.StatusRejected, ""), http.StatusBadRequest)

	// Each status change is visible to the organisation's users on their next request.
	want(t, "reject", verify(root, "", acme.ID, org.StatusRejected, "licence unreadable"), http.StatusOK)
	if s := status(); s != org.StatusRejected {
		t.Fatalf("after reject, /me shows %s", s)
	}
	want(t, "resubmit", e.upload(admin, acmeSlug, acme.ID, org.DocLicence, pdfBytes), http.StatusCreated)
	if s := status(); s != org.StatusPending {
		t.Fatalf("after resubmitting, /me shows %s", s)
	}
	ok := verify(root, "", acme.ID, org.StatusVerified, "")
	want(t, "verify", ok, http.StatusOK)
	reviewed := decodeTo[org.Organisation](t, ok)
	if reviewed.Status != org.StatusVerified || reviewed.ReviewedBy == nil || reviewed.ReviewedAt == nil {
		t.Fatalf("review not recorded: %s", ok.Body)
	}
	if s := status(); s != org.StatusVerified {
		t.Fatalf("after verify, /me shows %s", s)
	}

	// Document list is paged like every other list.
	docs := do(e.h, http.MethodGet, as(acmeSlug, "/orgs/"+acme.ID+"/documents?"+listq.ParamSize+"=1"), admin, "")
	want(t, "list documents", docs, http.StatusOK)
	if p := decodeTo[listq.Page[org.Document]](t, docs); p.Total != 2 || len(p.Data) != 1 {
		t.Fatalf("document page wrong: %s", docs.Body)
	}
}

func TestCache(t *testing.T) {
	e, counting, cached := newEnv(t, true)
	acme := e.register(tAcme, org.KindProducer)
	memberID := e.addUser(tStaff, auth.RoleMember, acme.ID)
	token := e.login(tStaff)
	me := as(e.slug(tAcme), "/facilities") // a route the member's role allows

	want(t, "first request", do(e.h, http.MethodGet, me, token, ""), http.StatusOK)
	before := counting.lookups.Load()
	want(t, "second request", do(e.h, http.MethodGet, me, token, ""), http.StatusOK)
	if n := counting.lookups.Load() - before; n != 0 {
		t.Fatalf("second request made %d database lookups, want 0", n)
	}

	// Removal through the invalidation function is instant.
	e.exec(`UPDATE userscope SET is_deleted = true WHERE user_id = $1::uuid`, memberID)
	cached.InvalidateUser(context.Background(), memberID)
	want(t, "after removal", do(e.h, http.MethodGet, me, token, ""), http.StatusForbidden)

	// Gaining access is never delayed by the cache.
	e.exec(`UPDATE userscope SET is_deleted = false WHERE user_id = $1::uuid`, memberID)
	want(t, "after re-adding", do(e.h, http.MethodGet, me, token, ""), http.StatusOK)

	// With Redis unreachable, requests are served from the database.
	dead, err := cache.New("redis://127.0.0.1:1/0", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer dead.Close()
	a := *e.a
	a.Store = auth.CachedStore{Store: auth.PgStore{Pool: e.pool}, Cache: dead}
	h := server.NewRouter(deps(e.pool.Ping, &a))
	// /me: this router has no database behind its feature handlers.
	want(t, "redis down", do(h, http.MethodGet, as(e.slug(tAcme), "/me"), token, ""), http.StatusOK)
}

func TestEnsureSuperadmin(t *testing.T) {
	e, _, _ := newEnv(t, false)
	ctx := context.Background()
	pg := auth.PgStore{Pool: e.pool}

	hasOne := func() (exists bool) {
		e.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM userscope us JOIN roles r ON r.id = us.role_id
		                                     WHERE r.name = $1 AND NOT us.is_deleted)`, auth.RoleSuperadmin).Scan(&exists)
		return exists
	}

	// A short password never creates a superadmin. It is an error only when one
	// is needed: a platform that already has its superadmin must still start.
	short := strings.Repeat("a", auth.MinPasswordLen-1)
	created, err := pg.EnsureSuperadmin(ctx, e.email("seed"), short)
	if created || (err == nil) != hasOne() {
		t.Fatalf("short superadmin password: created=%v err=%v existing=%v", created, err, hasOne())
	}

	// Creates one only when the platform has none; a second call never adds another.
	created, err = pg.EnsureSuperadmin(ctx, e.email("seed"), testPassword)
	if err != nil {
		t.Fatal(err)
	}
	if created {
		token := e.login("seed")
		got := decodeTo[meResponse](t, do(e.h, http.MethodGet, "/me", token, ""))
		if got.Role != auth.RoleSuperadmin || len(got.Permissions) != len(auth.AllPermissions) {
			t.Fatalf("seeded superadmin scope wrong: %+v", got)
		}
	}
	again, err := pg.EnsureSuperadmin(ctx, e.email("seed-two"), testPassword)
	if err != nil || (created && again) {
		t.Fatalf("second seed: created=%v err=%v", again, err)
	}
}
