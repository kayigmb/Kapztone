package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgStore is the Postgres-backed Store.
type PgStore struct {
	Pool *pgxpool.Pool
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s PgStore) UserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx,
		`SELECT id::text, email::text, name, password_hash
		 FROM users WHERE email = $1 AND status = 'active'`, email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash)
	return u, notFound(err)
}

func (s PgStore) ActiveUser(ctx context.Context, id string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx,
		`SELECT id::text, email::text, name
		 FROM users WHERE id = $1::uuid AND status = 'active'`, id,
	).Scan(&u.ID, &u.Email, &u.Name)
	return u, notFound(err)
}

func (s PgStore) Scope(ctx context.Context, userID, slug string) (Scope, error) {
	var (
		sc                                Scope
		oID, oName, oSlug, oKind, oStatus *string
	)
	// An organisation-specific scope wins over the platform-wide one.
	err := s.Pool.QueryRow(ctx,
		`SELECT o.id::text, o.name, o.slug, o.kind, o.status, r.name,
		        COALESCE(array_agg(p.name ORDER BY p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		 FROM userscope us
		 JOIN roles r ON r.id = us.role_id
		 LEFT JOIN organisations o ON o.id = us.organisation_id
		 LEFT JOIN role_permissions rp ON rp.role_id = r.id
		 LEFT JOIN permissions p ON p.id = rp.permission_id
		 WHERE us.user_id = $1::uuid AND (us.organisation_id IS NULL OR o.slug = $2)
		 GROUP BY us.id, o.id, r.name
		 ORDER BY us.organisation_id IS NULL
		 LIMIT 1`, userID, slug,
	).Scan(&oID, &oName, &oSlug, &oKind, &oStatus, &sc.Role, &sc.Permissions)
	if err != nil {
		return Scope{}, notFound(err)
	}
	if oID != nil {
		sc.Organisation = &Organisation{ID: *oID, Name: *oName, Slug: *oSlug, Kind: *oKind, Status: *oStatus}
	}
	return sc, nil
}

// EnsureSuperadmin creates the first superadmin if the platform has none.
// It is safe to call on every start and from several instances at once.
func (s PgStore) EnsureSuperadmin(ctx context.Context, email, password string) (created bool, err error) {
	if len(password) < 12 {
		return false, errors.New("superadmin password must be at least 12 characters")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("hash superadmin password: %w", err)
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	// Arbitrary constant key: serialises concurrent seeders.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7203114)`); err != nil {
		return false, err
	}
	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM userscope us JOIN roles r ON r.id = us.role_id
		                WHERE r.name = 'superadmin')`).Scan(&exists)
	if err != nil || exists {
		return false, err
	}

	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, 'Platform admin')
		 RETURNING id::text`, email, hash).Scan(&userID)
	if err != nil {
		return false, fmt.Errorf("create superadmin user: %w", err)
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO userscope (user_id, role_id)
		 SELECT $1::uuid, id FROM roles WHERE name = 'superadmin'`, userID)
	if err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}
