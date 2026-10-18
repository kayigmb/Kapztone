package passport

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
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
func respond(w http.ResponseWriter, status int, what string, v any, err error) {
	if err == nil {
		httpx.WriteJSON(w, status, v)
		return
	}
	// A weight outside tolerance also says by how much, for the form to show.
	var off ToleranceError
	if errors.As(err, &off) {
		httpx.WriteJSON(w, http.StatusConflict, map[string]any{
			"error": err.Error(), "difference": off.Difference(), "tolerance": off.Tolerance,
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
			log.Printf("record %s refusal: %v", rule, err)
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

// Get takes the passport's id or its lot code.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	id, err := h.Store.Resolve(r.Context(), chi.URLParam(r, httpx.ParamID))
	if err != nil {
		respond(w, http.StatusOK, "find passport", nil, err)
		return
	}
	p, err := h.Store.Get(r.Context(), id, gate)
	respond(w, http.StatusOK, "get passport", p, err)
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
	respond(w, http.StatusOK, "retire passport", p, err)
}

// StartHandover is the sender's half of a handover. It takes JSON, or a
// multipart form when photos come with it.
func (h *Handlers) StartHandover(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.signer(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in HandoverInput
	photos, ok := h.readBody(w, r, &in, func(get func(string) string) (err error) {
		in.ToOrg, in.Stage, in.Note = get(FieldToOrg), get(FieldStage), get(FieldNote)
		if in.Quantity, err = optional(get(FieldQuantity), parseInt); err != nil {
			return err
		}
		if in.Lat, err = optional(get(FieldLat), parseFloat); err != nil {
			return err
		}
		in.Lon, err = optional(get(FieldLon), parseFloat)
		return err
	})
	if !ok {
		return
	}
	if msg := in.Validate(); msg != "" {
		h.dropPhotos(r, photos)
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	out, err := h.Store.StartHandover(r.Context(), id, orgID, userID, in, photos, h.now())
	if err != nil {
		h.dropPhotos(r, photos)
	}
	h.countRefusal(r, orgID, err)
	respond(w, http.StatusCreated, "start handover", out, err)
}

func parseInt(s string) (int64, error)     { return strconv.ParseInt(s, 10, 64) }
func parseFloat(s string) (float64, error) { return strconv.ParseFloat(s, 64) }

// optional parses a form field that may be left empty.
func optional[T any](s string, parse func(string) (T, error)) (*T, error) {
	if s = strings.TrimSpace(s); s == "" {
		return nil, nil
	}
	v, err := parse(s)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// readBody decodes a JSON body into dst, or, for a multipart form, lets fill
// read its text fields and stores its photos in the bucket. It writes any
// refusal itself. Stored photos must be dropped if the request then fails.
func (h *Handlers) readBody(w http.ResponseWriter, r *http.Request, dst any, fill func(get func(string) string) error) ([]upload.File, bool) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return nil, httpx.DecodeJSON(w, r, dst)
	}
	// Room for every photo at its largest, plus the form around them.
	r.Body = http.MaxBytesReader(w, r.Body, MaxPhotos*upload.MaxBytes+1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			httpx.Error(w, http.StatusRequestEntityTooLarge, upload.MsgFileTooLarge)
		} else {
			httpx.Error(w, http.StatusBadRequest, MsgBadForm)
		}
		return nil, false
	}
	defer r.MultipartForm.RemoveAll()
	if err := fill(r.FormValue); err != nil {
		httpx.Error(w, http.StatusBadRequest, MsgNumbers)
		return nil, false
	}
	headers := r.MultipartForm.File[FormPhotos]
	if len(headers) > MaxPhotos {
		httpx.Error(w, http.StatusBadRequest, fmt.Sprintf("attach at most %d photos", MaxPhotos))
		return nil, false
	}
	var stored []upload.File
	for _, header := range headers {
		if header.Size > upload.MaxBytes {
			h.dropPhotos(r, stored)
			httpx.Error(w, http.StatusRequestEntityTooLarge, upload.MsgFileTooLarge)
			return nil, false
		}
		file, err := header.Open()
		if err != nil {
			h.dropPhotos(r, stored)
			httpx.ServerError(w, "open photo", err)
			return nil, false
		}
		f, ok := upload.Store(w, r, h.Files, file, header)
		file.Close()
		if !ok {
			h.dropPhotos(r, stored)
			return nil, false
		}
		stored = append(stored, f)
		if !slices.Contains(PhotoTypes, f.ContentType) {
			h.dropPhotos(r, stored)
			httpx.Error(w, http.StatusBadRequest, MsgPhotoType)
			return nil, false
		}
	}
	return stored, true
}

// dropPhotos removes stored photos whose handover was never recorded.
func (h *Handlers) dropPhotos(r *http.Request, photos []upload.File) {
	for _, f := range photos {
		h.Files.Remove(r.Context(), f.Key)
	}
}

// HandoverPhoto streams one photo of a handover to its sender or receiver.
func (h *Handlers) HandoverPhoto(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	photoID := chi.URLParam(r, ParamPhotoID)
	if !httpx.IsUUID(photoID) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	p, err := h.Store.HandoverPhoto(r.Context(), id, photoID, gate)
	if err != nil {
		respond(w, http.StatusOK, "get handover photo", nil, err)
		return
	}
	upload.Serve(w, r, h.Files, p.StorageKey, p.FileName, p.ContentType, p.CreatedAt)
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

// answer runs one of the three ways an awaiting handover ends. Accepting and
// rejecting are signed by the receiver; cancelling is not a signature.
func (h *Handlers) answer(w http.ResponseWriter, r *http.Request, what string, signed bool,
	do func(id, orgID, userID string) (Handover, error)) {

	orgID, userID, ok := acting(w, r)
	if ok && signed {
		ok = h.PIN(w, r)
	}
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

// Accept is the receiver's half: they record what they weighed, sign, and the
// passport changes holder. JSON, or a multipart form with photos.
func (h *Handlers) Accept(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.signer(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in AcceptInput
	photos, ok := h.readBody(w, r, &in, func(get func(string) string) (err error) {
		if in.ReceivedQuantity, err = optional(get(FieldReceivedQuantity), parseInt); err != nil {
			return err
		}
		if in.Lat, err = optional(get(FieldLat), parseFloat); err != nil {
			return err
		}
		in.Lon, err = optional(get(FieldLon), parseFloat)
		return err
	})
	if !ok {
		return
	}
	if msg := in.Validate(); msg != "" {
		h.dropPhotos(r, photos)
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	out, err := h.Store.Accept(r.Context(), id, orgID, userID, in, photos, h.RaiseFlag)
	if err != nil {
		h.dropPhotos(r, photos)
	}
	h.countRefusal(r, orgID, err)
	respond(w, http.StatusOK, "accept handover", out, err)
}

// Reject is the receiver refusing the goods, saying why (and maybe flagging the passport).
func (h *Handlers) Reject(w http.ResponseWriter, r *http.Request) {
	var in RejectInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	h.answer(w, r, "reject handover", true, func(id, orgID, userID string) (Handover, error) {
		return h.Store.Reject(r.Context(), id, orgID, userID, in, h.RaiseFlag)
	})
}

// Cancel is the sender withdrawing a handover nobody has answered yet.
func (h *Handlers) Cancel(w http.ResponseWriter, r *http.Request) {
	h.answer(w, r, "cancel handover", false, func(id, orgID, userID string) (Handover, error) {
		return h.Store.Cancel(r.Context(), id, orgID, userID)
	})
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
	respond(w, http.StatusCreated, "transform", t, err)
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
	respond(w, http.StatusOK, "genealogy", g, err)
}
