package passport

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

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
	respond(w, http.StatusCreated, MsgHandoverStarted, "start handover", out, err)
}

func parseInt(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) }

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
	return h.storePhotos(w, r, r.MultipartForm.File[FormPhotos])
}

// storePhotos keeps every photo or none: on any refusal the ones already stored are removed.
func (h *Handlers) storePhotos(w http.ResponseWriter, r *http.Request, headers []*multipart.FileHeader) ([]upload.File, bool) {
	if len(headers) > MaxPhotos {
		httpx.Error(w, http.StatusBadRequest, fmt.Sprintf("attach at most %d photos", MaxPhotos))
		return nil, false
	}
	var stored []upload.File
	for _, header := range headers {
		f, ok := h.storePhoto(w, r, header)
		if !ok {
			h.dropPhotos(r, stored)
			return nil, false
		}
		stored = append(stored, f)
	}
	return stored, true
}

func (h *Handlers) storePhoto(w http.ResponseWriter, r *http.Request, header *multipart.FileHeader) (upload.File, bool) {
	if header.Size > upload.MaxBytes {
		httpx.Error(w, http.StatusRequestEntityTooLarge, upload.MsgFileTooLarge)
		return upload.File{}, false
	}
	file, err := header.Open()
	if err != nil {
		httpx.ServerError(w, "open photo", err)
		return upload.File{}, false
	}
	f, ok := upload.Store(w, r, h.Files, file, header)
	file.Close()
	if ok && !slices.Contains(PhotoTypes, f.ContentType) {
		h.dropPhotos(r, []upload.File{f})
		httpx.Error(w, http.StatusBadRequest, MsgPhotoType)
		return upload.File{}, false
	}
	return f, ok
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
		respond(w, http.StatusOK, "", "get handover photo", nil, err)
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
	box := r.URL.Query().Get(ParamBox)
	if box != "" && !slices.Contains(Boxes, box) {
		listq.BadParam(w, ParamBox, Boxes)
		return
	}
	page, err := h.Store.FindHandovers(r.Context(), q, gate, status, box)
	respond(w, http.StatusOK, "", "list handovers", page, err)
}

func (h *Handlers) GetHandover(w http.ResponseWriter, r *http.Request) {
	id, gate, ok := target(w, r)
	if !ok {
		return
	}
	out, err := h.Store.GetHandover(r.Context(), id, gate)
	respond(w, http.StatusOK, "", "get handover", out, err)
}

// answer runs one of the three ways an awaiting handover ends. Accepting and
// rejecting are signed by the receiver; cancelling is not a signature.
func (h *Handlers) answer(w http.ResponseWriter, r *http.Request, message, what string, signed bool,
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
	respond(w, http.StatusOK, message, what, out, err)
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
	respond(w, http.StatusOK, MsgHandoverAccepted, "accept handover", out, err)
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
	h.answer(w, r, MsgHandoverRejected, "reject handover", true, func(id, orgID, userID string) (Handover, error) {
		return h.Store.Reject(r.Context(), id, orgID, userID, in, h.RaiseFlag)
	})
}

// Cancel is the sender withdrawing a handover nobody has answered yet.
func (h *Handlers) Cancel(w http.ResponseWriter, r *http.Request) {
	h.answer(w, r, MsgHandoverCancelled, "cancel handover", false, func(id, orgID, userID string) (Handover, error) {
		return h.Store.Cancel(r.Context(), id, orgID, userID)
	})
}
