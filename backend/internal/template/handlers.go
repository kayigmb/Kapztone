package template

import (
	"errors"
	"net/http"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

// Templates are platform-wide, so there is no organisation gate here:
// permissions alone decide who reads and who writes.
type Handlers struct {
	Store Store
}

func respond(w http.ResponseWriter, status int, what string, t Template, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrCommodityTaken):
		httpx.Error(w, http.StatusConflict, MsgCommodityTaken)
	case errors.Is(err, ErrInUse):
		httpx.Error(w, http.StatusConflict, MsgInUse)
	case err != nil:
		httpx.ServerError(w, what, err)
	default:
		httpx.WriteJSON(w, status, t)
	}
}

// input reads and validates a template body, or writes 400.
func input(w http.ResponseWriter, r *http.Request) (Input, bool) {
	var in Input
	if !httpx.DecodeJSON(w, r, &in) {
		return in, false
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return in, false
	}
	return in, true
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := input(w, r)
	if !ok {
		return
	}
	t, err := h.Store.Create(r.Context(), in)
	respond(w, http.StatusCreated, "create template", t, err)
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q, ok := listq.FromRequest(w, r, List)
	if !ok {
		return
	}
	page, err := h.Store.Find(r.Context(), q)
	if err != nil {
		httpx.ServerError(w, "list templates", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	t, err := h.Store.Get(r.Context(), id)
	respond(w, http.StatusOK, "get template", t, err)
}

// Update replaces the whole template.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	in, ok := input(w, r)
	if !ok {
		return
	}
	t, err := h.Store.Update(r.Context(), id, in)
	respond(w, http.StatusOK, "update template", t, err)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	if err := h.Store.Delete(r.Context(), id); err != nil {
		respond(w, 0, "delete template", Template{}, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
