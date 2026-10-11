package passport

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

type Handlers struct {
	Store Store
	// Now is the clock; nil means time.Now (tests set it).
	Now func() time.Time
}

func (h *Handlers) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

// errorStatus maps each refusal to its HTTP status and message. Anything not
// listed is a server error.
var errorStatus = map[error]struct {
	code int
	msg  string
}{
	ErrNotFound:          {http.StatusNotFound, httpx.MsgNotFound},
	ErrFacilityNotFound:  {http.StatusBadRequest, MsgFacilityNotFound},
	ErrTemplateNotFound:  {http.StatusBadRequest, MsgTemplateNotFound},
	ErrKindMismatch:      {http.StatusBadRequest, MsgKindMismatch},
	ErrUnitMismatch:      {http.StatusBadRequest, MsgUnitMismatch},
	ErrReceiverInvalid:   {http.StatusBadRequest, MsgReceiverInvalid},
	ErrFacilitySuspended: {http.StatusConflict, MsgFacilitySuspended},
	ErrNotActive:         {http.StatusConflict, MsgNotActive},
	ErrOpenHandover:      {http.StatusConflict, MsgOpenHandover},
	ErrHandoverClosed:    {http.StatusConflict, MsgHandoverClosed},
	ErrHandoverStale:     {http.StatusConflict, MsgHandoverStale},
	// Not holding a passport looks the same as the passport not existing.
	ErrNotHolder: {http.StatusNotFound, httpx.MsgNotFound},
}

// respond writes v, or the right error.
func respond(w http.ResponseWriter, status int, what string, v any, err error) {
	if err == nil {
		httpx.WriteJSON(w, status, v)
		return
	}
	var capacity CapacityError
	if errors.As(err, &capacity) {
		httpx.Error(w, http.StatusConflict, capacity.Error())
		return
	}
	for known, out := range errorStatus {
		if errors.Is(err, known) {
			httpx.Error(w, out.code, out.msg)
			return
		}
	}
	httpx.ServerError(w, what, err)
}

// acting returns the organisation and user the request acts as. Routes that
// change data require a verified organisation, so it is always present there.
func acting(w http.ResponseWriter, r *http.Request) (orgID, userID string, ok bool) {
	o := auth.CurrentScope(r.Context()).Organisation
	if o == nil {
		httpx.Error(w, http.StatusForbidden, auth.MsgNoOrgSelected)
		return "", "", false
	}
	return o.ID, auth.CurrentUser(r.Context()).ID, true
}

// Issue creates an origin passport for the acting organisation.
func (h *Handlers) Issue(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := acting(w, r)
	if !ok {
		return
	}
	var in IssueInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	p, err := h.Store.Issue(r.Context(), orgID, userID, in, h.now())
	respond(w, http.StatusCreated, "issue passport", p, err)
}

// statusFilter reads the optional ?status= and checks it against allowed.
func statusFilter(w http.ResponseWriter, r *http.Request, allowed []string) (string, bool) {
	status := r.URL.Query().Get(ParamStatus)
	if status != "" && !slices.Contains(allowed, status) {
		listq.BadParam(w, ParamStatus, allowed)
		return "", false
	}
	return status, true
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
	status, ok := statusFilter(w, r, Statuses)
	if !ok {
		return
	}
	page, err := h.Store.Find(r.Context(), q, gate, status)
	respond(w, http.StatusOK, "list passports", page, err)
}

// target reads the gate and the {id} a request is about.
func target(w http.ResponseWriter, r *http.Request) (id, gate string, ok bool) {
	if gate, ok = auth.Gate(w, r); !ok {
		return "", "", false
	}
	id, ok = httpx.URLID(w, r)
	return id, gate, ok
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	p, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "get passport", p, err)
}

// StartHandover is the sender's half of a handover.
func (h *Handlers) StartHandover(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := acting(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in HandoverInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	out, err := h.Store.StartHandover(r.Context(), id, orgID, userID, in, h.now())
	respond(w, http.StatusCreated, "start handover", out, err)
}

func (h *Handlers) ListHandovers(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, HandoverList)
	if !ok {
		return
	}
	status, ok := statusFilter(w, r, HandoverStatuses)
	if !ok {
		return
	}
	page, err := h.Store.FindHandovers(r.Context(), q, gate, status)
	respond(w, http.StatusOK, "list handovers", page, err)
}

func (h *Handlers) GetHandover(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	out, err := h.Store.GetHandover(r.Context(), id, gate)
	respond(w, http.StatusOK, "get handover", out, err)
}

// answer runs one of the three ways an awaiting handover ends.
func (h *Handlers) answer(w http.ResponseWriter, r *http.Request, what string,
	do func(id, orgID, userID string) (Handover, error)) {

	orgID, userID, ok := acting(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	out, err := do(id, orgID, userID)
	respond(w, http.StatusOK, what, out, err)
}

// Accept is the receiver's half: it signs, and the passport changes holder.
func (h *Handlers) Accept(w http.ResponseWriter, r *http.Request) {
	h.answer(w, r, "accept handover", func(id, orgID, userID string) (Handover, error) {
		return h.Store.Accept(r.Context(), id, orgID, userID)
	})
}

// Reject is the receiver refusing the goods, with a reason.
func (h *Handlers) Reject(w http.ResponseWriter, r *http.Request) {
	var in RejectInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	h.answer(w, r, "reject handover", func(id, orgID, userID string) (Handover, error) {
		return h.Store.Reject(r.Context(), id, orgID, userID, in.Reason)
	})
}

// Cancel is the sender withdrawing a handover nobody has answered yet.
func (h *Handlers) Cancel(w http.ResponseWriter, r *http.Request) {
	h.answer(w, r, "cancel handover", func(id, orgID, userID string) (Handover, error) {
		return h.Store.Cancel(r.Context(), id, orgID, userID)
	})
}
