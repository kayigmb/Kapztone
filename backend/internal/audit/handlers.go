package audit

import (
	"context"
	"encoding/csv"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

type Handlers struct {
	Store Store
}

// Record logs every successful (2xx) request to route. It runs inside the
// route's sign-in middleware, so the signed-in user and acting organisation
// are known. Only the URL's {id} is kept: request bodies (passwords, PINs)
// and other URL parts (invite tokens) never reach the log. A failure to log
// is reported, not passed on to the caller.
//
// ponytail: one global lock orders the chain, and the row is written after
// the change commits, so a crash in between loses that row. Upgrade: each
// handler writes its row in its own transaction, if either ever matters.
func (h *Handlers) Record(method, route string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK // body written without an explicit status
			}
			if status < 200 || status > 299 {
				return
			}
			e := Entry{Method: method, Route: route, Status: status}
			if u := auth.CurrentUser(r.Context()); u.ID != "" {
				e.ActorUserID = &u.ID
			}
			if o := auth.CurrentScope(r.Context()).Organisation; o != nil {
				e.OrganisationID = &o.ID
			}
			if id := chi.URLParam(r, httpx.ParamID); id != "" {
				e.SubjectID = &id
			}
			// The caller may already be gone; the row is still written.
			if err := h.Store.Add(context.WithoutCancel(r.Context()), e); err != nil {
				log.Printf("audit %s %s: %v", method, route, err)
			}
		})
	}
}

// List is the audit log, newest first, as a page of JSON or (format=csv)
// every matching row as a download. Superadmin sees all, an organisation
// admin their organisation's rows.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, ListConfig)
	if !ok {
		return
	}
	v := r.URL.Query()
	actor, route := v.Get(ParamActor), v.Get(ParamRoute)
	if actor != "" && !httpx.IsUUID(actor) {
		httpx.Error(w, http.StatusBadRequest, MsgBadActor)
		return
	}
	switch v.Get(ParamFormat) {
	case "":
		page, err := h.Store.List(r.Context(), q, g, actor, route)
		if err != nil {
			httpx.ServerError(w, "list audit", err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, page)
	case FormatCSV:
		out := csvDownload(w, FileAudit, AuditColumns)
		err := h.Store.Each(r.Context(), q, g, actor, route, func(e Entry) error {
			return out.Write(cells(strconv.FormatInt(e.Seq, 10), e.CreatedAt.UTC().Format(time.RFC3339),
				deref(e.ActorUserID), deref(e.ActorEmail), deref(e.OrganisationID), e.Method, e.Route,
				deref(e.SubjectID), strconv.Itoa(e.Status), e.Hash))
		})
		finish(out, "audit csv", err)
	default:
		httpx.Error(w, http.StatusBadRequest, MsgBadFormat)
	}
}

// Verify walks the whole chain.
func (h *Handlers) Verify(w http.ResponseWriter, r *http.Request) {
	c, err := h.Store.Verify(r.Context())
	if err != nil {
		httpx.ServerError(w, "verify audit", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, c)
}

// Volume is the quantity issued at origin per month, commodity and unit,
// between start_date and end_date (both optional).
func (h *Handlers) Volume(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, listq.Config{}) // only its dates are used
	if !ok {
		return
	}
	format := r.URL.Query().Get(ParamFormat)
	if format != "" && format != FormatCSV {
		httpx.Error(w, http.StatusBadRequest, MsgBadFormat)
		return
	}
	rows, err := h.Store.Volumes(r.Context(), g, q.Start, q.End)
	if err != nil {
		httpx.ServerError(w, "volume report", err)
		return
	}
	if format == "" {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"data": rows})
		return
	}
	out := csvDownload(w, FileVolume, VolumeColumns)
	for _, v := range rows {
		if err = out.Write(cells(v.Month, v.Commodity, v.Unit, strconv.FormatInt(v.Quantity, 10))); err != nil {
			break
		}
	}
	finish(out, "volume csv", err)
}

// cells guards a CSV row against spreadsheet formulas: a cell starting with
// = + - @ (an e-mail address can) is read by Excel as a formula, so it is
// written with a leading apostrophe.
func cells(row ...string) []string {
	for i, c := range row {
		if c != "" && strings.ContainsRune("=+-@\t\r", rune(c[0])) {
			row[i] = "'" + c
		}
	}
	return row
}

// csvDownload starts a CSV attachment and writes its heading row.
func csvDownload(w http.ResponseWriter, name string, columns []string) *csv.Writer {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	out := csv.NewWriter(w)
	out.Write(columns)
	return out
}

// finish flushes a download. Once rows are sent the status cannot change,
// so a failure part-way is only logged.
func finish(out *csv.Writer, what string, err error) {
	out.Flush()
	if err == nil {
		err = out.Error()
	}
	if err != nil {
		log.Printf("%s: %v", what, err)
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
