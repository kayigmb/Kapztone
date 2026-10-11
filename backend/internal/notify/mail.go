package notify

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
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
	// Send is smtp.SendMail unless a test replaces it.
	Send func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// NewMailer reads a mail server address written as
// smtp://user:password@host:port?from=no-reply@example.com
// (user and password are optional).
func NewMailer(pool *pgxpool.Pool, smtpURL string) (*Mailer, error) {
	u, err := url.Parse(smtpURL)
	if err != nil || u.Scheme != "smtp" || u.Hostname() == "" || u.Port() == "" || u.Query().Get("from") == "" {
		return nil, fmt.Errorf("must look like smtp://user:password@host:port?from=address")
	}
	m := &Mailer{Pool: pool, Addr: u.Host, From: u.Query().Get("from"), Send: smtp.SendMail}
	if user := u.User.Username(); user != "" {
		password, _ := u.User.Password()
		m.Auth = smtp.PlainAuth("", user, password, u.Hostname())
	}
	return m, nil
}

// Run sends until ctx ends. A mail server that cannot be reached is not an
// error to give up on: notifications wait and stay visible in the app.
func (m *Mailer) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := m.SendPending(ctx); err != nil {
			log.Printf("mailer: %v (will retry)", err)
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
		              WHERE s.organisation_id = n.organisation_id AND NOT s.is_deleted
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
	return []byte("From: " + m.From + "\r\n" +
		"To: " + strings.Join(to, ", ") + "\r\n" +
		"Subject: " + mailSubject + "\r\n" +
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" +
		clean(body) + "\r\n")
}
