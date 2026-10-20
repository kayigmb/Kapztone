package mailtmpl

import (
	"strings"
	"testing"
)

func TestRenderEscapesInput(t *testing.T) {
	html := Render(Page{
		Subject:     "Invite",
		Paragraphs:  []string{"Hello <b>you</b>", "Open https://example.com/a?b=c&d=e now"},
		ActionLabel: "Accept",
		ActionURL:   "https://example.com/a?b=c&d=e",
		LogoURL:     "https://example.com/logo.svg",
	})
	for _, bad := range []string{"<b>you</b>", "b=c&d=e\">", "<script"} {
		if strings.Contains(html, bad) {
			t.Fatalf("unescaped input in mail:\n%s", html)
		}
	}
	for _, want := range []string{
		`href="https://example.com/a?b=c&amp;d=e"`,
		`src="https://example.com/logo.svg"`,
		"Accept",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in mail:\n%s", want, html)
		}
	}
	if !strings.Contains(html, "ChainTrack") {
		t.Fatal("brand missing from mail")
	}
}

func TestRenderWithoutActionOrLogo(t *testing.T) {
	html := Render(Page{Subject: "Notice", Paragraphs: Split("one\n\ntwo")})
	if strings.Contains(html, "<a href=") {
		t.Fatal("no action should mean no links")
	}
	if strings.Contains(html, "<img") {
		t.Fatal("no logo should mean no image")
	}
	if strings.Count(html, "<p") < 2 {
		t.Fatal("paragraphs missing")
	}
}

func TestSplit(t *testing.T) {
	got := Split("a\r\n\r\nb\n\n\nc")
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("split = %q", got)
	}
}
