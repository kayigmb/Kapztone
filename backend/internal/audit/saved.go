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
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// CreateSaved stores one saved report definition.
func (h *Handlers) CreateSaved(w http.ResponseWriter, r *http.Request) {
	g, ok := auth.Gate(w, r)
	if !ok {
		return
	}
	in, ok := savedInput(w, r)
	if !ok {
		return
	}
	var actor *string
	if u := auth.CurrentUser(r.Context()); u.ID != "" {
		actor = &u.ID
	}
	saved, err := h.Store.CreateSaved(r.Context(), g, in.Name, in.Kind, in.Params, actor)
	if errors.Is(err, errTooManySaved) {
		httpx.Error(w, http.StatusTooManyRequests, "too many saved reports")
		return
	}
	if err != nil {
		httpx.ServerError(w, "save report", err)
		return
	}
	httpx.OK(w, http.StatusCreated, "Saved report kept.", saved)
}

type newSaved struct {
	Name   string          `json:"name"`
	Kind   string          `json:"kind"`
	Params json.RawMessage `json:"params"`
}

// savedInput checks the parameters now, so running the report later cannot fail on them.
func savedInput(w http.ResponseWriter, r *http.Request) (newSaved, bool) {
	var in newSaved
	if !httpx.DecodeJSON(w, r, &in) {
		return in, false
	}
	in.Name, in.Kind = strings.TrimSpace(in.Name), strings.TrimSpace(in.Kind)
	if len(in.Params) == 0 {
		in.Params = json.RawMessage(`{}`)
	}
	var p reportParams
	switch {
	case len(in.Name) == 0 || len(in.Name) > 120:
		httpx.Error(w, http.StatusBadRequest, MsgBadSavedName)
	case !slices.Contains(SavedKinds, in.Kind):
		httpx.Error(w, http.StatusBadRequest, MsgBadSavedKind)
	case json.Unmarshal(in.Params, &p) == nil && p.Actor != "" && !httpx.IsUUID(strings.TrimSpace(p.Actor)):
		httpx.Error(w, http.StatusBadRequest, MsgBadActor)
	default:
		return in, true
	}
	return in, false
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
	saved, ok := h.savedOf(w, r, g)
	if !ok {
		return
	}
	format, ok := formatOf(w, r)
	if !ok {
		return
	}
	if format == "" {
		format = FormatCSV
	}
	t, err := h.reportTable(r.Context(), g, saved)
	if err != nil {
		httpx.ServerError(w, saved.Kind+" report", err)
		return
	}
	file, err := h.keepReport(r.Context(), saved, format, t)
	if err != nil {
		httpx.ServerError(w, "keep report file", err)
		return
	}
	httpx.OK(w, http.StatusCreated, "Report generated and kept in the bucket.", file)
}

// keepReport renders the table and stores the file in the bucket, with its
// record; a file whose record fails is removed again.
func (h *Handlers) keepReport(ctx context.Context, saved Saved, format string, t table) (SavedFile, error) {
	content, contentType, err := renderReport(format, saved.Name, t.title, t.columns, t.rows)
	if err != nil {
		return SavedFile{}, err
	}
	key, sum, err := h.Files.SaveBytes(ctx, content, contentType)
	if err != nil {
		return SavedFile{}, err
	}
	file, err := h.Store.AddFile(ctx, saved.ID, format, key, sum, int64(len(content)))
	if err != nil {
		h.Files.Remove(ctx, key)
	}
	return file, err
}

func (h *Handlers) savedOf(w http.ResponseWriter, r *http.Request, g string) (Saved, bool) {
	id := chi.URLParam(r, httpx.ParamID)
	if !httpx.IsUUID(id) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return Saved{}, false
	}
	saved, err := h.Store.GetSaved(r.Context(), g, id)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return Saved{}, false
	}
	return saved, true
}

type table struct {
	title   string
	columns []string
	rows    [][]string
}

// reportTable runs a saved report with its stored parameters.
func (h *Handlers) reportTable(ctx context.Context, g string, saved Saved) (table, error) {
	var p reportParams
	_ = json.Unmarshal(saved.Params, &p)
	var start, end *time.Time
	if t, ok := httpx.ParseTime(strings.TrimSpace(p.StartDate)); ok {
		start = &t
	}
	if t, ok := httpx.ParseTime(strings.TrimSpace(p.EndDate)); ok {
		end = &t
	}
	switch saved.Kind {
	case SavedVolume:
		return h.volumeTable(ctx, g, start, end)
	case SavedBlocked:
		return h.blockedTable(ctx, g, start, end)
	case SavedOutcomes:
		return h.outcomesTable(ctx, g, start, end, strings.TrimSpace(p.Commodity))
	}
	return h.auditTable(ctx, g, start, end, p)
}

func (h *Handlers) volumeTable(ctx context.Context, g string, start, end *time.Time) (table, error) {
	t := table{title: TitleVolume, columns: VolumeColumns}
	vols, err := h.Store.Volumes(ctx, g, start, end)
	for _, v := range vols {
		t.rows = append(t.rows, []string{v.Month, v.Commodity, v.Unit,
			strconv.FormatInt(v.Quantity, 10), strconv.FormatInt(v.Exported, 10)})
	}
	return t, err
}

func (h *Handlers) blockedTable(ctx context.Context, g string, start, end *time.Time) (table, error) {
	t := table{title: TitleBlocked, columns: BlockedColumns}
	blocked, err := h.Store.Blocked(ctx, g, start, end)
	for _, b := range blocked {
		t.rows = append(t.rows, []string{b.Rule, strconv.FormatInt(b.Count, 10)})
	}
	return t, err
}

func (h *Handlers) outcomesTable(ctx context.Context, g string, start, end *time.Time, commodity string) (table, error) {
	t := table{title: TitleOutcomes, columns: OutcomesColumns}
	outcomes, err := h.Store.Outcomes(ctx, g, start, end, commodity)
	for _, o := range outcomes {
		t.rows = append(t.rows, []string{o.Period, strconv.FormatInt(o.Verified, 10),
			strconv.FormatInt(o.Warning, 10), strconv.FormatInt(o.Failed, 10),
			strconv.FormatInt(o.Total, 10), strconv.FormatInt(o.PctVerified, 10)})
	}
	return t, err
}

func (h *Handlers) auditTable(ctx context.Context, g string, start, end *time.Time, p reportParams) (table, error) {
	t := table{title: TitleAudit, columns: AuditColumns}
	q := listq.Query{Start: start, End: end}
	err := h.Store.Each(ctx, q, g, strings.TrimSpace(p.Actor), strings.TrimSpace(p.Route), func(e Entry) error {
		t.rows = append(t.rows, []string{strconv.FormatInt(e.Seq, 10), e.CreatedAt.UTC().Format(time.RFC3339),
			deref(e.ActorUserID), deref(e.ActorEmail), deref(e.OrganisationID), e.Method, e.Route,
			deref(e.SubjectID), strconv.Itoa(e.Status), e.Hash})
		return nil
	})
	return t, err
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
