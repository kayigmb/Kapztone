package facility

import (
	"errors"
	"net/http"
	"slices"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

type Handlers struct {
	Store Store
}

// target reads the gate and the {id} of the facility a request is about.
func target(w http.ResponseWriter, r *http.Request) (id, gate string, ok bool) {
	if gate, ok = auth.Gate(w, r); !ok {
		return "", "", false
	}
	id, ok = httpx.URLID(w, r)
	return id, gate, ok
}

// respond writes a facility, or the right error. A facility outside the
// caller's gate looks exactly like a missing one.
func respond(w http.ResponseWriter, status int, what string, f Facility, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrBadGeometry):
		httpx.Error(w, http.StatusBadRequest, MsgBadGeometry)
	case err != nil:
		httpx.ServerError(w, what, err)
	default:
		httpx.WriteJSON(w, status, f)
	}
}

// Create registers a facility for the acting organisation. The route requires
// that organisation to be verified.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	o := auth.CurrentScope(r.Context()).Organisation
	if o == nil {
		httpx.Error(w, http.StatusForbidden, auth.MsgNoOrgSelected)
		return
	}
	f, err := h.Store.Create(r.Context(), o.ID, in)
	respond(w, http.StatusCreated, "create facility", f, err)
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
	bbox, ok := ParseBBox(r.URL.Query().Get(ParamBBox))
	if !ok {
		httpx.Error(w, http.StatusBadRequest, MsgBadBBox)
		return
	}
	page, err := h.Store.Find(r.Context(), q, gate, status, bbox)
	if err != nil {
		httpx.ServerError(w, "list facilities", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "get facility", f, err)
}

// SetStatus suspends or reactivates a facility.
func (h *Handlers) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	var in StatusInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	f, err := h.Store.SetStatus(r.Context(), id, gate, in.Status)
	respond(w, http.StatusOK, "set facility status", f, err)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	if err := h.Store.Delete(r.Context(), id, gate); err != nil {
		respond(w, 0, "delete facility", Facility{}, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
