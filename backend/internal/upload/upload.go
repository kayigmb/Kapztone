// Package upload receives, stores and serves files that users attach: KYB
// documents and certificates. It is the one place the rules for an upload live.
// Files are kept in an S3-compatible bucket: AWS S3, Google Cloud Storage
// (through its S3-compatible endpoint) or RustFS in development.
package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
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

// Config says which bucket holds the files and how to reach it.
type Config struct {
	Endpoint  string // URL; its scheme (http or https) decides TLS
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

// Bucket is where uploaded files are kept. The bucket is private: files
// reach a browser only through the API, after its permission check.
type Bucket struct {
	client *minio.Client
	name   string
}

// New connects to the bucket and checks it exists. With create (development)
// a missing bucket is made; otherwise it must be created beforehand.
func New(ctx context.Context, c Config, create bool) (*Bucket, error) {
	u, err := url.Parse(c.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("storage endpoint must be an http or https URL, got %q", c.Endpoint)
	}
	client, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: u.Scheme == "https",
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}
	exists, err := client.BucketExists(ctx, c.Bucket)
	switch {
	case err != nil:
		return nil, fmt.Errorf("storage bucket %s: %w", c.Bucket, err)
	case !exists && !create:
		return nil, fmt.Errorf("storage bucket %s does not exist: create it first", c.Bucket)
	case !exists:
		if err := client.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{Region: c.Region}); err != nil {
			return nil, fmt.Errorf("create storage bucket %s: %w", c.Bucket, err)
		}
	}
	return &Bucket{client: client, name: c.Bucket}, nil
}

// File is a stored upload.
type File struct {
	Key         string // server-chosen object name in the bucket
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

// Store checks the file's type and saves it in the bucket, or writes the error.
func Store(w http.ResponseWriter, r *http.Request, b *Bucket, file multipart.File, header *multipart.FileHeader) (File, bool) {
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
	key, sum, err := b.save(r.Context(), file, header.Size, contentType)
	if err != nil {
		httpx.ServerError(w, "save upload", err)
		return File{}, false
	}
	return File{Key: key, SHA256: sum, Name: filepath.Base(header.Filename), ContentType: contentType, Size: header.Size}, true
}

// save writes the file under a random server-chosen name and returns that
// name and the SHA-256 of the content.
func (b *Bucket) save(ctx context.Context, src io.Reader, size int64, contentType string) (key, sum string, err error) {
	raw := make([]byte, 16)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	key = hex.EncodeToString(raw)

	hash := sha256.New()
	_, err = b.client.PutObject(ctx, b.name, key, io.TeeReader(src, hash), size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", "", err
	}
	return key, hex.EncodeToString(hash.Sum(nil)), nil
}

// Put replaces a stored file's content. Only tests use it, to alter a file.
func (b *Bucket) Put(ctx context.Context, key string, content []byte) error {
	_, err := b.client.PutObject(ctx, b.name, key, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{})
	return err
}

// SaveBytes stores generated content (saved reports) under a random key.
func (b *Bucket) SaveBytes(ctx context.Context, content []byte, contentType string) (key, sum string, err error) {
	raw := make([]byte, 16)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	key = hex.EncodeToString(raw)
	hash := sha256.Sum256(content)
	sum = hex.EncodeToString(hash[:])
	_, err = b.client.PutObject(ctx, b.name, key, bytes.NewReader(content), int64(len(content)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", "", err
	}
	return key, sum, nil
}

// Remove deletes a stored file whose record could not be saved.
func (b *Bucket) Remove(ctx context.Context, key string) {
	if err := b.client.RemoveObject(ctx, b.name, key, minio.RemoveObjectOptions{}); err != nil {
		logx.From(ctx).Warn("orphan file left in bucket", "key", key, logx.KeyError, err)
	}
}

// Open returns a stored file's content. The caller closes it.
func (b *Bucket) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := b.client.GetObject(ctx, b.name, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// GetObject is lazy: Stat surfaces a missing object or an unreachable bucket now.
	if _, err := obj.Stat(); err != nil {
		obj.Close()
		return nil, err
	}
	return obj, nil
}

// Serve streams a stored file as a download.
func Serve(w http.ResponseWriter, r *http.Request, b *Bucket, key, name, contentType string, modified time.Time) {
	f, err := b.Open(r.Context(), key)
	if err != nil {
		httpx.ServerError(w, "open upload", err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// The object can seek, so range requests keep working.
	http.ServeContent(w, r, "", modified, f.(io.ReadSeeker))
}
