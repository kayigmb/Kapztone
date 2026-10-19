// Package logx is the structured logging of the API and the relayer: one JSON
// (or text) line per event, each request's lines sharing its request id, and
// one access line per request saying who did what, how it went and how long it took.
package logx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// Formats and levels the LOG_FORMAT and LOG_LEVEL settings take.
const (
	FormatJSON = "json"
	FormatText = "text"
)

var Levels = map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}

// HeaderRequestID carries the request id from the web app to the API and back.
const HeaderRequestID = "X-Request-ID"

// Field names, the same on every line so a log search can rely on them.
const (
	KeyRequestID = "request_id"
	KeyMethod    = "method"
	KeyRoute     = "route"
	KeyPath      = "path"
	KeyStatus    = "status"
	KeyBytes     = "bytes"
	KeyDuration  = "duration_ms"
	KeyIP        = "ip"
	KeyUser      = "user_id"
	KeyOrg       = "org"
	KeyError     = "error"
	KeyErrorAt   = "error_at"
	KeyPanic     = "panic"
	KeyStack     = "stack"
)

// Setup makes the process logger and installs it as slog's default (so
// log.Printf from libraries lands in it too).
func Setup(w io.Writer, format, level string) (*slog.Logger, error) {
	lv, ok := Levels[strings.ToLower(level)]
	if !ok {
		return nil, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error, not %q", level)
	}
	opts := &slog.HandlerOptions{Level: lv}
	var h slog.Handler
	switch strings.ToLower(format) {
	case FormatJSON:
		h = slog.NewJSONHandler(w, opts)
	case FormatText:
		h = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("LOG_FORMAT must be %s or %s, not %q", FormatJSON, FormatText, format)
	}
	l := slog.New(h)
	slog.SetDefault(l)
	return l, nil
}

// Fatal logs an error and stops the process; for start-up failures only.
func Fatal(msg string, err error, args ...any) {
	slog.Error(msg, append(args, KeyError, err)...)
	os.Exit(1)
}

type loggerKey struct{}
type infoKey struct{}

// From is the logger for this request (carrying its request id), or the process logger.
func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// info is what later middleware learns about a request (who, for whom, what
// failed) and the access line reports.
type info struct {
	mu        sync.Mutex
	user, org string
	err       error
	errAt     string
}

func infoOf(ctx context.Context) *info {
	i, _ := ctx.Value(infoKey{}).(*info)
	return i
}

// SetUser records who is signed in, for the access line.
func SetUser(ctx context.Context, userID string) {
	if i := infoOf(ctx); i != nil {
		i.mu.Lock()
		i.user = userID
		i.mu.Unlock()
	}
}

// SetOrg records which organisation the request acts for, for the access line.
func SetOrg(ctx context.Context, slug string) {
	if i := infoOf(ctx); i != nil {
		i.mu.Lock()
		i.org = slug
		i.mu.Unlock()
	}
}

// recorder captures the status and size for the access line, and the error a
// handler answered 500 for (httpx.ServerError hands it over through RecordError).
type recorder struct {
	http.ResponseWriter
	status int
	bytes  int
	info   *info
}

func (r *recorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the real writer (flush, deadlines).
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// RecordError hands a handler's server error to the access line. w is the
// writer the handler was given; wrappers that Unwrap are seen through.
func RecordError(w http.ResponseWriter, what string, err error) {
	for w != nil {
		if r, ok := w.(*recorder); ok {
			r.info.mu.Lock()
			r.info.err, r.info.errAt = err, what
			r.info.mu.Unlock()
			return
		}
		u, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			break
		}
		w = u.Unwrap()
	}
	slog.Error("server error outside a request", KeyErrorAt, what, KeyError, err)
}

// A request id from the web app is kept if it looks like one; otherwise a new one is made.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Middleware gives every request an id (echoed in X-Request-ID), a logger
// carrying it, panic recovery, and one access line when it ends: 5xx at
// error level, 4xx at warn, the rest at info.
func Middleware(clientIP func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			id := r.Header.Get(HeaderRequestID)
			if !idPattern.MatchString(id) {
				id = newID()
			}
			w.Header().Set(HeaderRequestID, id)
			l := slog.Default().With(KeyRequestID, id)
			in := &info{}
			ctx := context.WithValue(context.WithValue(r.Context(), loggerKey{}, l), infoKey{}, in)
			rec := &recorder{ResponseWriter: w, info: in}

			defer func() {
				if p := recover(); p != nil {
					if p == http.ErrAbortHandler { // the client went away; nothing to report
						panic(p)
					}
					in.err, in.errAt = fmt.Errorf("%v", p), KeyPanic
					l.Error("panic", KeyPanic, fmt.Sprint(p), KeyStack, string(debug.Stack()))
					if rec.status == 0 {
						http.Error(rec, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				}
				status := rec.status
				if status == 0 {
					status = http.StatusOK
				}
				route := r.URL.Path
				if rc := chi.RouteContext(r.Context()); rc != nil && rc.RoutePattern() != "" {
					route = rc.RoutePattern()
				}
				in.mu.Lock()
				attrs := []any{
					KeyMethod, r.Method, KeyRoute, route, KeyPath, r.URL.Path, KeyStatus, status, KeyBytes, rec.bytes,
					KeyDuration, time.Since(start).Milliseconds(), KeyIP, clientIP(r),
				}
				if in.user != "" {
					attrs = append(attrs, KeyUser, in.user)
				}
				if in.org != "" {
					attrs = append(attrs, KeyOrg, in.org)
				}
				if in.err != nil {
					attrs = append(attrs, KeyErrorAt, in.errAt, KeyError, in.err.Error())
				}
				in.mu.Unlock()
				level := slog.LevelInfo
				switch {
				case status >= 500:
					level = slog.LevelError
				case status >= 400:
					level = slog.LevelWarn
				}
				l.Log(ctx, level, "request", attrs...)
			}()
			next.ServeHTTP(rec, r.WithContext(ctx))
		})
	}
}
