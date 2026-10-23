package notify

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/mailtmpl"
)

// Every e-mail carries the logo inline: text, then HTML plus a PNG the HTML
// points at with cid:logo.
func TestComposeAttachesLogoInline(t *testing.T) {
	m := &Mailer{FromHeader: `"ChainTrack" <no-reply@chaintrack.test>`}
	msg, err := mail.ReadMessage(bytes.NewReader(m.compose([]string{"a@chaintrack.test"}, "Hi", "plain", "<p>html</p>")))
	if err != nil {
		t.Fatal(err)
	}
	_, params, _ := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	outer := multipart.NewReader(msg.Body, params["boundary"])
	if _, err := outer.NextPart(); err != nil { // text/plain
		t.Fatal(err)
	}
	related, err := outer.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	ct, params, _ := mime.ParseMediaType(related.Header.Get("Content-Type"))
	if ct != "multipart/related" {
		t.Fatalf("second part is %q, want multipart/related", ct)
	}
	inner := multipart.NewReader(related, params["boundary"])
	if _, err := inner.NextPart(); err != nil { // text/html
		t.Fatal(err)
	}
	logo, err := inner.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if logo.Header.Get("Content-ID") != "<"+mailtmpl.LogoCID+">" || logo.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("logo part headers: %v", logo.Header)
	}
	body, err := io.ReadAll(logo)
	if err != nil {
		t.Fatal(err)
	}
	png, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(body), "\r\n", ""))
	if err != nil || !bytes.HasPrefix(png, []byte("\x89PNG")) {
		t.Fatalf("logo part is not a PNG (err %v)", err)
	}
}
