package passport

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

type Handlers struct {
	Store Store
	// Files holds handover photos.
	Files *upload.Bucket
	// RaiseFlag raises a flag inside a transaction (flag.RaiseTx), for a
	// rejection that also flags the passport.
	RaiseFlag RaiseFlag
	// PIN checks the signing PIN of every signing action (auth.Auth.CheckPIN),
	// writing the refusal itself.
	PIN func(http.ResponseWriter, *http.Request) bool
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
	ErrFacilityPending:   {http.StatusConflict, MsgFacilityPending},
	ErrLicenceExpired:    {http.StatusConflict, MsgLicenceExpired},
	ErrNotActive:         {http.StatusConflict, MsgNotActive},
	ErrOpenHandover:      {http.StatusConflict, MsgOpenHandover},
	ErrHandoverClosed:    {http.StatusConflict, MsgHandoverClosed},
	ErrHandoverStale:     {http.StatusConflict, MsgHandoverStale},
	ErrInputNotHeld:      {http.StatusBadRequest, MsgInputNotHeld},
	ErrMixedInputs:       {http.StatusBadRequest, MsgMixedInputs},
	ErrStageInvalid:      {http.StatusBadRequest, MsgStageInvalid},
	ErrHandoverStage:     {http.StatusBadRequest, MsgHandoverStage},
	// Not holding a passport looks the same as the passport not existing.
	ErrNotHolder: {http.StatusNotFound, httpx.MsgNotFound},
}

// respond writes v, or the right error.
func respond(w http.ResponseWriter, status int, message, what string, v any, err error) {
	if err == nil {
		httpx.OK(w, status, message, v)
		return
	}
	// A weight outside tolerance also says by how much, for the form to show.
	var off ToleranceError
	if errors.As(err, &off) {
		httpx.ErrorData(w, http.StatusConflict, err.Error(), map[string]any{
			"difference": off.Difference(), "tolerance": off.Tolerance,
			"sent": off.Sent, "received": off.Received,
		})
		return
	}
	// Refusals that carry numbers: their message says how far off the request was.
	var (
		capacity     CapacityError
		insufficient InsufficientError
		yield        YieldError
		evidence     EvidenceError
	)
	if errors.As(err, &capacity) || errors.As(err, &insufficient) || errors.As(err, &yield) || errors.As(err, &evidence) {
		httpx.Error(w, http.StatusConflict, err.Error())
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

// ruleOf names the rule a refusal comes from, "" if it is not one the
// blocked report counts.
func ruleOf(err error) string {
	var (
		capacity CapacityError
		yield    YieldError
		evidence EvidenceError
		off      ToleranceError
	)
	switch {
	case errors.As(err, &capacity):
		return audit.RuleCapacity
	case errors.As(err, &yield):
		return audit.RuleMassBalance
	case errors.As(err, &evidence):
		return audit.RuleEvidence
	case errors.As(err, &off):
		return audit.RuleTolerance
	}
	return ""
}

// countRefusal records a rule's refusal for the blocked report. It goes
// through the pool: the refused action's own transaction was rolled back.
func (h *Handlers) countRefusal(r *http.Request, orgID string, err error) {
	if rule := ruleOf(err); rule != "" {
		if err := audit.RecordRule(r.Context(), h.Store.Pool, rule, orgID); err != nil {
			logx.From(r.Context()).Error("record rule refusal failed", "rule", rule, logx.KeyError, err)
		}
	}
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

// signer is acting for a signing action: it also checks the signing PIN.
func (h *Handlers) signer(w http.ResponseWriter, r *http.Request) (orgID, userID string, ok bool) {
	if orgID, userID, ok = acting(w, r); ok {
		ok = h.PIN(w, r)
	}
	return orgID, userID, ok
}

// Issue creates an origin passport for the acting organisation.
func (h *Handlers) Issue(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.signer(w, r)
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
	h.countRefusal(r, orgID, err)
	respond(w, http.StatusCreated, MsgIssued, "issue passport", p, err)
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
	respond(w, http.StatusOK, "", "list passports", page, err)
}

// target reads the gate and the {id} a request is about.
func target(w http.ResponseWriter, r *http.Request) (id, gate string, ok bool) {
	if gate, ok = auth.Gate(w, r); !ok {
		return "", "", false
	}
	id, ok = httpx.URLID(w, r)
	return id, gate, ok
}

// Get takes the passport's id or its lot code.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	id, err := h.Store.Resolve(r.Context(), chi.URLParam(r, httpx.ParamID))
	if err != nil {
		respond(w, http.StatusOK, "", "find passport", nil, err)
		return
	}
	p, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "", "get passport", p, err)
}

// Retire closes a passport the acting organisation holds: the lot reached
// its final buyer.
func (h *Handlers) Retire(w http.ResponseWriter, r *http.Request) {
	orgID, _, ok := h.signer(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	p, err := h.Store.Retire(r.Context(), id, orgID)
	respond(w, http.StatusOK, MsgRetired, "retire passport", p, err)
}

// Transform processes, splits or merges passports the acting organisation holds.
func (h *Handlers) Transform(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.signer(w, r)
	if !ok {
		return
	}
	var in TransformInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	t, err := h.Store.Transform(r.Context(), orgID, userID, in, h.now(), h.RaiseFlag)
	h.countRefusal(r, orgID, err)
	respond(w, http.StatusCreated, MsgTransformed, "transform", t, err)
}

// Genealogy returns a passport's ancestors (the default) or descendants.
func (h *Handlers) Genealogy(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	direction := r.URL.Query().Get(ParamDirection)
	if direction == "" {
		direction = DirectionAncestors
	}
	if !slices.Contains(Directions, direction) {
		listq.BadParam(w, ParamDirection, Directions)
		return
	}
	g, err := h.Store.Genealogy(r.Context(), id, gate, direction)
	respond(w, http.StatusOK, "", "genealogy", g, err)
}
