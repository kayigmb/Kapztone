package facility

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

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
func respond(w http.ResponseWriter, status int, what string, f Facility, err error) {
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
		httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": overlap.Error(), "facility_id": id})
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

// Approve makes a pending facility active and puts it on the ledger: the platform only.
func (h *Handlers) Approve(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	f, err := h.Store.Approve(r.Context(), id, auth.CurrentUser(r.Context()).ID)
	respond(w, http.StatusOK, "approve facility", f, err)
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
		respond(w, 0, "record licence", f, err)
		return
	}
	if old != "" {
		h.Files.Remove(r.Context(), old)
	}
	respond(w, http.StatusOK, "", f, nil)
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
		respond(w, 0, "get licence", f, err)
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
		respond(w, 0, "delete facility", Facility{}, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
