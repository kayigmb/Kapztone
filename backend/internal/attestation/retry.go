package attestation

import (
	"context"
	"io"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// Retrying failed IPFS copies: every RetryEvery the certificates whose copy
// failed (and so have an ipfs_next_at) are pinned again once it is due. Each failure doubles the wait, from
// RetryFirst up to RetryLongest; after RetryMaxAttempts it stops (the file
// stays in the bucket and its hash on the chain, which is what proves it).
const (
	RetryEvery       = time.Minute
	RetryFirst       = time.Minute
	RetryLongest     = 6 * time.Hour
	RetryMaxAttempts = 10
	retryBatch       = 10
)

// Backoff is how long to wait after the given number of failed attempts.
func Backoff(attempts int) time.Duration {
	d := RetryFirst
	for i := 1; i < attempts && d < RetryLongest; i++ {
		d *= 2
	}
	return min(d, RetryLongest)
}

// PinFunc copies a file to IPFS and returns its content id (Pinata.Pin).
type PinFunc func(ctx context.Context, name string, content io.Reader) (string, error)

// Retrier pins certificates whose first copy to IPFS failed.
type Retrier struct {
	Store Store
	Files *upload.Bucket
	Pin   PinFunc
	Now   func() time.Time // nil: time.Now
}

func (r *Retrier) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

// Run retries until ctx ends.
func (r *Retrier) Run(ctx context.Context) {
	t := time.NewTicker(RetryEvery)
	defer t.Stop()
	for {
		if n, err := r.RetryDue(ctx); err != nil {
			logx.From(ctx).Error("ipfs retry round failed", logx.KeyError, err)
		} else if n > 0 {
			logx.From(ctx).Info("ipfs retry round", "tried", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// RetryDue tries each due certificate once and reports how many it tried.
// Rows are claimed with SKIP LOCKED, so two API processes never pin the same one.
func (r *Retrier) RetryDue(ctx context.Context) (int, error) {
	tx, err := r.Store.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx,
		`SELECT id::text, file_name, storage_key, ipfs_attempts FROM attestations
		 WHERE ipfs_cid IS NULL AND NOT is_deleted AND ipfs_attempts < $1 AND ipfs_next_at <= $2
		 ORDER BY ipfs_next_at LIMIT $3 FOR UPDATE SKIP LOCKED`, RetryMaxAttempts, r.now(), retryBatch)
	if err != nil {
		return 0, err
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (duePin, error) {
		var d duePin
		return d, row.Scan(&d.id, &d.name, &d.key, &d.attempts)
	})
	if err != nil {
		return 0, err
	}
	for _, d := range list {
		if err := r.retry(ctx, tx, d); err != nil {
			return 0, err
		}
	}
	return len(list), tx.Commit(ctx)
}

type duePin struct {
	id, name, key string
	attempts      int
}

// retry pins one certificate again; a failure pushes its next try back.
func (r *Retrier) retry(ctx context.Context, tx pgx.Tx, d duePin) error {
	cid, pinErr := r.pinOne(ctx, d.name, d.key)
	log := logx.From(ctx).With("attestation_id", d.id, "attempt", d.attempts+1)
	if pinErr == nil {
		_, err := tx.Exec(ctx, `UPDATE attestations SET ipfs_cid = $2, ipfs_last_error = NULL WHERE id = $1::uuid`, d.id, cid)
		if err == nil {
			log.Info("ipfs pinned on retry", "cid", cid)
		}
		return err
	}
	next := r.now().Add(Backoff(d.attempts + 1))
	_, err := tx.Exec(ctx,
		`UPDATE attestations SET ipfs_attempts = ipfs_attempts + 1, ipfs_next_at = $2, ipfs_last_error = $3 WHERE id = $1::uuid`,
		d.id, next, pinErr.Error())
	if err == nil {
		log.Warn("ipfs pin failed again", "next_at", next, logx.KeyError, pinErr)
	}
	return err
}

func (r *Retrier) pinOne(ctx context.Context, name, key string) (string, error) {
	f, err := r.Files.Open(ctx, key)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return r.Pin(ctx, name, f)
}

// RecordPinFailure notes the first failed copy, so the retrier picks it up after the first wait.
func (s Store) RecordPinFailure(ctx context.Context, id string, at time.Time, err error) error {
	_, dbErr := s.Pool.Exec(ctx,
		`UPDATE attestations SET ipfs_attempts = 1, ipfs_next_at = $2, ipfs_last_error = $3 WHERE id = $1::uuid`,
		id, at.Add(Backoff(1)), err.Error())
	return dbErr
}
