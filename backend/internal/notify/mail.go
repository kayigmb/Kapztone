package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

// Mailer sends the e-mail copy of each notification to the organisation's
// active users.
//
// ponytail: one loop in the API process, at-least-once: a crash between
// sending and marking sends that e-mail again. Move to its own worker with a
// provider's idempotency key if duplicates ever matter.
type Mailer struct {
	Pool *pgxpool.Pool
	Addr string // host:port
	Auth smtp.Auth
	From string
	// Send delivers one message; NewMailer sets it, tests replace it.
	Send func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// SMTP is a mail server and the account notifications are sent from.
type SMTP struct {
	Host     string
	Port     string
	User     string // also the sender address
	Password string // empty: the server needs no sign-in
	// TLS requires the connection to be encrypted (STARTTLS). Without it a
	// server that offers no encryption is written to in plain text.
	TLS bool
}

func NewMailer(pool *pgxpool.Pool, s SMTP) *Mailer {
	m := &Mailer{Pool: pool, Addr: net.JoinHostPort(s.Host, s.Port), From: s.User}
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
// ponytail: STARTTLS only (port 587); add tls.Dial if a provider needs port 465.
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

	// SKIP LOCKED: two API processes never pick the same notification.
	rows, err := tx.Query(ctx,
		`SELECT n.id::text, n.message,
		        ARRAY(SELECT u.email::text FROM userscope s JOIN users u ON u.id = s.user_id
		              WHERE s.organisation_id IS NOT DISTINCT FROM n.organisation_id AND NOT s.is_deleted
		                AND NOT u.is_deleted AND u.status = $3)
		 FROM notifications n
		 WHERE n.emailed_at IS NULL AND n.created_at > now() - make_interval(secs => $1) AND NOT n.is_deleted
		 ORDER BY n.created_at LIMIT $2 FOR UPDATE OF n SKIP LOCKED`, mailMaxAge.Seconds(), mailBatch, auth.UserActive)
	if err != nil {
		return err
	}
	type pending struct {
		id, message string
		to          []string
	}
	var batch []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.message, &p.to); err != nil {
			rows.Close()
			return err
		}
		batch = append(batch, p)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, p := range batch {
		if len(p.to) > 0 {
			if err := m.Send(m.Addr, m.Auth, m.From, p.to, m.message(p.to, p.message)); err != nil {
				return err // nothing marked; the whole batch is tried again
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE notifications SET emailed_at = now() WHERE id = $1::uuid`, p.id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// message builds a plain-text e-mail. Line breaks are stripped from the
// body's first line so nothing in a notification can add a header.
func (m *Mailer) message(to []string, body string) []byte {
	clean := strings.NewReplacer("\r", " ", "\n", " ").Replace
	return m.compose(to, mailSubject, clean(body))
}

// compose builds a plain-text e-mail; the subject cannot carry a line break.
func (m *Mailer) compose(to []string, subject, body string) []byte {
	subject = strings.NewReplacer("\r", " ", "\n", " ").Replace(subject)
	return []byte("From: " + m.From + "\r\n" +
		"To: " + strings.Join(to, ", ") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" +
		body + "\r\n")
}

// SendOne e-mails one person now, outside the notification queue (an invitation).
func (m *Mailer) SendOne(to, subject, body string) error {
	return m.Send(m.Addr, m.Auth, m.From, []string{to}, m.compose([]string{to}, subject, body))
}
