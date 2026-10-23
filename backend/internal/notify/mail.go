package notify

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/mailtmpl"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

// Mailer sends the e-mail copy of each notification to the organisation's
// active users.
//
// One loop in the API process, at-least-once: a crash between
// sending and marking sends that e-mail again. Move to its own worker with a
// provider's idempotency key if duplicates ever matter.
type Mailer struct {
	Pool *pgxpool.Pool
	Addr string // host:port
	Auth smtp.Auth
	// From is the envelope sender (a bare address); FromHeader is what the
	// recipient sees, e.g. ChainTrack <noreply@chaintrack.digital>.
	From       string
	FromHeader string
	// LogoURL is the absolute URL of the ChainTrack mark shown atop HTML
	// mail; empty means the header shows the name as text.
	LogoURL string
	// Send delivers one message; NewMailer sets it, tests replace it.
	Send func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// SMTP is a mail server and the login it needs; the sender every recipient
// sees comes from MAIL_FROM instead (a display name and address).
type SMTP struct {
	Host     string
	Port     string
	User     string // the login; empty with Password means no sign-in
	Password string
	FromName string // display name, e.g. ChainTrack
	FromAddr string // envelope sender, e.g. noreply@chaintrack.digital
	// TLS requires the connection to be encrypted (STARTTLS). Without a
	// server that offers no encryption is written to in plain text.
	TLS bool
}

func NewMailer(pool *pgxpool.Pool, s SMTP) *Mailer {
	header := (&mail.Address{Name: s.FromName, Address: s.FromAddr}).String()
	m := &Mailer{Pool: pool, Addr: net.JoinHostPort(s.Host, s.Port), From: s.FromAddr, FromHeader: header}
	if s.Password != "" {
		m.Auth = smtp.PlainAuth("", s.User, s.Password, s.Host)
	}
	m.Send = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return send(addr, a, from, to, msg, s.TLS)
	}
	return m
}

// send is smtp.SendMail, except that with requireTLS a server that does not
// offer STARTTLS is refused instead of being written to in plain text.
// STARTTLS only (port 587); add tls.Dial if a provider needs port 465.
func send(addr string, a smtp.Auth, from string, to []string, msg []byte, requireTLS bool) error {
	host, _, _ := net.SplitHostPort(addr)
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	} else if requireTLS {
		return fmt.Errorf("mail server %s offers no encryption (STARTTLS); nothing was sent", addr)
	}
	if a != nil {
		if err := c.Auth(a); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// Run sends until ctx ends. A mail server that cannot be reached is not an
// error to give up on: notifications wait and stay visible in the app.
func (m *Mailer) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := m.SendPending(ctx); err != nil {
			logx.From(ctx).Warn("mailer: will retry", logx.KeyError, err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(mailInterval):
		}
	}
}

// SendPending e-mails one batch of notifications that have not been sent yet.
func (m *Mailer) SendPending(ctx context.Context) error {
	tx, err := m.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch, err := pendingMail(ctx, tx)
	if err != nil {
		return err
	}
	for _, p := range batch {
		if len(p.to) > 0 {
			if err := m.Send(m.Addr, m.Auth, m.From, p.to, m.message(p.to, p.message)); err != nil {
				// Keep what was sent, back this one off (10 s, 20 s, 40 s, ...); the rest wait for the next round.
				if _, uerr := tx.Exec(ctx,
					`UPDATE notifications SET mail_attempts = mail_attempts + 1,
					        mail_next_at = now() + make_interval(secs => $2 * 2 ^ mail_attempts) WHERE id = $1::uuid`,
					p.id, mailInterval.Seconds()); uerr != nil {
					return uerr
				}
				if cerr := tx.Commit(ctx); cerr != nil {
					return cerr
				}
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE notifications SET emailed_at = now() WHERE id = $1::uuid`, p.id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type pendingNote struct {
	id, message string
	to          []string
}

// pendingMail locks the next batch with its recipients. SKIP LOCKED: two API
// processes never pick the same notification.
func pendingMail(ctx context.Context, tx pgx.Tx) ([]pendingNote, error) {
	rows, err := tx.Query(ctx,
		`SELECT n.id::text, n.message,
		        ARRAY(SELECT u.email::text FROM userscope s JOIN users u ON u.id = s.user_id
		              WHERE s.organisation_id IS NOT DISTINCT FROM n.organisation_id AND NOT s.is_deleted
		                AND NOT u.is_deleted AND u.status = $3 AND u.email_notifications)
		 FROM notifications n
		 WHERE n.emailed_at IS NULL AND n.created_at > now() - make_interval(secs => $1) AND NOT n.is_deleted
		   AND n.mail_attempts < $4 AND (n.mail_next_at IS NULL OR n.mail_next_at <= now())
		 ORDER BY n.created_at LIMIT $2 FOR UPDATE OF n SKIP LOCKED`, mailMaxAge.Seconds(), mailBatch, auth.UserActive, mailMaxAttempts)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (pendingNote, error) {
		var p pendingNote
		err := row.Scan(&p.id, &p.message, &p.to)
		return p, err
	})
}

// message builds a notification e-mail: the plain-text body the app already
// showed, plus its HTML twin. Line breaks are stripped from the body's first
// line so nothing in a notification can add a header.
func (m *Mailer) message(to []string, body string) []byte {
	clean := strings.NewReplacer("\r", " ", "\n", " ").Replace
	text := clean(body)
	html := mailtmpl.Render(mailtmpl.Page{
		Subject:    mailSubject,
		Preheader:  text,
		Paragraphs: mailtmpl.Split(body),
		LogoURL:    m.LogoURL,
	})
	return m.compose(to, mailSubject, text, html)
}

// compose builds a multipart e-mail: the plain-text body every client reads
// plus an HTML twin. The subject cannot carry a line break.
func (m *Mailer) compose(to []string, subject, text, html string) []byte {
	subject = strings.NewReplacer("\r", " ", "\n", " ").Replace(subject)
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		copy(raw, []byte("chaintrack-fallback-boundary"))
	}
	boundary := hex.EncodeToString(raw)
	return []byte("From: " + m.FromHeader + "\r\n" +
		"To: " + strings.Join(to, ", ") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n" +
		"--" + boundary + "\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" + text + "\r\n" +
		"--" + boundary + "\r\nContent-Type: text/html; charset=utf-8\r\n\r\n" + html + "\r\n" +
		"--" + boundary + "--\r\n")
}

// SendOne e-mails one person now, outside the notification queue (an invitation).
func (m *Mailer) SendOne(to, subject, text, html string) error {
	return m.Send(m.Addr, m.Auth, m.From, []string{to}, m.compose([]string{to}, subject, text, html))
}
