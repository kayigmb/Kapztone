package audit

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
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
				logx.From(r.Context()).Error("audit row not written", logx.KeyMethod, method, logx.KeyRoute, route, logx.KeyError, err)
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
	format, ok := formatOf(w, r)
	if !ok {
		return
	}
	if format == "" {
		page, err := h.Store.List(r.Context(), q, g, actor, route)
		if err != nil {
			httpx.ServerError(w, "list audit", err)
			return
		}
		httpx.OK(w, http.StatusOK, "", page)
		return
	}
	out := export(w, format, FileAudit, TitleAudit, describe(q, route), AuditColumns)
	err := h.Store.Each(r.Context(), q, g, actor, route, func(e Entry) error {
		return out.Row(strconv.FormatInt(e.Seq, 10), e.CreatedAt.UTC().Format(time.RFC3339),
			deref(e.ActorUserID), deref(e.ActorEmail), deref(e.OrganisationID), e.Method, e.Route,
			deref(e.SubjectID), strconv.Itoa(e.Status), e.Hash)
	})
	if err != nil {
		logx.From(r.Context()).Error("audit export cut short", logx.KeyError, err)
	}
	out.Close(r.Context())
}

// Verify walks the whole chain.
func (h *Handlers) Verify(w http.ResponseWriter, r *http.Request) {
	c, err := h.Store.Verify(r.Context())
	if err != nil {
		httpx.ServerError(w, "verify audit", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", c)
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
	format, ok := formatOf(w, r)
	if !ok {
		return
	}
	rows, err := h.Store.Volumes(r.Context(), g, q.Start, q.End)
	if err != nil {
		httpx.ServerError(w, "volume report", err)
		return
	}
	if format == "" {
		httpx.OK(w, http.StatusOK, "", map[string]any{"data": rows})
		return
	}
	out := export(w, format, FileVolume, TitleVolume, describe(q, ""), VolumeColumns)
	for _, v := range rows {
		if out.Row(v.Month, v.Commodity, v.Unit, strconv.FormatInt(v.Quantity, 10), strconv.FormatInt(v.Exported, 10)) != nil {
			break
		}
	}
	out.Close(r.Context())
}

// Blocked counts the actions each rule refused between start_date and end_date
// (both optional): capacity, mass balance, evidence, tolerance, and rejected handovers.
func (h *Handlers) Blocked(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, listq.Config{}) // only its dates are used
	if !ok {
		return
	}
	format, ok := formatOf(w, r)
	if !ok {
		return
	}
	rows, err := h.Store.Blocked(r.Context(), g, q.Start, q.End)
	if err != nil {
		httpx.ServerError(w, "blocked report", err)
		return
	}
	if format == "" {
		httpx.OK(w, http.StatusOK, "", map[string]any{"data": rows})
		return
	}
	out := export(w, format, FileBlocked, TitleBlocked, describe(q, ""), BlockedColumns)
	for _, b := range rows {
		if out.Row(b.Rule, strconv.FormatInt(b.Count, 10)) != nil {
			break
		}
	}
	out.Close(r.Context())
}

// formatOf reads ?format=: "" (JSON), FormatCSV or FormatPDF.
func formatOf(w http.ResponseWriter, r *http.Request) (string, bool) {
	f := r.URL.Query().Get(ParamFormat)
	if f != "" && !slices.Contains(Formats, f) {
		httpx.Error(w, http.StatusBadRequest, MsgBadFormat)
		return "", false
	}
	return f, true
}

// describe says which rows a download holds, for the PDF's subtitle.
func describe(q listq.Query, route string) string {
	day := func(t *time.Time) string {
		if t == nil {
			return "…"
		}
		return t.UTC().Format("2 Jan 2006")
	}
	s := "From " + day(q.Start) + " to " + day(q.End)
	if route != "" {
		s += " · route " + route
	}
	return s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
