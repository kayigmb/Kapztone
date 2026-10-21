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
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrResolved    = errors.New("already resolved")
	ErrBadAssignee = errors.New("not an investigator")
)

type Flag struct {
	ID             string  `json:"id"`
	PassportID     string  `json:"passport_id"`
	OrganisationID string  `json:"organisation_id"` // who raised it
	Reason         string  `json:"reason"`
	DataHash       string  `json:"data_hash"`
	Status         string  `json:"status"`
	Code           *string `json:"code,omitempty"` // FLG-YEAR-NNNN, set on raise
	Severity       string  `json:"severity"`       // low, medium, high
	RequestInfo    *string `json:"request_info,omitempty"`
	// RaisedBy is where it came from: SourceUser, SourceRules or SourcePublic.
	RaisedBy string `json:"raised_by"`
	// Who is looking into it; null until an investigator assigns someone.
	AssigneeUserID *string    `json:"assignee_user_id"`
	AssigneeName   *string    `json:"assignee_name"`
	ResolutionNote *string    `json:"resolution_note"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	TxHash         *string    `json:"tx_hash"`
	AnchoredAt     *time.Time `json:"anchored_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	// Revoked is how many passports a confirmed flag revoked; set on the resolve answer only.
	Revoked int `json:"revoked,omitempty"`
}

// Severities for flags.
const (
	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"
)

var Severities = []string{SeverityLow, SeverityMedium, SeverityHigh}

// List is what the list endpoint lets callers search and sort by.
var List = listq.Config{
	Alias:  "f",
	Search: []string{"reason", "status", "severity", "code"},
	Sort:   []string{"status", "severity", listq.ColCreatedAt},
}

const (
	columns = `f.id::text, f.passport_id::text, f.organisation_id::text, f.reason, f.data_hash, f.status,
	           f.code, f.severity, f.request_info,
	           f.source, f.assignee_user_id::text, (SELECT u.name FROM users u WHERE u.id = f.assignee_user_id),
	           f.resolution_note, f.resolved_at, f.tx_hash, f.anchored_at, f.created_at, f.updated_at`
	gateColumn = "f.organisation_id"
)

func scan(row pgx.CollectableRow) (Flag, error) {
	var f Flag
	err := row.Scan(&f.ID, &f.PassportID, &f.OrganisationID, &f.Reason, &f.DataHash, &f.Status,
		&f.Code, &f.Severity, &f.RequestInfo,
		&f.RaisedBy, &f.AssigneeUserID, &f.AssigneeName, &f.ResolutionNote, &f.ResolvedAt, &f.TxHash, &f.AnchoredAt, &f.CreatedAt, &f.UpdatedAt)
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
	Severity   string `json:"severity"`
	// Source is set by the code raising it, never by the caller; "" is SourceUser.
	Source string `json:"-"`
}

func (in *RaiseInput) Validate() string {
	in.Reason = strings.TrimSpace(in.Reason)
	in.Severity = strings.TrimSpace(strings.ToLower(in.Severity))
	if in.Severity == "" {
		in.Severity = SeverityMedium
	}
	switch {
	case !httpx.IsUUID(in.PassportID):
		return "passport_id is required"
	case in.Reason == "":
		return MsgReasonNeeded
	case len(in.Reason) > MaxTextLen:
		return fmt.Sprintf("reason may be up to %d characters", MaxTextLen)
	case !slices.Contains(Severities, in.Severity):
		return "severity must be one of: " + strings.Join(Severities, ", ")
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

	f, err := RaiseTx(ctx, tx, orgID, userID, holder, in)
	if err != nil {
		return Flag{}, err
	}
	return f, tx.Commit(ctx)
}

// RaiseTx is Raise inside the caller's transaction, so another change (a
// rejected handover) and its flag are saved together.
func RaiseTx(ctx context.Context, tx pgx.Tx, orgID, userID, holder string, in RaiseInput) (Flag, error) {
	var id string
	if err := tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&id); err != nil {
		return Flag{}, err
	}
	hash, err := chain.DataHash(record{id, in.PassportID, orgID, in.Reason})
	if err != nil {
		return Flag{}, err
	}
	dataHash := chain.Hex(hash)
	if in.Source == "" {
		in.Source = SourceUser
	}

	code := "FLG-" + time.Now().UTC().Format("2006") + "-" + strings.ToUpper(id[:8])
	f, err := one(tx.Query(ctx,
		`INSERT INTO flags AS f (id, passport_id, organisation_id, reason, data_hash, status, code, severity, raised_by, source)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9::uuid, $10) RETURNING `+columns,
		id, in.PassportID, orgID, in.Reason, dataHash, StatusOpen, code, in.Severity, userID, in.Source))
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
	// The platform resolves flags, so it hears of every new one.
	if err := notify.Add(ctx, tx, notify.Platform, notify.KindFlagRaised, id, MsgNotifyRaisedPlatform); err != nil {
		return Flag{}, err
	}
	return f, nil
}

// RaiseFlag is RaiseTx in the shape the passport package takes
// (passport.RaiseFlag): handover rejections and the rules raise flags with it.
func RaiseFlag(ctx context.Context, tx pgx.Tx, orgID, userID, holder, passportID, reason, source string) error {
	_, err := RaiseTx(ctx, tx, orgID, userID, holder, RaiseInput{PassportID: passportID, Reason: reason, Source: source})
	return err
}

// Find returns the flags an organisation raised (all of them for the platform).
func (s Store) Find(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Flag], error) {
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM flags f",
		[]string{auth.GateSQL(gateColumn, 1), `($2 = '' OR f.status = $2 OR ($2 = $3 AND f.status <> $4))`},
		[]any{gate, status, StatusResolved, StatusOpen}, scan)
}

func (s Store) Get(ctx context.Context, id, gate string) (Flag, error) {
	return one(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM flags f WHERE f.id = $1::uuid AND NOT f.is_deleted AND `+auth.GateSQL(gateColumn, 2), id, gate))
}

// Assign names who looks into a flag (nil: nobody). The assignee must be an
// investigator: an active user with a platform-wide scope, or a member of a
// verified regulator.
func (s Store) Assign(ctx context.Context, id string, assignee *string) (Flag, error) {
	if assignee != nil {
		var ok bool
		err := s.Pool.QueryRow(ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM users u JOIN userscope us ON us.user_id = u.id AND NOT us.is_deleted
			   LEFT JOIN organisations o ON o.id = us.organisation_id AND NOT o.is_deleted
			   WHERE u.id = $1::uuid AND u.status = $2 AND NOT u.is_deleted
			     AND (us.organisation_id IS NULL OR (o.kind = $3 AND o.status = $4)))`,
			*assignee, auth.UserActive, org.KindRegulator, auth.OrgVerified).Scan(&ok)
		if err != nil {
			return Flag{}, err
		}
		if !ok {
			return Flag{}, ErrBadAssignee
		}
	}
	return one(s.Pool.Query(ctx,
		`UPDATE flags AS f SET assignee_user_id = $2::uuid WHERE f.id = $1::uuid AND NOT f.is_deleted RETURNING `+columns,
		id, assignee))
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
	// Auth checks the signing PIN on resolution.
	Auth *auth.Auth
}

// investigator reports whether the caller looks into flags for the whole
// platform: the superadmin (flag:resolve), or anyone acting for a verified
// regulator. For flags only, they are not limited to their own organisation.
func investigator(r *http.Request) bool {
	s := auth.CurrentScope(r.Context())
	if slices.Contains(s.Permissions, auth.FlagResolve) {
		return true
	}
	o := s.Organisation
	return o != nil && o.Kind == org.KindRegulator && o.Status == auth.OrgVerified
}

// gate is auth.Gate, widened to every organisation for an investigator.
func gate(w http.ResponseWriter, r *http.Request) (string, bool) {
	if investigator(r) {
		return "", true
	}
	return auth.Gate(w, r)
}

func respond(w http.ResponseWriter, status int, message, what string, v any, err error) {
	switch {
	case err == nil:
		httpx.OK(w, status, message, v)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrResolved):
		httpx.Error(w, http.StatusConflict, MsgResolved)
	case errors.Is(err, ErrBadAssignee):
		httpx.Error(w, http.StatusBadRequest, MsgAssignee)
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
	respond(w, http.StatusCreated, MsgRaised, "raise flag", f, err)
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	gate, ok := gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, List)
	if !ok {
		return
	}
	status := r.URL.Query().Get(ParamStatus)
	if status != "" && status != StatusResolved && !slices.Contains(Statuses, status) {
		listq.BadParam(w, ParamStatus, append(slices.Clone(Statuses), StatusResolved))
		return
	}
	page, err := h.Store.Find(r.Context(), q, gate, status)
	respond(w, http.StatusOK, "", "list flags", page, err)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	gate, ok := gate(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "", "get flag", f, err)
}

// Assign answers PUT /flags/{id}/assignee {"user_id": "…" | null}: an investigator only.
func (h *Handlers) Assign(w http.ResponseWriter, r *http.Request) {
	if !investigator(r) {
		httpx.Error(w, http.StatusForbidden, auth.MsgNotAllowed)
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in struct {
		UserID *string `json:"user_id"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.UserID != nil && !httpx.IsUUID(*in.UserID) {
		httpx.Error(w, http.StatusBadRequest, MsgAssignee)
		return
	}
	f, err := h.Store.Assign(r.Context(), id, in.UserID)
	respond(w, http.StatusOK, MsgAssigned, "assign flag", f, err)
}

// Resolve clears or confirms a flag: an investigator only, with their signing PIN.
func (h *Handlers) Resolve(w http.ResponseWriter, r *http.Request) {
	if !investigator(r) {
		httpx.Error(w, http.StatusForbidden, auth.MsgNotAllowed)
		return
	}
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
	if !h.Auth.CheckPIN(w, r) {
		return
	}
	f, err := h.Store.Resolve(r.Context(), id, auth.CurrentUser(r.Context()).ID, in)
	respond(w, http.StatusOK, resolvedMessage[in.Outcome], "resolve flag", f, err)
}
