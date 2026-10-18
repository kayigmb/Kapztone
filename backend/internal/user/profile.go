package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
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
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Status      string       `json:"status"`
	TwoFactor   bool         `json:"two_factor"`
	Memberships []Membership `json:"memberships"`
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

// All lists every user on the platform with their memberships.
// ponytail: plain array, not paged; move to listq once the platform has thousands of users.
func (s Store) All(ctx context.Context) ([]Account, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT u.id::text, u.name, u.email::text, u.status, u.totp_enabled,
		        COALESCE(json_agg(json_build_object(
		            'organisation_id', us.organisation_id,
		            'organisation', COALESCE(o.name, ''),
		            'role', r.name,
		            'status', CASE WHEN us.is_deleted THEN $1 ELSE $2 END)
		          ORDER BY o.name NULLS FIRST) FILTER (WHERE us.id IS NOT NULL), '[]')
		 FROM users u
		 LEFT JOIN userscope us ON us.user_id = u.id
		 LEFT JOIN organisations o ON o.id = us.organisation_id
		 LEFT JOIN roles r ON r.id = us.role_id
		 WHERE NOT u.is_deleted
		 GROUP BY u.id
		 ORDER BY u.name`, StatusSuspended, StatusActive)
	if err != nil {
		return nil, err
	}
	accounts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Account, error) {
		var a Account
		var memberships []byte
		if err := row.Scan(&a.ID, &a.Name, &a.Email, &a.Status, &a.TwoFactor, &memberships); err != nil {
			return a, err
		}
		return a, json.Unmarshal(memberships, &a.Memberships)
	})
	if accounts == nil {
		accounts = []Account{}
	}
	return accounts, err
}

// MyOrganisationsHandler answers GET /me/organisations.
func (h *Handlers) MyOrganisations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.Store.MyOrganisations(r.Context(), auth.CurrentUser(r.Context()).ID)
	if err != nil {
		httpx.ServerError(w, "list my organisations", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, orgs)
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
	httpx.WriteJSON(w, http.StatusOK, u)
}
