// Package upload receives, stores and serves files that users attach: KYB
// documents and certificates. It is the one place the rules for an upload live.
package upload

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// Every value in this file is declared here once and referenced everywhere
// else, including tests.
const (
	MaxBytes = 10 << 20

	FormFile = "file" // multipart field holding the file
)

// File types accepted, detected from the file's content.
var AllowedContentTypes = []string{"application/pdf", "image/png", "image/jpeg"}

// Messages shown to callers.
const (
	MsgFileNeeded   = "attach the file in the 'file' field"
	MsgFileTooLarge = "file is larger than 10 MB"
	MsgFileType     = "file must be a PDF, PNG or JPEG"
)

// File is a stored upload.
type File struct {
	Key         string // server-chosen name under the upload directory
	SHA256      string // hex
	Name        string // the uploader's file name, reduced to its base
	ContentType string // decided from the content
	Size        int64
}

// Parse limits the request body and returns the uploaded file, or writes the
// error. The form's other fields can be read once it returns. The caller
// closes the file and removes r.MultipartForm.
func Parse(w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, bool) {
	// The body limit leaves room for the multipart envelope around the file.
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytes+1<<20)
	file, header, err := r.FormFile(FormFile)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			httpx.Error(w, http.StatusRequestEntityTooLarge, MsgFileTooLarge)
			return nil, nil, false
		}
		httpx.Error(w, http.StatusBadRequest, MsgFileNeeded)
		return nil, nil, false
	}
	if header.Size > MaxBytes {
		file.Close()
		r.MultipartForm.RemoveAll()
		httpx.Error(w, http.StatusRequestEntityTooLarge, MsgFileTooLarge)
		return nil, nil, false
	}
	return file, header, true
}

// Store checks the file's type and saves it under dir, or writes the error.
func Store(w http.ResponseWriter, dir string, file multipart.File, header *multipart.FileHeader) (File, bool) {
	// Decide the type from the bytes; the uploader's file name and header are not trusted.
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	contentType := http.DetectContentType(head[:n])
	if !slices.Contains(AllowedContentTypes, contentType) {
		httpx.Error(w, http.StatusBadRequest, MsgFileType)
		return File{}, false
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		httpx.ServerError(w, "rewind upload", err)
		return File{}, false
	}
	key, sum, err := save(dir, file)
	if err != nil {
		httpx.ServerError(w, "save upload", err)
		return File{}, false
	}
	return File{Key: key, SHA256: sum, Name: filepath.Base(header.Filename), ContentType: contentType, Size: header.Size}, true
}

// save writes the file under a random server-chosen name and returns that
// name and the SHA-256 of the content.
func save(dir string, src io.Reader) (key, sum string, err error) {
	raw := make([]byte, 16)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	key = hex.EncodeToString(raw)
	path := filepath.Join(dir, key)

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

// Remove deletes a stored file whose record could not be saved.
func Remove(dir, key string) { os.Remove(filepath.Join(dir, key)) }

// Serve streams a stored file as a download.
func Serve(w http.ResponseWriter, r *http.Request, dir, key, name, contentType string, modified time.Time) {
	f, err := os.Open(filepath.Join(dir, key))
	if err != nil {
		httpx.ServerError(w, "open upload", err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", modified, f)
}
