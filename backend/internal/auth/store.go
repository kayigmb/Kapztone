package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgStore is the Postgres-backed Store. Every query ignores deleted rows: a
// deleted user, organisation, role link or scope must grant nothing.
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
		`SELECT id::text, email::text, name, password_hash, totp_secret, totp_enabled
		 FROM users WHERE email = $1 AND status = $2 AND NOT is_deleted`, email, UserActive,
	).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.TOTPSecret, &u.TOTPEnabled)
	return u, notFound(err)
}

func (s PgStore) ActiveUser(ctx context.Context, id string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx,
		`SELECT id::text, email::text, name
		 FROM users WHERE id = $1::uuid AND status = $2 AND NOT is_deleted`, id, UserActive,
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
		`SELECT o.id::text, o.name, o.slug, o.kind, o.status, r.name, us.organisation_id IS NULL,
		        COALESCE(array_agg(p.name ORDER BY p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		 FROM userscope us
		 JOIN roles r ON r.id = us.role_id AND NOT r.is_deleted
		 LEFT JOIN organisations o ON o.id = us.organisation_id
		 LEFT JOIN role_permissions rp ON rp.role_id = r.id AND NOT rp.is_deleted
		 LEFT JOIN permissions p ON p.id = rp.permission_id AND NOT p.is_deleted
		 WHERE us.user_id = $1::uuid AND NOT us.is_deleted
		   AND (us.organisation_id IS NULL OR (o.slug = $2 AND NOT o.is_deleted))
		 GROUP BY us.id, o.id, r.name
		 ORDER BY us.organisation_id IS NULL
		 LIMIT 1`, userID, slug,
	).Scan(&oID, &oName, &oSlug, &oKind, &oStatus, &sc.Role, &sc.Platform, &sc.Permissions)
	if err != nil {
		return Scope{}, notFound(err)
	}
	if oID != nil {
		sc.Organisation = &Organisation{ID: *oID, Name: *oName, Slug: *oSlug, Kind: *oKind, Status: *oStatus}
	}
	return sc, nil
}

func (s PgStore) SetTOTP(ctx context.Context, userID, sealedSecret string, enabled bool) error {
	_, err := s.Pool.Exec(ctx,
		// A new secret starts a new step history; the same secret keeps its
		// burned steps, so the code that enabled 2FA cannot also sign in.
		`UPDATE users SET totp_enabled = $3,
		        totp_last_step = CASE WHEN totp_secret = $2 THEN totp_last_step ELSE 0 END,
		        totp_secret = $2
		 WHERE id = $1::uuid AND NOT is_deleted`, userID, sealedSecret, enabled)
	return err
}

// UseTOTPStep moves the user's last used time-step forward. The comparison
// and the update are one statement, so two requests cannot both win.
func (s PgStore) UseTOTPStep(ctx context.Context, userID string, step int64) (bool, error) {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE users SET totp_last_step = $2
		 WHERE id = $1::uuid AND totp_last_step < $2 AND NOT is_deleted`, userID, step)
	return tag.RowsAffected() == 1, err
}

func (s PgStore) PINHash(ctx context.Context, userID string) (string, error) {
	var hash *string
	err := s.Pool.QueryRow(ctx,
		`SELECT pin_hash FROM users WHERE id = $1::uuid AND NOT is_deleted`, userID).Scan(&hash)
	if err != nil || hash == nil {
		return "", notFound(err)
	}
	return *hash, nil
}

func (s PgStore) SetPIN(ctx context.Context, userID, hash string) error {
	_, err := s.Pool.Exec(ctx, `UPDATE users SET pin_hash = $2 WHERE id = $1::uuid AND NOT is_deleted`, userID, hash)
	return err
}

func (s PgStore) SetRecoveryCodes(ctx context.Context, userID string, hashes []string) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx,
		`UPDATE recovery_codes SET is_deleted = true WHERE user_id = $1::uuid AND NOT is_deleted`, userID); err != nil {
		return err
	}
	if len(hashes) > 0 {
		if _, err := tx.Exec(ctx,
			`INSERT INTO recovery_codes (user_id, code_hash) SELECT $1::uuid, unnest($2::text[])`, userID, hashes); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// UseRecoveryCode marks the code used in one statement, so two requests
// cannot both spend it.
func (s PgStore) UseRecoveryCode(ctx context.Context, userID, hash string) (bool, error) {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE recovery_codes SET used_at = now()
		 WHERE user_id = $1::uuid AND code_hash = $2 AND used_at IS NULL AND NOT is_deleted`, userID, hash)
	return tag.RowsAffected() == 1, err
}

// EnsureSuperadmin creates the first superadmin if the platform has none.
// It needs SyncRoles to have run. Safe to call on every start and from
// several instances at once.
func (s PgStore) EnsureSuperadmin(ctx context.Context, email, password string) (created bool, err error) {
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
		                WHERE r.name = $1 AND NOT us.is_deleted)`, RoleSuperadmin).Scan(&exists)
	if err != nil || exists {
		return false, err
	}

	// The length rule (by environment) lives in config, which runs first.
	hash, err := HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("hash superadmin password: %w", err)
	}

	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, status) VALUES ($1, $2, $3, $4)
		 RETURNING id::text`, email, hash, superadminName, UserActive).Scan(&userID)
	if err != nil {
		return false, fmt.Errorf("create superadmin user: %w", err)
	}
	tag, err := tx.Exec(ctx,
		`INSERT INTO userscope (user_id, role_id)
		 SELECT $1::uuid, id FROM roles WHERE name = $2 AND NOT is_deleted`, userID, RoleSuperadmin)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() != 1 {
		return false, fmt.Errorf("role %q is missing: run SyncRoles first", RoleSuperadmin)
	}
	return true, tx.Commit(ctx)
}
