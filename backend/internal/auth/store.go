package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		`SELECT id::text, email::text, name, password_hash, totp_secret, totp_enabled, token_version,
		        NOT is_verified
		 FROM users WHERE email = $1 AND status = $2 AND NOT is_deleted`, email, UserActive,
	).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.TOTPSecret, &u.TOTPEnabled, &u.TokenVersion, &u.EmailUnverified)
	return u, notFound(err)
}

func (s PgStore) ActiveUser(ctx context.Context, id string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx,
		`SELECT id::text, email::text, name, token_version
		 FROM users WHERE id = $1::uuid AND status = $2 AND NOT is_deleted`, id, UserActive,
	).Scan(&u.ID, &u.Email, &u.Name, &u.TokenVersion)
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

// ---- AccountStore ----

func (s PgStore) UserByID(ctx context.Context, userID string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx,
		`SELECT id::text, email::text, name, password_hash, token_version
		 FROM users WHERE id = $1::uuid AND NOT is_deleted`, userID,
	).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.TokenVersion)
	return u, notFound(err)
}

func (s PgStore) RecentPasswordHashes(ctx context.Context, userID string, n int) ([]string, error) {
	rows, err := s.Pool.Query(ctx,
		`(SELECT password_hash FROM users WHERE id = $1::uuid AND NOT is_deleted)
		 UNION ALL
		 (SELECT password_hash FROM password_history WHERE user_id = $1::uuid AND NOT is_deleted
		  ORDER BY created_at DESC LIMIT $2)`, userID, max(n-1, 0))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[string])
}

func (s PgStore) SetPassword(ctx context.Context, userID, hash string) (int, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	// The password being replaced goes into the history; the current one is
	// always compared too (see RecentPasswordHashes).
	if _, err := tx.Exec(ctx,
		`INSERT INTO password_history (user_id, password_hash)
		 SELECT id, password_hash FROM users WHERE id = $1::uuid AND NOT is_deleted`, userID); err != nil {
		return 0, err
	}
	var ver int
	err = tx.QueryRow(ctx,
		`UPDATE users SET password_hash = $2, token_version = token_version + 1
		 WHERE id = $1::uuid AND NOT is_deleted RETURNING token_version`, userID, hash).Scan(&ver)
	if err != nil {
		return 0, notFound(err)
	}
	return ver, tx.Commit(ctx)
}

func (s PgStore) BumpTokenVersion(ctx context.Context, userID string) (int, error) {
	var ver int
	err := s.Pool.QueryRow(ctx,
		`UPDATE users SET token_version = token_version + 1
		 WHERE id = $1::uuid AND NOT is_deleted RETURNING token_version`, userID).Scan(&ver)
	return ver, notFound(err)
}

func (s PgStore) AddLink(ctx context.Context, userID, kind, tokenHash string, expires time.Time) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO password_resets (user_id, kind, token_hash, expires_at) VALUES ($1::uuid, $2, $3, $4)`,
		userID, kind, tokenHash, expires)
	return err
}

func (s PgStore) LinkUser(ctx context.Context, kind, tokenHash string, now time.Time) (string, error) {
	var userID string
	err := s.Pool.QueryRow(ctx,
		`SELECT user_id::text FROM password_resets
		 WHERE kind = $1 AND token_hash = $2 AND used_at IS NULL AND expires_at > $3 AND NOT is_deleted`,
		kind, tokenHash, now).Scan(&userID)
	return userID, notFound(err)
}

// UseLink spends the link in one statement, so two requests cannot both use it.
func (s PgStore) UseLink(ctx context.Context, kind, tokenHash string, now time.Time) (string, error) {
	var userID string
	err := s.Pool.QueryRow(ctx,
		`UPDATE password_resets SET used_at = $3
		 WHERE kind = $1 AND token_hash = $2 AND used_at IS NULL AND expires_at > $3 AND NOT is_deleted
		 RETURNING user_id::text`, kind, tokenHash, now).Scan(&userID)
	return userID, notFound(err)
}

func (s PgStore) Lock(ctx context.Context, userID string) error {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE users SET status = $2, token_version = token_version + 1
		 WHERE id = $1::uuid AND NOT is_deleted`, userID, UserDisabled)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (s PgStore) MarkVerified(ctx context.Context, userID string) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE users SET email_verified_at = COALESCE(email_verified_at, now())
		 WHERE id = $1::uuid AND NOT is_deleted`, userID)
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
		`INSERT INTO users (email, password_hash, name, status, email_verified_at) VALUES ($1, $2, $3, $4, now())
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
