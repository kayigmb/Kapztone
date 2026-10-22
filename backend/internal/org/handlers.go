package org

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

type Handlers struct {
	Store Store
	// Files holds KYB documents.
	Files *upload.Bucket
	// InvalidateOrg clears cached access data after an organisation changes.
	InvalidateOrg func(ctx context.Context, slug string)
	// Secrets seals each organisation's signing key before it is stored.
	Secrets *secret.Box
	// SendVerify emails a new account the link that confirms its email; nil sends nothing.
	// It returns the link in dev only (see auth.SendVerifyTo).
	SendVerify func(ctx context.Context, email string) (devLink string)
}

// newSigningKey makes the key an organisation receives when first verified.
func (h *Handlers) newSigningKey() (NewKey, error) {
	hexKey, address, err := chain.NewKey()
	if err != nil {
		return NewKey{}, err
	}
	sealed, err := h.Secrets.Seal(hexKey)
	return NewKey{Address: address.Hex(), Sealed: sealed}, err
}

// visibleOrg loads the organisation named in the URL if the caller may see it.
// An organisation outside the caller's gate looks exactly like a missing one.
func (h *Handlers) visibleOrg(w http.ResponseWriter, r *http.Request) (Organisation, bool) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return Organisation{}, false
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return Organisation{}, false
	}
	o, err := h.Store.Get(r.Context(), id, g)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return Organisation{}, false
	}
	if err != nil {
		httpx.ServerError(w, "get organisation", err)
		return Organisation{}, false
	}
	return o, true
}

// Register is public: a new business signs up with its first user.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var in Registration
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if msg := in.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		httpx.ServerError(w, "hash password", err)
		return
	}
	o, err := h.Store.Register(r.Context(), in, hash)
	switch {
	case errors.Is(err, ErrSlugTaken):
		httpx.Error(w, http.StatusConflict, MsgSlugTaken)
	case errors.Is(err, ErrEmailTaken):
		httpx.Error(w, http.StatusConflict, MsgEmailTaken)
	case err != nil:
		httpx.ServerError(w, "register organisation", err)
	default:
		var link string
		if h.SendVerify != nil {
			link = h.SendVerify(r.Context(), in.Email)
		}
		httpx.OK(w, http.StatusCreated, MsgRegistered, struct {
			Organisation
			VerifyLink string `json:"verify_link,omitempty"` // dev only
		}{o, link})
	}
}

// EmailAvailability lets sign-up catch a duplicate before the final submit.
func (h *Handlers) EmailAvailability(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("email")))
	if !validEmail(email) {
		httpx.Error(w, http.StatusBadRequest, "email is not a valid address")
		return
	}
	available, err := h.Store.EmailAvailable(r.Context(), email)
	if err != nil {
		httpx.ServerError(w, "check email availability", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", map[string]bool{"available": available})
}

// listOf reads a comma-separated filter whose every value must be allowed; none means no filter.
func listOf(w http.ResponseWriter, r *http.Request, param string, allowed []string) ([]string, bool) {
	out := []string{} // never nil: a NULL array would match nothing
	for v := range strings.SplitSeq(r.URL.Query().Get(param), ",") {
		if v = strings.TrimSpace(v); v == "" {
			continue
		}
		if !slices.Contains(allowed, v) {
			listq.BadParam(w, param, allowed)
			return nil, false
		}
		out = append(out, v)
	}
	return out, true
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, OrgList)
	if !ok {
		return
	}
	// Both filters take several values, comma-separated: ?status=rejected,changes_requested.
	status, ok := listOf(w, r, ParamStatus, Statuses)
	if !ok {
		return
	}
	kinds, ok := listOf(w, r, ParamKind, Kinds)
	if !ok {
		return
	}
	page, err := h.Store.List(r.Context(), q, g, status, kinds)
	if err != nil {
		httpx.ServerError(w, "list organisations", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", page)
}

// Directory lets a verified organisation find others to hand lots to. It
// shows only what a receiver needs to be recognised: name, kind, address.
func (h *Handlers) Directory(w http.ResponseWriter, r *http.Request) {
	q, ok := listq.FromRequest(w, r, DirectoryList)
	if !ok {
		return
	}
	kind := r.URL.Query().Get(ParamKind)
	if kind != "" && !slices.Contains(Kinds, kind) {
		listq.BadParam(w, ParamKind, Kinds)
		return
	}
	// The route requires a verified acting organisation, so it is set.
	own := auth.CurrentScope(r.Context()).Organisation.ID
	page, err := h.Store.Directory(r.Context(), q, own, kind)
	if err != nil {
		httpx.ServerError(w, "list directory", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", page)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	if o, ok := h.visibleOrg(w, r); ok {
		httpx.OK(w, http.StatusOK, "", o)
	}
}

// Review approves or rejects an organisation.
func (h *Handlers) Review(w http.ResponseWriter, r *http.Request) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}
	var rv Review
	if !httpx.DecodeJSON(w, r, &rv) {
		return
	}
	if msg := rv.Validate(); msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	// Generated up front; the store uses it only if the organisation has no key yet.
	key, err := h.newSigningKey()
	if err != nil {
		httpx.ServerError(w, "create signing key", err)
		return
	}
	o, err = h.Store.Review(r.Context(), o.ID, rv, auth.CurrentUser(r.Context()).ID, key)
	switch {
	case errors.Is(err, ErrNoDocuments):
		httpx.Error(w, http.StatusConflict, MsgNoDocuments)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case err != nil:
		httpx.ServerError(w, "review organisation", err)
	default:
		h.InvalidateOrg(r.Context(), o.Slug)
		httpx.OK(w, http.StatusOK, reviewMessage[rv.Decision], o)
	}
}

// Suspend answers POST /orgs/{id}/suspend: a verified organisation stops
// acting (RequireVerified refuses it) until reactivated.
func (h *Handlers) Suspend(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	if in.Reason = strings.TrimSpace(in.Reason); in.Reason == "" {
		httpx.Error(w, http.StatusBadRequest, MsgSuspendReason)
		return
	}
	h.setStatus(w, r, StatusVerified, StatusSuspended, in.Reason, MsgNotVerified, MsgOrgSuspended)
}

// Reactivate answers POST /orgs/{id}/reactivate: a suspended organisation is verified again.
func (h *Handlers) Reactivate(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, StatusSuspended, StatusVerified, "", MsgNotSuspended, MsgOrgReactivated)
}

func (h *Handlers) setStatus(w http.ResponseWriter, r *http.Request, from, to, reason, wrong, done string) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}
	o, err := h.Store.SetStatus(r.Context(), o.ID, from, to, reason, auth.CurrentUser(r.Context()).ID)
	switch {
	case errors.Is(err, ErrWrongStatus):
		httpx.Error(w, http.StatusConflict, wrong)
	case err != nil:
		httpx.ServerError(w, "set organisation status", err)
	default:
		h.InvalidateOrg(r.Context(), o.Slug)
		httpx.OK(w, http.StatusOK, done, o)
	}
}

// Accreditations lists which certificate types an organisation may issue.
func (h *Handlers) Accreditations(w http.ResponseWriter, r *http.Request) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}
	list, err := h.Store.Accreditations(r.Context(), o.ID)
	if err != nil {
		httpx.ServerError(w, "list accreditations", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", list)
}

// SetAccreditations replaces the certificate types an organisation may issue.
// An empty list withdraws every accreditation.
func (h *Handlers) SetAccreditations(w http.ResponseWriter, r *http.Request) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}
	var in []AccreditationInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	grants, msg := ParseGrants(in)
	if msg != "" {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}
	list, err := h.Store.SetAccreditations(r.Context(), o.ID, grants)
	switch {
	case errors.Is(err, ErrNotIssuer):
		httpx.Error(w, http.StatusConflict, MsgNotIssuer)
	case errors.Is(err, ErrUnknownType):
		httpx.Error(w, http.StatusBadRequest, MsgUnknownCertType)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
	case err != nil:
		httpx.ServerError(w, "set accreditations", err)
	case len(list) == 0:
		httpx.OK(w, http.StatusOK, MsgAccreditationsRemoved, list)
	default:
		httpx.OK(w, http.StatusOK, MsgAccreditationsSaved, list)
	}
}

func (h *Handlers) ListDocuments(w http.ResponseWriter, r *http.Request) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, DocList)
	if !ok {
		return
	}
	page, err := h.Store.ListDocuments(r.Context(), q, o.ID)
	if err != nil {
		httpx.ServerError(w, "list documents", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", page)
}

// Upload stores a KYB document for the organisation.
func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}
	file, header, ok := upload.Parse(w, r)
	if !ok {
		return
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()

	docType := r.FormValue(FormType)
	if !slices.Contains(DocTypes, docType) {
		httpx.Error(w, http.StatusBadRequest, FormType+" must be one of: "+strings.Join(DocTypes, ", "))
		return
	}
	f, ok := upload.Store(w, r, h.Files, file, header)
	if !ok {
		return
	}
	doc, statusChanged, err := h.Store.AddDocument(r.Context(), Document{
		OrganisationID: o.ID,
		Type:           docType,
		FileName:       f.Name,
		ContentType:    f.ContentType,
		SizeBytes:      f.Size,
		SHA256:         f.SHA256,
		StorageKey:     f.Key,
		UploadedBy:     auth.CurrentUser(r.Context()).ID,
	})
	if err != nil {
		h.Files.Remove(r.Context(), f.Key) // do not keep a file with no record
		httpx.ServerError(w, "record document", err)
		return
	}
	if statusChanged {
		h.InvalidateOrg(r.Context(), o.Slug)
	}
	httpx.OK(w, http.StatusCreated, MsgDocumentUploaded, doc)
}

// Download streams a stored document to a caller allowed to see it.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	id, ok := httpx.URLID(w, r)
	if !ok {
		return
	}
	d, err := h.Store.GetDocument(r.Context(), id, g)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if err != nil {
		httpx.ServerError(w, "get document", err)
		return
	}
	upload.Serve(w, r, h.Files, d.StorageKey, d.FileName, d.ContentType, d.UpdatedAt)
}
