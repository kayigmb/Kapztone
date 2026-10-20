// Package mailtmpl renders the HTML half of every e-mail ChainTrack sends.
// It is a leaf package on purpose: auth, user and notify all build mail,
// and notify already imports auth, so the templates cannot live in any of
// them. The plain-text half stays where it was, written by the callers.
package mailtmpl

import (
	"html"
	"strings"
)

// Page is one e-mail. Paragraphs are plain text; links in them become
// clickable, and ActionURL (when set, with ActionLabel) becomes the button.
// LogoURL is an absolute https URL of the ChainTrack mark; empty means the
// header shows the name as text instead of an image email clients may block.
type Page struct {
	Subject     string
	Preheader   string
	Paragraphs  []string
	ActionLabel string
	ActionURL   string
	Footnote    string
	LogoURL     string
}

// Split turns a plain-text body into paragraphs on blank lines.
func Split(body string) []string {
	var out []string
	for _, p := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Render builds the full HTML document: table layout, inline styles only
// (mail clients ignore stylesheets), navy header, teal action button.
func Render(p Page) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html><body style="margin:0;padding:0;background-color:#eef2f1;">`)
	// Hidden preview text shown beside the subject in inboxes.
	if p.Preheader != "" {
		b.WriteString(`<div style="display:none;max-height:0;overflow:hidden;opacity:0;">`)
		b.WriteString(html.EscapeString(p.Preheader))
		b.WriteString(`</div>`)
	}
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#eef2f1;padding:24px 12px;">`)
	b.WriteString(`<tr><td align="center">`)
	b.WriteString(`<table role="presentation" width="480" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#0f2a43;border-radius:12px 12px 0 0;">`)
	b.WriteString(`<tr><td style="padding:20px 24px;">`)
	if p.LogoURL != "" {
		b.WriteString(`<img src="`)
		b.WriteString(html.EscapeString(p.LogoURL))
		b.WriteString(`" alt="ChainTrack" width="32" height="32" style="display:block;border:0;" />`)
	} else {
		b.WriteString(`<span style="color:#ffffff;font-family:Arial,sans-serif;font-size:20px;font-weight:bold;">ChainTrack</span>`)
	}
	b.WriteString(`</td></tr></table>`)
	b.WriteString(`<table role="presentation" width="480" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#ffffff;border-radius:0 0 12px 12px;">`)
	b.WriteString(`<tr><td style="padding:24px;font-family:Arial,sans-serif;color:#1c2b3a;">`)
	b.WriteString(`<h1 style="margin:0 0 16px;font-size:20px;line-height:1.4;">`)
	b.WriteString(html.EscapeString(p.Subject))
	b.WriteString(`</h1>`)
	for _, para := range p.Paragraphs {
		b.WriteString(`<p style="margin:0 0 12px;font-size:15px;line-height:1.6;">`)
		b.WriteString(linkify(para))
		b.WriteString(`</p>`)
	}
	if p.ActionURL != "" && p.ActionLabel != "" {
		b.WriteString(`<p style="margin:20px 0 8px;"><a href="`)
		b.WriteString(html.EscapeString(p.ActionURL))
		b.WriteString(`" style="display:inline-block;background-color:#087566;color:#ffffff;font-family:Arial,sans-serif;font-size:15px;font-weight:bold;text-decoration:none;padding:12px 28px;border-radius:8px;">`)
		b.WriteString(html.EscapeString(p.ActionLabel))
		b.WriteString(`</a></p>`)
	}
	if p.Footnote != "" {
		b.WriteString(`<p style="margin:12px 0 0;font-size:13px;line-height:1.5;color:#5b6b7a;">`)
		b.WriteString(html.EscapeString(p.Footnote))
		b.WriteString(`</p>`)
	}
	b.WriteString(`</td></tr></table>`)
	b.WriteString(`<p style="font-family:Arial,sans-serif;font-size:12px;color:#5b6b7a;">Sent by ChainTrack. If you were not expecting this email, you can ignore it.</p>`)
	b.WriteString(`</td></tr></table></body></html>`)
	return b.String()
}

// linkify escapes text and turns http(s) URLs into links.
func linkify(text string) string {
	var out strings.Builder
	for i, word := range strings.Fields(text) {
		if i > 0 {
			out.WriteString(" ")
		}
		if (strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://")) && !strings.ContainsAny(word, "<>\"'") {
			out.WriteString(`<a href="`)
			out.WriteString(html.EscapeString(word))
			out.WriteString(`" style="color:#087566;word-break:break-all;">`)
			out.WriteString(html.EscapeString(word))
			out.WriteString(`</a>`)
			continue
		}
		out.WriteString(html.EscapeString(word))
	}
	return out.String()
}
