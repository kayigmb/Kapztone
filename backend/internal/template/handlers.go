package template

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

// Templates are platform-wide, so there is no organisation gate here:
// permissions alone decide who reads and who writes.
type Handlers struct {
	Store Store
}

func respond(w http.ResponseWriter, status int, message, what string, t Template, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrCommodityTaken):
		httpx.Error(w, http.StatusConflict, MsgCommodityTaken)
	case errors.Is(err, ErrInUse):
		httpx.Error(w, http.StatusConflict, MsgInUse)
	case errors.Is(err, ErrNoDraft):
		httpx.Error(w, http.StatusConflict, MsgNoDraft)
	case err != nil:
		httpx.ServerError(w, what, err)
	default:
		httpx.OK(w, status, message, t)
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
	if errors.Is(err, ErrUnknownKind) {
		httpx.Error(w, http.StatusBadRequest, h.unknownKind(r.Context()))
		return
	}
	respond(w, http.StatusCreated, MsgCreated, "create template", t, err)
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
	httpx.OK(w, http.StatusOK, "", page)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	t, err := h.Store.Get(r.Context(), id)
	respond(w, http.StatusOK, "", "get template", t, err)
}

// Update saves a whole new version of the template as its draft.
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
	if errors.Is(err, ErrUnknownKind) {
		httpx.Error(w, http.StatusBadRequest, h.unknownKind(r.Context()))
		return
	}
	respond(w, http.StatusOK, MsgDraftSaved, "update template", t, err)
}

// unknownKind names the active origin kinds, for a template naming one that
// is retired or invented. The store guards every write path; this only
// formats the refusal.
func (h *Handlers) unknownKind(ctx context.Context) string {
	kinds, err := h.Store.ActiveKinds(ctx)
	if err != nil {
		return MsgUnknownKind
	}
	names := make([]string, 0, len(kinds))
	for _, k := range kinds {
		names = append(names, k.Key)
	}
	if len(names) == 0 {
		return MsgUnknownKind
	}
	return "originFacilityKind must be one of: " + strings.Join(names, ", ")
}

// Publish makes the draft the template's rules, as the next version.
func (h *Handlers) Publish(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	t, err := h.Store.Publish(r.Context(), id)
	if errors.Is(err, ErrUnknownKind) {
		httpx.Error(w, http.StatusBadRequest, h.unknownKind(r.Context()))
		return
	}
	respond(w, http.StatusOK, MsgPublished, "publish template", t, err)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	if err := h.Store.Delete(r.Context(), id); err != nil {
		respond(w, 0, "", "delete template", Template{}, err)
		return
	}
	httpx.OK(w, http.StatusOK, MsgDeleted, nil)
}
