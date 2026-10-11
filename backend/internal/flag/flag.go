package flag

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
)

var (
	ErrNotFound = errors.New("not found")
	ErrResolved = errors.New("already resolved")
)

type Flag struct {
	ID             string     `json:"id"`
	PassportID     string     `json:"passport_id"`
	OrganisationID string     `json:"organisation_id"` // who raised it
	Reason         string     `json:"reason"`
	DataHash       string     `json:"data_hash"`
	Status         string     `json:"status"`
	ResolutionNote *string    `json:"resolution_note"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	TxHash         *string    `json:"tx_hash"`
	AnchoredAt     *time.Time `json:"anchored_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	// Revoked is how many passports a confirmed flag revoked; set on the resolve answer only.
	Revoked int `json:"revoked,omitempty"`
}

// List is what the list endpoint lets callers search and sort by.
var List = listq.Config{
	Alias:  "f",
	Search: []string{"reason", "status"},
	Sort:   []string{"status", listq.ColCreatedAt},
}

const (
	columns = `f.id::text, f.passport_id::text, f.organisation_id::text, f.reason, f.data_hash, f.status,
	           f.resolution_note, f.resolved_at, f.tx_hash, f.anchored_at, f.created_at, f.updated_at`
	gateColumn = "f.organisation_id"
)

func scan(row pgx.CollectableRow) (Flag, error) {
	var f Flag
	err := row.Scan(&f.ID, &f.PassportID, &f.OrganisationID, &f.Reason, &f.DataHash, &f.Status,
		&f.ResolutionNote, &f.ResolvedAt, &f.TxHash, &f.AnchoredAt, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func one(rows pgx.Rows, err error) (Flag, error) {
	if err != nil {
		return Flag{}, err
	}
	f, err := pgx.CollectExactlyOneRow(rows, scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Flag{}, ErrNotFound
	}
	return f, err
}

// RaiseInput is a concern about one passport.
type RaiseInput struct {
	PassportID string `json:"passport_id"`
	Reason     string `json:"reason"`
}

func (in *RaiseInput) Validate() string {
	in.Reason = strings.TrimSpace(in.Reason)
	switch {
	case !httpx.IsUUID(in.PassportID):
		return "passport_id is required"
	case in.Reason == "":
		return MsgReasonNeeded
	case len(in.Reason) > MaxTextLen:
		return fmt.Sprintf("reason may be up to %d characters", MaxTextLen)
	}
	return ""
}

// ResolveInput is the resolver's decision.
type ResolveInput struct {
	Outcome string `json:"outcome"`
	Note    string `json:"note"`
}

func (in *ResolveInput) Validate() string {
	in.Note = strings.TrimSpace(in.Note)
	switch {
	case !slices.Contains(Outcomes, in.Outcome):
		return "outcome must be one of: " + strings.Join(Outcomes, ", ")
	case len(in.Note) > MaxTextLen:
		return fmt.Sprintf("note may be up to %d characters", MaxTextLen)
	}
	return ""
}

// record is what a flag's data hash covers. That hash is the reason the
// chain records, both for the flag and for any revocation it leads to.
type record struct {
	ID             string `json:"id"`
	PassportID     string `json:"passport_id"`
	OrganisationID string `json:"organisation_id"`
	Reason         string `json:"reason"`
}

type Store struct {
	Pool *pgxpool.Pool
}

// Raise records a concern about a passport held by holder, and queues it for the chain.
func (s Store) Raise(ctx context.Context, orgID, userID, holder string, in RaiseInput) (Flag, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Flag{}, err
	}
	defer tx.Rollback(ctx)

	var id string
	if err := tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&id); err != nil {
		return Flag{}, err
	}
	hash, err := chain.DataHash(record{id, in.PassportID, orgID, in.Reason})
	if err != nil {
		return Flag{}, err
	}
	dataHash := chain.Hex(hash)

	f, err := one(tx.Query(ctx,
		`INSERT INTO flags AS f (id, passport_id, organisation_id, reason, data_hash, status, raised_by)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7::uuid) RETURNING `+columns,
		id, in.PassportID, orgID, in.Reason, dataHash, StatusOpen, userID))
	if err != nil {
		return Flag{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, from_org, quantity, data_hash)
		 VALUES ($1::uuid, $2::uuid, $3, $4::uuid, 0, $5)`, in.PassportID, id, passport.EventFlag, orgID, dataHash); err != nil {
		return Flag{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindFlag, id, chain.PassportReason{PassportID: in.PassportID, ReasonHash: dataHash})
	if err != nil {
		return Flag{}, err
	}
	if holder != orgID {
		if err := notify.Add(ctx, tx, holder, notify.KindFlagRaised, id, MsgNotifyRaised); err != nil {
			return Flag{}, err
		}
	}
	return f, tx.Commit(ctx)
}

// Find returns the flags an organisation raised (all of them for the platform).
func (s Store) Find(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Flag], error) {
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM flags f",
		[]string{auth.GateSQL(gateColumn, 1), `($2 = '' OR f.status = $2)`}, []any{gate, status}, scan)
}

func (s Store) Get(ctx context.Context, id, gate string) (Flag, error) {
	return one(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM flags f WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2), id, gate))
}

// Resolve closes an open flag. Confirming it revokes the passport and every
// passport made from it, in the same transaction.
func (s Store) Resolve(ctx context.Context, id, userID string, in ResolveInput) (Flag, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Flag{}, err
	}
	defer tx.Rollback(ctx)

	var status, passportID, dataHash string
	err = tx.QueryRow(ctx,
		`SELECT status, passport_id::text, data_hash FROM flags WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, id,
	).Scan(&status, &passportID, &dataHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Flag{}, ErrNotFound
	}
	if err != nil {
		return Flag{}, err
	}
	if status != StatusOpen {
		return Flag{}, ErrResolved
	}

	f, err := one(tx.Query(ctx,
		`UPDATE flags AS f SET status = $2, resolution_note = NULLIF($3, ''), resolved_by = $4::uuid, resolved_at = now()
		 WHERE f.id = $1::uuid RETURNING `+columns, id, in.Outcome, in.Note, userID))
	if err != nil {
		return Flag{}, err
	}
	if in.Outcome == StatusConfirmed {
		if f.Revoked, err = passport.RevokeTree(ctx, tx, passportID, id, dataHash); err != nil {
			return Flag{}, err
		}
	}
	if err := notify.Add(ctx, tx, f.OrganisationID, notify.KindFlagResolved, id, MsgNotifyResolved); err != nil {
		return Flag{}, err
	}
	return f, tx.Commit(ctx)
}

type Handlers struct {
	Store Store
	// Passports answers whether the caller may see the passport they flag.
	Passports passport.Store
}

func respond(w http.ResponseWriter, status int, what string, v any, err error) {
	switch {
	case err == nil:
		httpx.WriteJSON(w, status, v)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrResolved):
		httpx.Error(w, http.StatusConflict, MsgResolved)
	default:
		httpx.ServerError(w, what, err)
	}
}

// Raise flags a passport the acting organisation can see.
func (h *Handlers) Raise(w http.ResponseWriter, r *http.Request) {
	o := auth.CurrentScope(r.Context()).Organisation
	if o == nil {
		httpx.Error(w, http.StatusForbidden, auth.MsgNoOrgSelected)
		return
	}
	var in RaiseInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	// A passport outside the caller's view looks exactly like a missing one.
	p, err := h.Passports.Get(r.Context(), in.PassportID, o.ID)
	if errors.Is(err, passport.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if err != nil {
		httpx.ServerError(w, "get passport", err)
		return
	}
	f, err := h.Store.Raise(r.Context(), o.ID, auth.CurrentUser(r.Context()).ID, p.OrganisationID, in)
	respond(w, http.StatusCreated, "raise flag", f, err)
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, List)
	if !ok {
		return
	}
	status := r.URL.Query().Get(ParamStatus)
	if status != "" && !slices.Contains(Statuses, status) {
		listq.BadParam(w, ParamStatus, Statuses)
		return
	}
	page, err := h.Store.Find(r.Context(), q, gate, status)
	respond(w, http.StatusOK, "list flags", page, err)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "get flag", f, err)
}

// Resolve clears or confirms a flag. The route's permission decides who may.
func (h *Handlers) Resolve(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in ResolveInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	f, err := h.Store.Resolve(r.Context(), id, auth.CurrentUser(r.Context()).ID, in)
	respond(w, http.StatusOK, "resolve flag", f, err)
}
