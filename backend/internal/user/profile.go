package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

// MyOrganisation is one organisation the signed-in person can act for: the
// organisation switcher lists these.
type MyOrganisation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Kind string `json:"kind"`
	Role string `json:"role"`
}

// Account is a user as the platform admin sees them: every organisation they
// belong or belonged to.
type Account struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Status    string `json:"status"`
	TwoFactor bool   `json:"two_factor"`
	// EmailVerified: the person opened the link that confirms their email.
	EmailVerified bool         `json:"email_verified"`
	CreatedAt     time.Time    `json:"created_at"`
	Memberships   []Membership `json:"memberships"`
}

type Membership struct {
	OrganisationID *string `json:"organisation_id"` // nil: the platform
	Organisation   string  `json:"organisation"`    // "" for the platform
	Role           string  `json:"role"`
	Status         string  `json:"status"`
}

// MyOrganisations lists the organisations the signed-in person belongs to (live memberships only).
func (s Store) MyOrganisations(ctx context.Context, userID string) ([]MyOrganisation, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT o.id::text, o.name, o.slug, o.kind, r.name
		 FROM userscope us
		 JOIN organisations o ON o.id = us.organisation_id AND NOT o.is_deleted
		 JOIN roles r ON r.id = us.role_id
		 WHERE us.user_id = $1::uuid AND NOT us.is_deleted
		 ORDER BY o.name`, userID)
	if err != nil {
		return nil, err
	}
	orgs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[MyOrganisation])
	if orgs == nil {
		orgs = []MyOrganisation{}
	}
	return orgs, err
}

// SetName changes the person's display name and returns the updated user.
func (s Store) SetName(ctx context.Context, userID, name string) (auth.User, error) {
	var u auth.User
	err := s.Pool.QueryRow(ctx,
		`UPDATE users SET name = $2 WHERE id = $1::uuid AND NOT is_deleted RETURNING id::text, email::text, name`,
		userID, name).Scan(&u.ID, &u.Email, &u.Name)
	return u, err
}

// SetEmailNotifications turns the e-mail copy of notifications on or off.
func (s Store) SetEmailNotifications(ctx context.Context, userID string, on bool) error {
	_, err := s.Pool.Exec(ctx, `UPDATE users SET email_notifications = $2 WHERE id = $1::uuid AND NOT is_deleted`, userID, on)
	return err
}

// AccountList is what GET /users lets the platform search and sort by.
var AccountList = listq.Config{
	Alias:  "m",
	Search: []string{"name", "email"},
	Sort:   []string{"name", "email", "status", listq.ColCreatedAt},
}

// All lists one page of every user on the platform with their memberships;
// a non-empty id keeps only that user.
func (s Store) All(ctx context.Context, q listq.Query, id string) (listq.Page[Account], error) {
	var where []string
	args := []any{StatusSuspended, StatusActive}
	if id != "" {
		where, args = []string{"m.id = $3::uuid"}, append(args, id)
	}
	return listq.Fetch(ctx, s.Pool, q, AccountList,
		`SELECT m.id::text, m.name, m.email, m.status, m.totp_enabled, m.is_verified, m.created_at, m.memberships`,
		`FROM (SELECT u.id, u.name, u.email::text AS email, u.status, u.totp_enabled, u.is_verified, u.created_at, false AS is_deleted,
		              COALESCE(json_agg(json_build_object(
		                  'organisation_id', us.organisation_id,
		                  'organisation', COALESCE(o.name, ''),
		                  'role', r.name,
		                  'status', CASE WHEN us.is_deleted THEN $1 ELSE $2 END)
		                ORDER BY o.name NULLS FIRST) FILTER (WHERE us.id IS NOT NULL), '[]') AS memberships
		       FROM users u
		       LEFT JOIN userscope us ON us.user_id = u.id
		       LEFT JOIN organisations o ON o.id = us.organisation_id
		       LEFT JOIN roles r ON r.id = us.role_id
		       WHERE NOT u.is_deleted
		       GROUP BY u.id) m`,
		where, args, func(row pgx.CollectableRow) (Account, error) {
			var a Account
			var memberships []byte
			if err := row.Scan(&a.ID, &a.Name, &a.Email, &a.Status, &a.TwoFactor, &a.EmailVerified, &a.CreatedAt, &memberships); err != nil {
				return a, err
			}
			return a, json.Unmarshal(memberships, &a.Memberships)
		})
}

// Account is one user as the platform admin sees them.
func (s Store) Account(ctx context.Context, id string) (Account, error) {
	page, err := s.All(ctx, listq.Query{Page: 1, Size: 1, SortCol: "name"}, id)
	if err == nil && len(page.Data) == 0 {
		err = ErrNotMember
	}
	if err != nil {
		return Account{}, err
	}
	return page.Data[0], nil
}

// SetActive turns the whole account on or off. Turning it off also voids
// every sign-in, so the person is signed out everywhere at once.
func (s Store) SetActive(ctx context.Context, id string, active bool) error {
	status := auth.UserDisabled
	if active {
		status = auth.UserActive
	}
	tag, err := s.Pool.Exec(ctx,
		`UPDATE users SET status = $2,
		        token_version = token_version + CASE WHEN $2 = $3 THEN 1 ELSE 0 END
		 WHERE id = $1::uuid AND NOT is_deleted`, id, status, auth.UserDisabled)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotMember
	}
	return err
}

// MyOrganisationsHandler answers GET /me/organisations.
func (h *Handlers) MyOrganisations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.Store.MyOrganisations(r.Context(), auth.CurrentUser(r.Context()).ID)
	if err != nil {
		httpx.ServerError(w, "list my organisations", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", orgs)
}

// SetName answers PUT /me: the signed-in person changes their own name.
func (h *Handlers) SetName(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLen {
		httpx.Error(w, http.StatusBadRequest, MsgNameLength)
		return
	}
	id := auth.CurrentUser(r.Context()).ID
	u, err := h.Store.SetName(r.Context(), id, name)
	if err != nil {
		httpx.ServerError(w, "set name", err)
		return
	}
	h.InvalidateUser(r.Context(), id) // the cached user carries the name
	httpx.OK(w, http.StatusOK, MsgNameSaved, u)
}

// SetEmailNotifications answers PUT /me/notifications: the signed-in person
// turns the e-mail copy of their notifications on or off.
func (h *Handlers) SetEmailNotifications(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email *bool `json:"email"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Email == nil {
		httpx.Error(w, http.StatusBadRequest, MsgEmailNotificationsRequired)
		return
	}
	if err := h.Store.SetEmailNotifications(r.Context(), auth.CurrentUser(r.Context()).ID, *in.Email); err != nil {
		httpx.ServerError(w, "set email notifications", err)
		return
	}
	msg := MsgEmailNotificationsOff
	if *in.Email {
		msg = MsgEmailNotificationsOn
	}
	httpx.OK(w, http.StatusOK, msg, map[string]bool{"email_notifications": *in.Email})
}
