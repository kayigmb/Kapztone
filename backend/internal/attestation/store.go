package attestation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrNotAccredited   = errors.New("not accredited")
	ErrSubjectNotFound = errors.New("subject not found")
	ErrAlreadyRevoked  = errors.New("already revoked")
)

type Attestation struct {
	ID          string    `json:"id"`
	IssuerOrgID string    `json:"issuer_org_id"`
	SubjectType string    `json:"subject_type"`
	SubjectID   string    `json:"subject_id"`
	Type        string    `json:"type"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	// DocHash is the SHA-256 of the file, the value anchored on-chain.
	DocHash     string `json:"doc_hash"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	StorageKey  string `json:"-"`
	// IPFSCID is set when the file was also pinned to IPFS.
	IPFSCID      *string    `json:"ipfs_cid"`
	Status       string     `json:"status"`
	RevokeReason *string    `json:"revoke_reason"`
	RevokedAt    *time.Time `json:"revoked_at"`
	TxHash       *string    `json:"tx_hash"`
	RevokeTxHash *string    `json:"revoke_tx_hash"`
	AnchoredAt   *time.Time `json:"anchored_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// List is what the list endpoint lets callers search and sort by.
var List = listq.Config{
	Alias:  "a",
	Search: []string{"type", "file_name"},
	Sort:   []string{"type", "valid_from", "valid_to", listq.ColCreatedAt},
}

const (
	// A valid certificate past valid_to reads as expired ($%[1]d = valid, $%[2]d = expired).
	shownStatusFmt = `CASE WHEN a.status = $%[1]d AND a.valid_to < now() THEN $%[2]d ELSE a.status END`
	columnsFmt     = `a.id::text, a.issuer_org_id::text, a.subject_type, a.subject_id::text, a.type,
	                  a.valid_from, a.valid_to, a.doc_hash, a.file_name, a.content_type, a.size_bytes,
	                  a.storage_key, a.ipfs_cid, ` + shownStatusFmt + `, a.revoke_reason, a.revoked_at,
	                  a.tx_hash, a.revoke_tx_hash, a.anchored_at, a.created_at, a.updated_at`
)

func columns(validArg, expiredArg int) string { return fmt.Sprintf(columnsFmt, validArg, expiredArg) }

func scan(row pgx.CollectableRow) (Attestation, error) {
	var a Attestation
	err := row.Scan(&a.ID, &a.IssuerOrgID, &a.SubjectType, &a.SubjectID, &a.Type, &a.ValidFrom, &a.ValidTo,
		&a.DocHash, &a.FileName, &a.ContentType, &a.SizeBytes, &a.StorageKey, &a.IPFSCID, &a.Status,
		&a.RevokeReason, &a.RevokedAt, &a.TxHash, &a.RevokeTxHash, &a.AnchoredAt, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func one(rows pgx.Rows, err error) (Attestation, error) {
	if err != nil {
		return Attestation{}, err
	}
	a, err := pgx.CollectExactlyOneRow(rows, scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attestation{}, ErrNotFound
	}
	return a, err
}

type Store struct {
	Pool *pgxpool.Pool
}

// tellOwner notifies whoever holds the subject, unless they are the issuer.
func tellOwner(ctx context.Context, tx pgx.Tx, a Attestation, kind, message string) error {
	var owner string
	// The table name comes from subjectTable, never from data.
	err := tx.QueryRow(ctx,
		`SELECT organisation_id::text FROM `+subjectTable[a.SubjectType]+` WHERE id = $1::uuid`, a.SubjectID).Scan(&owner)
	if err != nil || owner == a.IssuerOrgID {
		return err
	}
	return notify.Add(ctx, tx, owner, kind, a.ID, message)
}

// Issue records a certificate and queues it for the chain. The issuer must
// be accredited and the subject must exist; the contract checks the first again.
func (s Store) Issue(ctx context.Context, orgID, userID string, in Input, f upload.File) (Attestation, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Attestation{}, err
	}
	defer tx.Rollback(ctx)

	var accredited, exists bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM organisations WHERE id = $1::uuid AND status = $2 AND accredited AND NOT is_deleted),
		        EXISTS (SELECT 1 FROM `+subjectTable[in.SubjectType]+` WHERE id = $3::uuid AND NOT is_deleted)`,
		orgID, auth.OrgVerified, in.SubjectID).Scan(&accredited, &exists)
	switch {
	case err != nil:
		return Attestation{}, err
	case !accredited:
		return Attestation{}, ErrNotAccredited
	case !exists:
		return Attestation{}, ErrSubjectNotFound
	}

	docHash := "0x" + f.SHA256
	a, err := one(tx.Query(ctx,
		`INSERT INTO attestations AS a (issuer_org_id, subject_type, subject_id, type, valid_from, valid_to, doc_hash,
		                                storage_key, file_name, content_type, size_bytes, status, issued_by)
		 VALUES ($3::uuid, $4, $5::uuid, $6, $7, $8, $9, $10, $11, $12, $13, $1, $14::uuid)
		 RETURNING `+columns(1, 2),
		StatusValid, StatusExpired, orgID, in.SubjectType, in.SubjectID, in.Type, in.ValidFrom, in.ValidTo, docHash,
		f.Key, f.Name, f.ContentType, f.Size, userID))
	if err != nil {
		return Attestation{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindAttestationIssue, a.ID, chain.AttestationIssue{
		AttestationID: a.ID, IssuerOrg: orgID, SubjectID: in.SubjectID, DocHash: docHash,
		ValidFrom: in.ValidFrom.Unix(), ValidTo: in.ValidTo.Unix(),
	})
	if err != nil {
		return Attestation{}, err
	}
	if err := tellOwner(ctx, tx, a, notify.KindCertificateIssued, MsgNotifyIssued); err != nil {
		return Attestation{}, err
	}
	return a, tx.Commit(ctx)
}

// SetCID records where the file was pinned on IPFS.
func (s Store) SetCID(ctx context.Context, id, cid string) (Attestation, error) {
	return one(s.Pool.Query(ctx,
		`UPDATE attestations AS a SET ipfs_cid = $4 WHERE a.id = $3::uuid RETURNING `+columns(1, 2),
		StatusValid, StatusExpired, id, cid))
}

// Find returns certificates, optionally those about one subject or in one status.
func (s Store) Find(ctx context.Context, q listq.Query, subjectID, status string) (listq.Page[Attestation], error) {
	// The status filter is applied to the status as read, so "expired" works.
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns(1, 2), "FROM attestations a",
		[]string{
			`($3 = '' OR a.subject_id = NULLIF($3, '')::uuid)`,
			`($4 = '' OR ` + fmt.Sprintf(shownStatusFmt, 1, 2) + ` = $4)`,
		},
		[]any{StatusValid, StatusExpired, subjectID, status}, scan)
}

func (s Store) Get(ctx context.Context, id string) (Attestation, error) {
	return one(s.Pool.Query(ctx,
		`SELECT `+columns(1, 2)+` FROM attestations a WHERE a.id = $3::uuid AND NOT a.is_deleted`,
		StatusValid, StatusExpired, id))
}

// Revoke withdraws a certificate for good. Only its issuer (or the platform)
// may; gate is the caller's data gate.
func (s Store) Revoke(ctx context.Context, id, gate, userID, reason string) (Attestation, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Attestation{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx,
		`SELECT a.status FROM attestations a
		 WHERE a.id = $1::uuid AND NOT a.is_deleted AND `+auth.GateSQL("a.issuer_org_id", 2)+` FOR UPDATE`,
		id, gate).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attestation{}, ErrNotFound
	}
	if err != nil {
		return Attestation{}, err
	}
	if status == StatusRevoked {
		return Attestation{}, ErrAlreadyRevoked
	}

	a, err := one(tx.Query(ctx,
		`UPDATE attestations AS a SET status = $3, revoke_reason = $4, revoked_by = $5::uuid, revoked_at = now()
		 WHERE a.id = $6::uuid RETURNING `+columns(1, 2),
		StatusValid, StatusExpired, StatusRevoked, reason, userID, id))
	if err != nil {
		return Attestation{}, err
	}
	if err := chain.Enqueue(ctx, tx, chain.KindAttestationRevoke, id, chain.AttestationRevoke{AttestationID: id}); err != nil {
		return Attestation{}, err
	}
	if err := tellOwner(ctx, tx, a, notify.KindCertificateRevoked, MsgNotifyRevoked); err != nil {
		return Attestation{}, err
	}
	return a, tx.Commit(ctx)
}

// ForSubjects returns every certificate about any of the given passports or facilities.
func (s Store) ForSubjects(ctx context.Context, subjectIDs []string) ([]Attestation, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+columns(1, 2)+` FROM attestations a
		 WHERE a.subject_id = ANY($3::uuid[]) AND NOT a.is_deleted ORDER BY a.created_at, a.id`,
		StatusValid, StatusExpired, subjectIDs)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scan)
}
