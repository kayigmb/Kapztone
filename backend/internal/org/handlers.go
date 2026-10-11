package org

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
)

type Handlers struct {
	Store     Store
	UploadDir string
	// InvalidateOrg clears cached access data after an organisation changes.
	InvalidateOrg func(ctx context.Context, slug string)
	// Secrets seals each organisation's signing key before it is stored.
	Secrets *secret.Box
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
		httpx.WriteJSON(w, http.StatusCreated, o)
	}
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
	status := r.URL.Query().Get(ParamStatus)
	if status != "" && !slices.Contains(Statuses, status) {
		listq.BadParam(w, ParamStatus, Statuses)
		return
	}
	page, err := h.Store.List(r.Context(), q, g, status)
	if err != nil {
		httpx.ServerError(w, "list organisations", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	if o, ok := h.visibleOrg(w, r); ok {
		httpx.WriteJSON(w, http.StatusOK, o)
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
		httpx.WriteJSON(w, http.StatusOK, o)
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
	httpx.WriteJSON(w, http.StatusOK, page)
}

// Upload stores a KYB document for the organisation.
func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	o, ok := h.visibleOrg(w, r)
	if !ok {
		return
	}

	// The body limit leaves room for the multipart envelope around the file.
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadBytes+1<<20)
	file, header, err := r.FormFile(FormFile)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			httpx.Error(w, http.StatusRequestEntityTooLarge, MsgFileTooLarge)
			return
		}
		httpx.Error(w, http.StatusBadRequest, MsgFileNeeded)
		return
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()

	if header.Size > MaxUploadBytes {
		httpx.Error(w, http.StatusRequestEntityTooLarge, MsgFileTooLarge)
		return
	}
	docType := r.FormValue(FormType)
	if !slices.Contains(DocTypes, docType) {
		httpx.Error(w, http.StatusBadRequest, FormType+" must be one of: "+strings.Join(DocTypes, ", "))
		return
	}

	// Decide the type from the bytes; the uploader's file name and header are not trusted.
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	contentType := http.DetectContentType(head[:n])
	if !slices.Contains(AllowedContentTypes, contentType) {
		httpx.Error(w, http.StatusBadRequest, MsgFileType)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		httpx.ServerError(w, "rewind upload", err)
		return
	}

	key, sum, err := h.save(file)
	if err != nil {
		httpx.ServerError(w, "save upload", err)
		return
	}
	doc, statusChanged, err := h.Store.AddDocument(r.Context(), Document{
		OrganisationID: o.ID,
		Type:           docType,
		FileName:       filepath.Base(header.Filename),
		ContentType:    contentType,
		SizeBytes:      header.Size,
		SHA256:         sum,
		StorageKey:     key,
		UploadedBy:     auth.CurrentUser(r.Context()).ID,
	})
	if err != nil {
		os.Remove(filepath.Join(h.UploadDir, key)) // do not keep a file with no record
		httpx.ServerError(w, "record document", err)
		return
	}
	if statusChanged {
		h.InvalidateOrg(r.Context(), o.Slug)
	}
	httpx.WriteJSON(w, http.StatusCreated, doc)
}

// save writes the file under a random server-chosen name and returns that
// name and the SHA-256 of the content.
func (h *Handlers) save(src io.Reader) (key, sum string, err error) {
	raw := make([]byte, 16)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	key = hex.EncodeToString(raw)
	path := filepath.Join(h.UploadDir, key)

	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", "", err
	}
	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(dst, hash), src)
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(path)
		return "", "", err
	}
	return key, hex.EncodeToString(hash.Sum(nil)), nil
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
	f, err := os.Open(filepath.Join(h.UploadDir, d.StorageKey))
	if err != nil {
		httpx.ServerError(w, "open document", err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", d.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": d.FileName}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", d.UpdatedAt, f)
}
