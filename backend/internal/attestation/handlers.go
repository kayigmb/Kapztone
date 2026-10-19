package attestation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

type Handlers struct {
	Store Store
	// Files holds certificate files.
	Files *upload.Bucket
	// PIN checks the signing PIN: issuing and revoking are signed (auth.Auth.CheckPIN).
	PIN func(http.ResponseWriter, *http.Request) bool
	// Pin copies an issued certificate's file to IPFS and returns its
	// content id. nil means files are kept in the bucket only.
	Pin PinFunc
}

// Input is a certificate's details, sent as form fields beside the file.
type Input struct {
	SubjectType string
	SubjectID   string
	Type        string // an evidence key such as "export_permit"
	ValidFrom   time.Time
	ValidTo     time.Time
}

// ParseInput reads and checks the form fields. get returns one field's
// value. It returns the first problem as a message that is safe to show, or "".
func ParseInput(get func(string) string) (Input, string) {
	in := Input{SubjectType: get(FormSubjectType), SubjectID: get(FormSubjectID), Type: strings.TrimSpace(get(FormType))}
	from, okFrom := httpx.ParseTime(get(FormValidFrom))
	to, okTo := httpx.ParseTime(get(FormValidTo))
	in.ValidFrom, in.ValidTo = from, to
	switch {
	case !slices.Contains(SubjectTypes, in.SubjectType):
		return in, FormSubjectType + " must be one of: " + strings.Join(SubjectTypes, ", ")
	case !httpx.IsUUID(in.SubjectID):
		return in, FormSubjectID + " is required"
	case !template.ValidKey(in.Type):
		return in, FormType + " must be " + template.KeyRule
	case !okFrom || !okTo || !to.After(from) || from.Unix() < 0: // the chain holds times as unsigned seconds
		return in, MsgDates
	}
	return in, ""
}

func respond(w http.ResponseWriter, status int, message, what string, v any, err error) {
	switch {
	case err == nil:
		httpx.OK(w, status, message, v)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case errors.Is(err, ErrNotAccredited):
		httpx.Error(w, http.StatusForbidden, MsgNotAccredited)
	case errors.Is(err, ErrSubjectNotFound):
		httpx.Error(w, http.StatusBadRequest, MsgSubjectNotFound)
	case errors.Is(err, ErrAlreadyRevoked):
		httpx.Error(w, http.StatusConflict, MsgAlreadyRevoked)
	default:
		httpx.ServerError(w, what, err)
	}
}

// Issue records a certificate for the acting, accredited organisation.
func (h *Handlers) Issue(w http.ResponseWriter, r *http.Request) {
	o := auth.CurrentScope(r.Context()).Organisation
	if o == nil {
		httpx.Error(w, http.StatusForbidden, auth.MsgNoOrgSelected)
		return
	}
	if !h.PIN(w, r) {
		return
	}
	file, header, ok := upload.Parse(w, r)
	if !ok {
		return
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()

	in, msg := ParseInput(r.FormValue)
	if msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	f, ok := upload.Store(w, r, h.Files, file, header)
	if !ok {
		return
	}
	a, err := h.Store.Issue(r.Context(), o.ID, auth.CurrentUser(r.Context()).ID, in, f)
	if err != nil {
		h.Files.Remove(r.Context(), f.Key) // do not keep a file with no record
		respond(w, 0, "", "issue certificate", nil, err)
		return
	}
	// Pinned only once the certificate is recorded, so a refused file is
	// never published. A failed pin does not undo the certificate: the hash
	// on-chain is the proof; IPFS only makes the file available elsewhere.
	if h.Pin != nil {
		if pinned, err := h.pin(r.Context(), a); err != nil {
			logx.From(r.Context()).Warn("ipfs pin failed; will retry", "attestation_id", a.ID, logx.KeyError, err)
			if err := h.Store.RecordPinFailure(r.Context(), a.ID, time.Now(), err); err != nil {
				logx.From(r.Context()).Error("ipfs retry not scheduled", "attestation_id", a.ID, logx.KeyError, err)
			}
		} else {
			a = pinned
		}
	}
	httpx.OK(w, http.StatusCreated, MsgIssued, a)
}

func (h *Handlers) pin(ctx context.Context, a Attestation) (Attestation, error) {
	f, err := h.Files.Open(ctx, a.StorageKey)
	if err != nil {
		return a, err
	}
	defer f.Close()
	cid, err := h.Pin(ctx, a.FileName, f)
	if err != nil {
		return a, err
	}
	return h.Store.SetCID(ctx, a.ID, cid)
}

// List shows an organisation the certificates it issued and those about its
// own lots and facilities; the platform sees every certificate. The public
// verification page shows a lot's certificates to anyone.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, List)
	if !ok {
		return
	}
	scope := r.URL.Query().Get(ParamScope)
	if scope != "" && !slices.Contains(Scopes, scope) {
		listq.BadParam(w, ParamScope, Scopes)
		return
	}
	status := r.URL.Query().Get(ParamStatus)
	if status != "" && !slices.Contains(Statuses, status) {
		listq.BadParam(w, ParamStatus, Statuses)
		return
	}
	subjectID := r.URL.Query().Get(ParamSubjectID)
	if subjectID != "" && !httpx.IsUUID(subjectID) {
		httpx.Error(w, http.StatusBadRequest, ParamSubjectID+" must be an id")
		return
	}
	issuerID := r.URL.Query().Get(ParamIssuerID)
	if issuerID != "" && !httpx.IsUUID(issuerID) {
		httpx.Error(w, http.StatusBadRequest, ParamIssuerID+" must be an id")
		return
	}
	page, err := h.Store.Find(r.Context(), q, Filter{Gate: gate, Scope: scope, SubjectID: subjectID, IssuerID: issuerID, Status: status})
	respond(w, http.StatusOK, "", "list certificates", page, err)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	a, err := h.Store.Get(r.Context(), id)
	respond(w, http.StatusOK, "", "get certificate", a, err)
}

// Download streams the certificate's file.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	a, err := h.Store.Get(r.Context(), id)
	if err != nil {
		respond(w, 0, "", "get certificate", nil, err)
		return
	}
	upload.Serve(w, r, h.Files, a.StorageKey, a.FileName, a.ContentType, a.CreatedAt)
}

// RevokeInput is the issuer's reason for withdrawing a certificate.
type RevokeInput struct {
	Reason string `json:"reason"`
}

func (in *RevokeInput) Validate() string {
	in.Reason = strings.TrimSpace(in.Reason)
	switch {
	case in.Reason == "":
		return MsgReasonNeeded
	case len(in.Reason) > MaxReasonLen:
		return fmt.Sprintf("reason may be up to %d characters", MaxReasonLen)
	}
	return ""
}

func (h *Handlers) Revoke(w http.ResponseWriter, r *http.Request) {
	gate, ok := auth.Gate(w, r)
	if !ok || !h.PIN(w, r) {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	var in RevokeInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	a, err := h.Store.Revoke(r.Context(), id, gate, auth.CurrentUser(r.Context()).ID, in.Reason)
	respond(w, http.StatusOK, MsgRevoked, "revoke certificate", a, err)
}
