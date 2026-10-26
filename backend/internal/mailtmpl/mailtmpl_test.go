package mailtmpl

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestRenderEscapesInput(t *testing.T) {
	html := Render(Page{
		Subject:     "Invite",
		Paragraphs:  []string{"Hello <b>you</b>", "Open https://example.com/a?b=c&d=e now"},
		ActionLabel: "Accept",
		ActionURL:   "https://example.com/a?b=c&d=e",
	})
	for _, bad := range []string{"<b>you</b>", "b=c&d=e\">", "<script"} {
		if strings.Contains(html, bad) {
			t.Fatalf("unescaped input in mail:\n%s", html)
		}
	}
	for _, want := range []string{
		`href="https://example.com/a?b=c&amp;d=e"`,
		`src="cid:` + LogoCID + `"`,
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

func TestRenderWithoutAction(t *testing.T) {
	html := Render(Page{Subject: "Notice", Paragraphs: Split("one\n\ntwo")})
	if strings.Contains(html, "<a href=") {
		t.Fatal("no action should mean no links")
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

// The logo part is a PNG in base64 lines of at most 76 characters.
func TestLogoPart(t *testing.T) {
	for _, line := range strings.Split(LogoPart, "\r\n") {
		if len(line) > 76 {
			t.Fatalf("base64 line of %d characters", len(line))
		}
	}
	png, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(LogoPart, "\r\n", ""))
	if err != nil || !bytes.HasPrefix(png, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatalf("logo part is not a PNG (err %v)", err)
	}
}

func TestLetter(t *testing.T) {
	if got := Letter(" Ada ", "Body.\n"); got != "Hello Ada,\n\nBody.\n\nThe ChainTrack team" {
		t.Fatalf("letter = %q", got)
	}
	if got := Letter("", "Body."); !strings.HasPrefix(got, "Hello,\n\n") {
		t.Fatalf("letter without a name = %q", got)
	}
}
