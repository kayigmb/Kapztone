package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/dbx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var (
	ErrInviteInvalid  = errors.New("invite invalid")
	ErrAlreadyMember  = errors.New("already a member")
	ErrLastAdmin      = errors.New("last admin")
	ErrNotMember      = errors.New("not a member")
	ErrNotSuspended   = errors.New("not suspended")
	ErrBadPassword    = errors.New("bad password")
	ErrNameNeeded     = errors.New("name needed")
	ErrPasswordPolicy = errors.New("password does not meet policy")
)

// Member is one person's place in an organisation.
type Member struct {
	ID     string    `json:"id"` // the user's id
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Status string    `json:"status"`
	Since  time.Time `json:"since"`
}

type Invite struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Accepted is what accepting an invitation did.
type Accepted struct {
	UserID       string `json:"-"`
	Organisation string `json:"organisation"` // the organisation's address (slug), to sign in at
	NewAccount   bool   `json:"new_account"`
}

type Store struct {
	Pool *pgxpool.Pool
}

// MemberList is what GET /users lets an organisation search and sort by.
var MemberList = listq.Config{
	Alias:  "m",
	Search: []string{"name", "email", "role"},
	Sort:   []string{"name", "email", "role", "status", listq.ColCreatedAt},
}

// List returns one page of everyone who belongs or belonged to the
// organisation, each with their current place: a live scope means active,
// only a deleted one means suspended. created_at is when they joined.
func (s Store) List(ctx context.Context, q listq.Query, orgID string) (listq.Page[Member], error) {
	return listq.Fetch(ctx, s.Pool, q, MemberList,
		`SELECT m.id::text, m.name, m.email, m.role, m.status, m.created_at`,
		`FROM (SELECT DISTINCT ON (u.id) u.id, u.name, u.email::text AS email, r.name AS role,
		              CASE WHEN us.is_deleted THEN $2 ELSE $3 END AS status, us.created_at, false AS is_deleted
		       FROM userscope us
		       JOIN users u ON u.id = us.user_id AND NOT u.is_deleted
		       JOIN roles r ON r.id = us.role_id
		       WHERE us.organisation_id = $1::uuid
		       ORDER BY u.id, us.is_deleted, us.updated_at DESC) m`,
		nil, []any{orgID, StatusSuspended, StatusActive}, pgx.RowToStructByPos[Member])
}

func (s Store) CreateInvite(ctx context.Context, orgID, email, role, tokenHash, invitedBy string, expires time.Time) (Invite, error) {
	var inv Invite
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO invites (organisation_id, email, role, token_hash, expires_at, invited_by)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6::uuid)
		 RETURNING id::text, email::text, role, expires_at`,
		orgID, email, role, tokenHash, expires, invitedBy).Scan(&inv.ID, &inv.Email, &inv.Role, &inv.ExpiresAt)
	return inv, err
}

// Accept spends an invitation. A new e-mail gets an account with name and
// password; an existing account proves it is theirs with its password. Either
// way the person joins the organisation with the invited role.
func (s Store) Accept(ctx context.Context, tokenHash, name, password string) (Accepted, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Accepted{}, err
	}
	defer tx.Rollback(ctx)

	var inviteID, orgID, email, role, slug string
	err = tx.QueryRow(ctx,
		`SELECT i.id::text, i.organisation_id::text, i.email::text, i.role, o.slug
		 FROM invites i JOIN organisations o ON o.id = i.organisation_id AND NOT o.is_deleted
		 WHERE i.token_hash = $1 AND i.accepted_at IS NULL AND i.expires_at > now() AND NOT i.is_deleted
		 FOR UPDATE OF i`, tokenHash).Scan(&inviteID, &orgID, &email, &role, &slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Accepted{}, ErrInviteInvalid
	}
	if err != nil {
		return Accepted{}, err
	}

	out := Accepted{Organisation: slug}
	var hash string
	err = tx.QueryRow(ctx,
		`SELECT id::text, password_hash FROM users WHERE email = $1 AND status = $2 AND NOT is_deleted`,
		email, auth.UserActive).Scan(&out.UserID, &hash)
	switch {
	case errors.Is(err, pgx.ErrNoRows) && name == "":
		return Accepted{}, ErrNameNeeded
	case errors.Is(err, pgx.ErrNoRows):
		if auth.PasswordFormatProblem(password) != "" {
			return Accepted{}, ErrPasswordPolicy
		}
		newHash, err := auth.HashPassword(password)
		if err != nil {
			return Accepted{}, err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO users (email, password_hash, name, status, email_verified_at) VALUES ($1, $2, $3, $4, now()) RETURNING id::text`,
			email, newHash, name, auth.UserActive).Scan(&out.UserID); err != nil {
			return Accepted{}, err
		}
		out.NewAccount = true
	case err != nil:
		return Accepted{}, err
	case !auth.CheckPassword(hash, password):
		return Accepted{}, ErrBadPassword
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO userscope (user_id, role_id, organisation_id)
		 SELECT $1::uuid, id, $2::uuid FROM roles WHERE name = $3 AND NOT is_deleted`, out.UserID, orgID, role)
	if _, unique := dbx.UniqueViolation(err); unique {
		return Accepted{}, ErrAlreadyMember
	}
	if err != nil {
		return Accepted{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE invites SET accepted_at = now() WHERE id = $1::uuid`, inviteID); err != nil {
		return Accepted{}, err
	}
	return out, tx.Commit(ctx)
}

// lockOrg serialises membership changes in one organisation, so two admins
// cannot both demote each other and leave none.
func lockOrg(ctx context.Context, tx pgx.Tx, orgID string) error {
	_, err := tx.Exec(ctx, `SELECT 1 FROM organisations WHERE id = $1::uuid FOR UPDATE`, orgID)
	return err
}

// liveRole is the member's current role in the organisation.
func liveRole(ctx context.Context, tx pgx.Tx, orgID, userID string) (string, error) {
	var role string
	err := tx.QueryRow(ctx,
		`SELECT r.name FROM userscope us JOIN roles r ON r.id = us.role_id
		 WHERE us.organisation_id = $1::uuid AND us.user_id = $2::uuid AND NOT us.is_deleted`, orgID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotMember
	}
	return role, err
}

// keepsAnAdmin refuses a change that would take away the last admin.
func keepsAnAdmin(ctx context.Context, tx pgx.Tx, orgID, currentRole string) error {
	if currentRole != auth.RoleOrgAdmin {
		return nil
	}
	var admins int
	err := tx.QueryRow(ctx,
		`SELECT count(*) FROM userscope us JOIN roles r ON r.id = us.role_id
		 WHERE us.organisation_id = $1::uuid AND r.name = $2 AND NOT us.is_deleted`, orgID, auth.RoleOrgAdmin).Scan(&admins)
	if err == nil && admins <= 1 {
		return ErrLastAdmin
	}
	return err
}

// change runs one membership change under the organisation lock.
func (s Store) change(ctx context.Context, orgID string, fn func(tx pgx.Tx) error) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := lockOrg(ctx, tx, orgID); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) SetRole(ctx context.Context, orgID, userID, role string) error {
	return s.change(ctx, orgID, func(tx pgx.Tx) error {
		current, err := liveRole(ctx, tx, orgID, userID)
		if err != nil || current == role {
			return err
		}
		if err := keepsAnAdmin(ctx, tx, orgID, current); err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`UPDATE userscope SET role_id = (SELECT id FROM roles WHERE name = $3 AND NOT is_deleted)
			 WHERE organisation_id = $1::uuid AND user_id = $2::uuid AND NOT is_deleted`, orgID, userID, role)
		return err
	})
}

func (s Store) Suspend(ctx context.Context, orgID, userID string) error {
	return s.change(ctx, orgID, func(tx pgx.Tx) error {
		current, err := liveRole(ctx, tx, orgID, userID)
		if err != nil {
			return err
		}
		if err := keepsAnAdmin(ctx, tx, orgID, current); err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`UPDATE userscope SET is_deleted = true
			 WHERE organisation_id = $1::uuid AND user_id = $2::uuid AND NOT is_deleted`, orgID, userID)
		return err
	})
}

// Reinstate brings back the member's most recent place in the organisation.
func (s Store) Reinstate(ctx context.Context, orgID, userID string) error {
	return s.change(ctx, orgID, func(tx pgx.Tx) error {
		if _, err := liveRole(ctx, tx, orgID, userID); !errors.Is(err, ErrNotMember) {
			if err == nil {
				return ErrNotSuspended
			}
			return err
		}
		tag, err := tx.Exec(ctx,
			`UPDATE userscope SET is_deleted = false WHERE id = (
			   SELECT us.id FROM userscope us JOIN users u ON u.id = us.user_id AND NOT u.is_deleted
			   WHERE us.organisation_id = $1::uuid AND us.user_id = $2::uuid AND us.is_deleted
			   ORDER BY us.updated_at DESC LIMIT 1)`, orgID, userID)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNotMember
		}
		return err
	})
}
