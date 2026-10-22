package audit

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

type Handlers struct {
	Store Store
	Files *upload.Bucket
}

var errNoBucket = errBucket("report bucket is not configured")

type errBucket string

func (e errBucket) Error() string { return string(e) }

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

// Outcomes is the verification share per month.
func (h *Handlers) Outcomes(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	q, ok := listq.FromRequest(w, r, listq.Config{}) // only dates used
	if !ok {
		return
	}
	format, ok := formatOf(w, r)
	if !ok {
		return
	}
	commodity := strings.TrimSpace(r.URL.Query().Get("commodity"))
	rows, err := h.Store.Outcomes(r.Context(), g, q.Start, q.End, commodity)
	if err != nil {
		httpx.ServerError(w, "verification outcomes", err)
		return
	}
	if format == "" {
		httpx.OK(w, http.StatusOK, "", map[string]any{"data": rows})
		return
	}
	out := export(w, format, FileOutcomes, TitleOutcomes, describe(q, commodity), OutcomesColumns)
	for _, o := range rows {
		if out.Row(o.Period, strconv.FormatInt(o.Verified, 10), strconv.FormatInt(o.Warning, 10),
			strconv.FormatInt(o.Failed, 10), strconv.FormatInt(o.Total, 10),
			strconv.FormatInt(o.PctVerified, 10)) != nil {
			break
		}
	}
	out.Close(r.Context())
}

// CreateSaved stores one saved report definition.
func (h *Handlers) CreateSaved(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	var in struct {
		Name   string          `json:"name"`
		Kind   string          `json:"kind"`
		Params json.RawMessage `json:"params"`
	}
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Kind = strings.TrimSpace(in.Kind)
	if len(in.Name) == 0 || len(in.Name) > 120 {
		httpx.Error(w, http.StatusBadRequest, MsgBadSavedName)
		return
	}
	if !slices.Contains(SavedKinds, in.Kind) {
		httpx.Error(w, http.StatusBadRequest, MsgBadSavedKind)
		return
	}
	if len(in.Params) == 0 {
		in.Params = json.RawMessage(`{}`)
	}
	// Checked now, so running the report later cannot fail on it.
	var p reportParams
	if json.Unmarshal(in.Params, &p) == nil && p.Actor != "" && !httpx.IsUUID(strings.TrimSpace(p.Actor)) {
		httpx.Error(w, http.StatusBadRequest, MsgBadActor)
		return
	}
	var actor *string
	if u := auth.CurrentUser(r.Context()); u.ID != "" {
		actor = &u.ID
	}
	saved, err := h.Store.CreateSaved(r.Context(), g, in.Name, in.Kind, in.Params, actor)
	if err != nil {
		if errors.Is(err, errTooManySaved) {
			httpx.Error(w, http.StatusTooManyRequests, "too many saved reports")
			return
		}
		httpx.ServerError(w, "save report", err)
		return
	}
	httpx.OK(w, http.StatusCreated, "Saved report kept.", saved)
}

// ListSaved lists saved report definitions.
func (h *Handlers) ListSaved(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	rows, err := h.Store.ListSaved(r.Context(), g)
	if err != nil {
		httpx.ServerError(w, "list saved reports", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", map[string]any{"data": rows})
}

// reportParams carries the date window stored on a saved report.
type reportParams struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Commodity string `json:"commodity"`
	Route     string `json:"route"`
	Actor     string `json:"actor"`
}

// RunSaved generates the report now and keeps the file in the bucket.
func (h *Handlers) RunSaved(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	if h.Files == nil {
		httpx.ServerError(w, "report bucket", errNoBucket)
		return
	}
	id := chi.URLParam(r, httpx.ParamID)
	if !httpx.IsUUID(id) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	saved, err := h.Store.GetSaved(r.Context(), g, id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	format := r.URL.Query().Get(ParamFormat)
	if format == "" {
		format = FormatCSV
	}
	if !slices.Contains(Formats, format) {
		httpx.Error(w, http.StatusBadRequest, MsgBadFormat)
		return
	}
	var p reportParams
	_ = json.Unmarshal(saved.Params, &p)
	var start, end *time.Time
	if t, ok := httpx.ParseTime(strings.TrimSpace(p.StartDate)); ok {
		start = &t
	}
	if t, ok := httpx.ParseTime(strings.TrimSpace(p.EndDate)); ok {
		end = &t
	}
	columns, rows := []string{}, [][]string{}
	title := TitleSaved
	switch saved.Kind {
	case SavedVolume:
		title = TitleVolume
		columns = VolumeColumns
		vols, err := h.Store.Volumes(r.Context(), g, start, end)
		if err != nil {
			httpx.ServerError(w, "volume report", err)
			return
		}
		for _, v := range vols {
			rows = append(rows, []string{v.Month, v.Commodity, v.Unit,
				strconv.FormatInt(v.Quantity, 10), strconv.FormatInt(v.Exported, 10)})
		}
	case SavedBlocked:
		title = TitleBlocked
		columns = BlockedColumns
		blocked, err := h.Store.Blocked(r.Context(), g, start, end)
		if err != nil {
			httpx.ServerError(w, "blocked report", err)
			return
		}
		for _, b := range blocked {
			rows = append(rows, []string{b.Rule, strconv.FormatInt(b.Count, 10)})
		}
	case SavedOutcomes:
		title = TitleOutcomes
		columns = OutcomesColumns
		outcomes, err := h.Store.Outcomes(r.Context(), g, start, end, strings.TrimSpace(p.Commodity))
		if err != nil {
			httpx.ServerError(w, "verification outcomes", err)
			return
		}
		for _, o := range outcomes {
			rows = append(rows, []string{o.Period, strconv.FormatInt(o.Verified, 10),
				strconv.FormatInt(o.Warning, 10), strconv.FormatInt(o.Failed, 10),
				strconv.FormatInt(o.Total, 10), strconv.FormatInt(o.PctVerified, 10)})
		}
	default:
		title = TitleAudit
		columns = AuditColumns
		q := listq.Query{Start: start, End: end}
		err := h.Store.Each(r.Context(), q, g, strings.TrimSpace(p.Actor), strings.TrimSpace(p.Route), func(e Entry) error {
			rows = append(rows, []string{strconv.FormatInt(e.Seq, 10), e.CreatedAt.UTC().Format(time.RFC3339),
				deref(e.ActorUserID), deref(e.ActorEmail), deref(e.OrganisationID), e.Method, e.Route,
				deref(e.SubjectID), strconv.Itoa(e.Status), e.Hash})
			return nil
		})
		if err != nil {
			httpx.ServerError(w, "audit report", err)
			return
		}
	}
	content, contentType, err := renderReport(format, saved.Name, title, columns, rows)
	if err != nil {
		httpx.ServerError(w, "render report", err)
		return
	}
	key, sum, err := h.Files.SaveBytes(r.Context(), content, contentType)
	if err != nil {
		httpx.ServerError(w, "save report file", err)
		return
	}
	file, err := h.Store.AddFile(r.Context(), saved.ID, format, key, sum, int64(len(content)))
	if err != nil {
		h.Files.Remove(r.Context(), key)
		httpx.ServerError(w, "record report file", err)
		return
	}
	httpx.OK(w, http.StatusCreated, "Report generated and kept in the bucket.", file)
}

// ListFiles lists a saved report's generated files.
func (h *Handlers) ListFiles(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, httpx.ParamID)
	if !httpx.IsUUID(id) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if _, err := h.Store.GetSaved(r.Context(), g, id); err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	files, err := h.Store.ListFiles(r.Context(), g, id)
	if err != nil {
		httpx.ServerError(w, "list report files", err)
		return
	}
	httpx.OK(w, http.StatusOK, "", map[string]any{"data": files})
}

// DownloadFile streams one generated file from the bucket.
func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	if h.Files == nil {
		httpx.ServerError(w, "report bucket", errNoBucket)
		return
	}
	id := chi.URLParam(r, httpx.ParamID)
	if !httpx.IsUUID(id) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	file, err := h.Store.GetFile(r.Context(), g, id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	contentType := "text/csv"
	name := FileSaved + "." + file.Format
	if file.Format == FormatPDF {
		contentType = "application/pdf"
	}
	upload.Serve(w, r, h.Files, file.BucketKey, name, contentType, file.CreatedAt)
}

func renderReport(format, name, title string, columns []string, rows [][]string) ([]byte, string, error) {
	if format == FormatPDF {
		rec := httptest.NewRecorder()
		out := &pdfExport{w: rec, name: name, title: title, filters: "", columns: columns}
		for _, row := range rows {
			if err := out.Row(row...); err != nil {
				return nil, "", err
			}
		}
		out.Close(context.Background())
		res := rec.Result()
		defer res.Body.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(res.Body); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "application/pdf", nil
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(cells(columns...))
	for _, row := range rows {
		_ = w.Write(cells(row...))
	}
	w.Flush()
	return buf.Bytes(), "text/csv", w.Error()
}
