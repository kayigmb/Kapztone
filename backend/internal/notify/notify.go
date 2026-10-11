package notify

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

type Notification struct {
	ID             string     `json:"id"`
	OrganisationID string     `json:"organisation_id"`
	Kind           string     `json:"kind"`
	SubjectID      string     `json:"subject_id"`
	Message        string     `json:"message"`
	ReadAt         *time.Time `json:"read_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// List is what the list endpoint lets callers search and sort by.
var List = listq.Config{
	Alias:  "n",
	Search: []string{"message", "kind"},
	Sort:   []string{"kind", listq.ColCreatedAt},
}

const (
	columns = `n.id::text, n.organisation_id::text, n.kind, n.subject_id::text, n.message,
	           n.read_at, n.created_at, n.updated_at`
	gateColumn = "n.organisation_id"
)

func scan(row pgx.CollectableRow) (Notification, error) {
	var n Notification
	err := row.Scan(&n.ID, &n.OrganisationID, &n.Kind, &n.SubjectID, &n.Message, &n.ReadAt, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

// Add records a notification for an organisation. Like chain.Enqueue it is
// called with the transaction that makes the change it reports, so an
// organisation is never told about something that did not happen.
func Add(ctx context.Context, tx pgx.Tx, orgID, kind, subjectID, message string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO notifications (organisation_id, kind, subject_id, message) VALUES ($1::uuid, $2, $3::uuid, $4)`,
		orgID, kind, subjectID, message)
	return err
}

// ponytail: read state is per organisation, not per user: one colleague
// reading a notification marks it read for all. Add a per-user table if that
// starts to matter.
type Handlers struct {
	Pool *pgxpool.Pool
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
	unread := r.URL.Query().Get(ParamUnread) == "true"
	page, err := listq.Fetch(r.Context(), h.Pool, q, List, "SELECT "+columns, "FROM notifications n",
		[]string{auth.GateSQL(gateColumn, 1), `(NOT $2 OR n.read_at IS NULL)`}, []any{gate, unread}, scan)
	if err != nil {
		httpx.ServerError(w, "list notifications", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

// Read marks a notification as read. Reading it twice changes nothing.
func (h *Handlers) Read(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	rows, err := h.Pool.Query(r.Context(),
		`UPDATE notifications AS n SET read_at = COALESCE(n.read_at, now())
		 WHERE n.id = $1::uuid AND NOT n.is_deleted AND `+auth.GateSQL(gateColumn, 2)+` RETURNING `+columns, id, gate)
	if err != nil {
		httpx.ServerError(w, "read notification", err)
		return
	}
	n, err := pgx.CollectExactlyOneRow(rows, scan)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case err != nil:
		httpx.ServerError(w, "read notification", err)
	default:
		httpx.WriteJSON(w, http.StatusOK, n)
	}
}
