package facility

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

type Handlers struct {
	Store Store
	// Files holds licence documents.
	Files *upload.Bucket
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
func respond(w http.ResponseWriter, status int, message, what string, f Facility, err error) {
	var overlap OverlapError
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrBadGeometry):
		httpx.Error(w, http.StatusBadRequest, MsgBadGeometry)
	case errors.Is(err, ErrNotPending):
		httpx.Error(w, http.StatusConflict, MsgNotPending)
	case errors.Is(err, ErrPending):
		httpx.Error(w, http.StatusConflict, MsgPending)
	case errors.Is(err, ErrNoLicence):
		httpx.Error(w, http.StatusNotFound, MsgNoLicenceFile)
	case errors.As(err, &overlap):
		var id *string // null unless the facility is the caller's own
		if overlap.FacilityID != "" {
			id = &overlap.FacilityID
		}
		httpx.ErrorData(w, http.StatusConflict, overlap.Error(), map[string]any{"facility_id": id})
	case err != nil:
		httpx.ServerError(w, what, err)
	default:
		httpx.OK(w, status, message, f)
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
	if errors.Is(err, ErrUnknownKind) {
		httpx.Error(w, http.StatusBadRequest, h.unknownKind(r.Context()))
		return
	}
	respond(w, http.StatusCreated, MsgRegistered(in.Name), "create facility", f, err)
}

// unknownKind names the active kinds, for a facility whose kind is retired or invented.
func (h *Handlers) unknownKind(ctx context.Context) string {
	kinds, err := h.Store.ActiveKindKeys(ctx)
	if err != nil || len(kinds) == 0 {
		return MsgBadKind
	}
	return MsgUnknownKind(kinds)
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
	kind := r.URL.Query().Get(ParamKind)
	if kind != "" {
		kinds, err := h.Store.AllKindKeys(r.Context())
		if err != nil {
			httpx.ServerError(w, "list kinds", err)
			return
		}
		if !slices.Contains(kinds, kind) {
			listq.BadParam(w, ParamKind, kinds)
			return
		}
	}
	bbox, ok := ParseBBox(r.URL.Query().Get(ParamBBox))
	if !ok {
		httpx.Error(w, http.StatusBadRequest, MsgBadBBox)
		return
	}
	page, err := h.Store.Find(r.Context(), q, gate, status, kind, bbox)
	if err != nil {
		httpx.ServerError(w, "list facilities", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", page)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "", "get facility", f, err)
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
	respond(w, http.StatusOK, statusMessage[in.Status], "set facility status", f, err)
}

// Approve makes a pending facility active and puts it on the ledger: the platform only.
func (h *Handlers) Approve(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Approve(r.Context(), id, auth.CurrentUser(r.Context()).ID)
	respond(w, http.StatusOK, MsgApproved, "approve facility", f, err)
}

// UploadLicence stores the facility's licence document, replacing any earlier
// one. A renewed licence also sends its new end date (form field licence_valid_to).
func (h *Handlers) UploadLicence(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	file, header, ok := upload.Parse(w, r)
	if !ok {
		return
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()

	validTo := strings.TrimSpace(r.FormValue(FormLicenceValidTo))
	if _, err := time.Parse(DateLayout, validTo); validTo != "" && err != nil {
		httpx.Error(w, http.StatusBadRequest, MsgLicenceDate)
		return
	}
	u, ok := upload.Store(w, r, h.Files, file, header)
	if !ok {
		return
	}
	f, old, err := h.Store.SetLicence(r.Context(), id, gate, Licence{
		FileName: u.Name, ContentType: u.ContentType, SizeBytes: u.Size, SHA256: u.SHA256, StorageKey: u.Key,
	}, validTo)
	if err != nil {
		h.Files.Remove(r.Context(), u.Key) // do not keep a file with no record
		respond(w, 0, "", "record licence", f, err)
		return
	}
	if old != "" {
		h.Files.Remove(r.Context(), old)
	}
	respond(w, http.StatusOK, MsgLicenceSaved, "", f, nil)
}

// Licence streams the facility's licence document.
func (h *Handlers) Licence(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Get(r.Context(), id, gate)
	if err == nil && f.LicenceDocument == nil {
		err = ErrNoLicence
	}
	if err != nil {
		respond(w, 0, "", "get licence", f, err)
		return
	}
	l := f.LicenceDocument
	upload.Serve(w, r, h.Files, l.StorageKey, l.FileName, l.ContentType, f.UpdatedAt)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	if err := h.Store.Delete(r.Context(), id, gate); err != nil {
		respond(w, 0, "", "delete facility", Facility{}, err)
		return
	}
	httpx.OK(w, http.StatusOK, MsgDeleted, nil)
}

// KindInput is what the platform sends to create or change a facility kind.
type KindInput struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Sub    string `json:"sub"`
	Active bool   `json:"active"`
}

func (in *KindInput) normalise() {
	in.Key = strings.TrimSpace(strings.ToLower(in.Key))
	in.Label = strings.TrimSpace(in.Label)
	in.Sub = strings.TrimSpace(in.Sub)
}

// Kinds lists active facility kinds for forms and filters.
func (h *Handlers) Kinds(w http.ResponseWriter, r *http.Request) {
	kinds, err := h.Store.Kinds(r.Context())
	if err != nil {
		httpx.ServerError(w, "list kinds", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", map[string]any{"data": kinds})
}

// AllKinds includes deactivated kinds for the platform management screen.
func (h *Handlers) AllKinds(w http.ResponseWriter, r *http.Request) {
	kinds, err := h.Store.AllKinds(r.Context())
	if err != nil {
		httpx.ServerError(w, "list all kinds", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", map[string]any{"data": kinds})
}

// CreateKind adds a facility kind. Only the platform curates kinds.
func (h *Handlers) CreateKind(w http.ResponseWriter, r *http.Request) {
	var in KindInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.normalise()
	if msg := ValidateKind(in.Key, in.Label, in.Sub, true); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	k, err := h.Store.CreateKind(r.Context(), Kind{Key: in.Key, Label: in.Label, Sub: in.Sub, Active: true})
	if errors.Is(err, ErrKindTaken) {
		httpx.Error(w, http.StatusConflict, MsgKindTaken)
		return
	}
	if err != nil {
		httpx.ServerError(w, "create kind", err)
		return
	}
	httpx.OK(w, http.StatusCreated, MsgKindCreated, k)
}

// UpdateKind changes a kind's label, description or availability.
// Deactivating hides it from forms; what already uses it keeps working.
func (h *Handlers) UpdateKind(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, ParamKindKey)
	if !keyPattern.MatchString(key) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	var in KindInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.normalise()
	if msg := ValidateKind("", in.Label, in.Sub, false); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	k, err := h.Store.UpdateKind(r.Context(), key, Kind{Label: in.Label, Sub: in.Sub, Active: in.Active})
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if err != nil {
		httpx.ServerError(w, "update kind", err)
		return
	}
	httpx.OK(w, http.StatusOK, MsgKindUpdated, k)
}

// DeleteKind removes a kind nobody uses. Facilities and templates keep
// referencing their kinds, so an in-use kind is refused, not cascade-deleted.
func (h *Handlers) DeleteKind(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, ParamKindKey)
	if !keyPattern.MatchString(key) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	err := h.Store.DeleteKind(r.Context(), key)
	switch {
	case errors.Is(err, ErrKindInUse):
		httpx.Error(w, http.StatusConflict, MsgKindInUse)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case err != nil:
		httpx.ServerError(w, "delete kind", err)
	default:
		httpx.OK(w, http.StatusOK, MsgKindDeleted, nil)
	}
}
