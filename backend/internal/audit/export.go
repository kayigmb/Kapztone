package audit

import (
	"context"
	"encoding/csv"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

// exporter writes one table as a download. CSV streams each row as it comes;
// PDF lays the rows out on landscape A4 pages when it is closed.
type exporter interface {
	Row(cells ...string) error
	Close(ctx context.Context) // a failure part-way can only be logged: the status is already sent
}

// export starts a download of a table in format (FormatCSV or FormatPDF).
// title and filters head the PDF; name is the file name without extension.
func export(w http.ResponseWriter, format, name, title, filters string, columns []string) exporter {
	switch format {
	case FormatPDF:
		return &pdfExport{w: w, name: name, title: title, filters: filters, columns: columns}
	default:
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", attachment(name+".csv"))
		out := csv.NewWriter(w)
		out.Write(columns)
		return &csvExport{out: out, name: name}
	}
}

func attachment(file string) string {
	return mime.FormatMediaType("attachment", map[string]string{"filename": file})
}

type csvExport struct {
	out  *csv.Writer
	name string
	err  error
}

func (c *csvExport) Row(row ...string) error {
	if c.err == nil {
		c.err = c.out.Write(cells(row...))
	}
	return c.err
}

func (c *csvExport) Close(ctx context.Context) {
	c.out.Flush()
	if c.err == nil {
		c.err = c.out.Error()
	}
	if c.err != nil {
		logx.From(ctx).Error("export cut short", "export", c.name, logx.KeyError, c.err)
	}
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

// PDF layout, in millimetres and points.
const (
	pdfMargin   = 10.0
	pdfRowH     = 5.5
	pdfFontSize = 7.5
	// MaxPDFRows caps a PDF; past it the last line says how many were left out
	// and points to the CSV. ponytail: a cap, not streaming; raise it or page the PDF if anyone needs more.
	MaxPDFRows = 5000
)

type pdfExport struct {
	w                    http.ResponseWriter
	name, title, filters string
	columns              []string
	rows                 [][]string
	dropped              int
}

func (p *pdfExport) Row(row ...string) error {
	if len(p.rows) >= MaxPDFRows {
		p.dropped++
		return nil
	}
	p.rows = append(p.rows, append([]string(nil), row...))
	return nil
}

func (p *pdfExport) Close(ctx context.Context) {
	doc := fpdf.New("L", "mm", "A4", "")
	doc.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	doc.SetAutoPageBreak(true, pdfMargin)
	tr := doc.UnicodeTranslatorFromDescriptor("") // cp1252: accents in names and places survive
	pageW, _ := doc.GetPageSize()
	colW := (pageW - 2*pdfMargin) / float64(len(p.columns))

	header := func() {
		doc.SetFont("Helvetica", "B", pdfFontSize)
		doc.SetFillColor(238, 242, 241)
		for _, c := range p.columns {
			doc.CellFormat(colW, pdfRowH+1, tr(fit(doc, c, colW)), "B", 0, "L", true, 0, "")
		}
		doc.Ln(-1)
		doc.SetFont("Helvetica", "", pdfFontSize)
	}
	doc.SetHeaderFunc(func() {
		if doc.PageNo() > 1 {
			header()
		}
	})
	doc.SetFooterFunc(func() {
		doc.SetY(-pdfMargin)
		doc.SetFont("Helvetica", "", 7)
		doc.CellFormat(0, 4, fmt.Sprintf("ChainTrack · %s · page %d", p.title, doc.PageNo()), "", 0, "R", false, 0, "")
	})

	doc.AddPage()
	doc.SetFont("Helvetica", "B", 14)
	doc.CellFormat(0, 8, tr(p.title), "", 1, "L", false, 0, "")
	doc.SetFont("Helvetica", "", 8)
	doc.CellFormat(0, 5, tr(fmt.Sprintf("Generated %s UTC · %s", time.Now().UTC().Format("2 Jan 2006 15:04"), p.filters)), "", 1, "L", false, 0, "")
	doc.Ln(2)
	header()
	for _, row := range p.rows {
		for i := range p.columns {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			doc.CellFormat(colW, pdfRowH, tr(fit(doc, cell, colW)), "B", 0, "L", false, 0, "")
		}
		doc.Ln(-1)
	}
	if len(p.rows) == 0 {
		doc.CellFormat(0, pdfRowH, "No rows.", "", 1, "L", false, 0, "")
	}
	if p.dropped > 0 {
		doc.Ln(2)
		doc.CellFormat(0, pdfRowH, fmt.Sprintf("%d more rows left out; download the CSV for every row.", p.dropped), "", 1, "L", false, 0, "")
	}

	p.w.Header().Set("Content-Type", "application/pdf")
	p.w.Header().Set("Content-Disposition", attachment(p.name+".pdf"))
	if err := doc.Output(p.w); err != nil {
		logx.From(ctx).Error("pdf export failed", "export", p.name, logx.KeyError, err)
	}
}

// fit shortens text with an ellipsis so it stays inside its column.
func fit(doc *fpdf.Fpdf, s string, width float64) string {
	limit := width - 1.5
	if doc.GetStringWidth(s) <= limit {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && doc.GetStringWidth(string(r)+"...") > limit {
		r = r[:len(r)-1]
	}
	return string(r) + "..."
}
