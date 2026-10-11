package org

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/dbx"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrSlugTaken   = errors.New("slug taken")
	ErrEmailTaken  = errors.New("email taken")
	ErrNoDocuments = errors.New("no documents")
	ErrNotIssuer   = errors.New("not an issuer")
)

// Unique indexes from the migrations, used to tell which value collided.
const (
	uniqueSlugIndex  = "organisations_slug_key"
	uniqueEmailIndex = "users_email_key"
)

type Organisation struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	ReviewedBy   *string    `json:"reviewed_by"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewReason *string    `json:"review_reason"`
	// SigningAddress identifies the organisation on-chain; set once it is verified.
	SigningAddress *string `json:"signing_address"`
	// Accredited organisations may issue certificates.
	Accredited bool      `json:"accredited"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Document struct {
	ID             string    `json:"id"`
	OrganisationID string    `json:"organisation_id"`
	Type           string    `json:"type"`
	FileName       string    `json:"file_name"`
	ContentType    string    `json:"content_type"`
	SizeBytes      int64     `json:"size_bytes"`
	SHA256         string    `json:"sha256"`
	StorageKey     string    `json:"-"`
	UploadedBy     string    `json:"uploaded_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// What each list endpoint lets callers search and sort by.
var (
	OrgList = listq.Config{
		Alias:  "o",
		Search: []string{"name", "slug"},
		Sort:   []string{"name", "slug", "kind", "status", listq.ColCreatedAt},
	}
	DocList = listq.Config{
		Alias:  "d",
		Search: []string{"file_name", "type"},
		Sort:   []string{"file_name", "type", "size_bytes", listq.ColCreatedAt},
	}
)

const (
	orgColumns = `o.id::text, o.name, o.slug, o.kind, o.status, o.reviewed_by::text,
	              o.reviewed_at, o.review_reason, o.signing_address, o.accredited, o.created_at, o.updated_at`
	docColumns = `d.id::text, d.organisation_id::text, d.type, d.file_name, d.content_type,
	              d.size_bytes, d.sha256, d.storage_key, d.uploaded_by::text, d.created_at, d.updated_at`
)

func scanOrg(row pgx.CollectableRow) (Organisation, error) {
	var o Organisation
	err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.Kind, &o.Status, &o.ReviewedBy,
		&o.ReviewedAt, &o.ReviewReason, &o.SigningAddress, &o.Accredited, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func scanDoc(row pgx.CollectableRow) (Document, error) {
	var d Document
	err := row.Scan(&d.ID, &d.OrganisationID, &d.Type, &d.FileName, &d.ContentType,
		&d.SizeBytes, &d.SHA256, &d.StorageKey, &d.UploadedBy, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

type Store struct {
	Pool *pgxpool.Pool
}

// Register creates the organisation (pending), its first user, and that
// user's org_admin scope, all or nothing.
func (s Store) Register(ctx context.Context, in Registration, passwordHash string) (Organisation, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Organisation{}, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`INSERT INTO organisations AS o (name, slug, kind, status) VALUES ($1, $2, $3, $4)
		 RETURNING `+orgColumns, in.Name, in.Slug, in.Kind, StatusPending)
	if err != nil {
		return Organisation{}, err
	}
	o, err := pgx.CollectExactlyOneRow(rows, scanOrg)
	if err != nil {
		return Organisation{}, uniqueErr(err)
	}

	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, status) VALUES ($1, $2, $3, $4)
		 RETURNING id::text`, in.Email, passwordHash, in.UserName, auth.UserActive).Scan(&userID)
	if err != nil {
		return Organisation{}, uniqueErr(err)
	}

	tag, err := tx.Exec(ctx,
		`INSERT INTO userscope (user_id, role_id, organisation_id)
		 SELECT $1::uuid, id, $2::uuid FROM roles WHERE name = $3 AND NOT is_deleted`,
		userID, o.ID, auth.RoleOrgAdmin)
	if err != nil {
		return Organisation{}, err
	}
	if tag.RowsAffected() != 1 {
		return Organisation{}, errors.New("role " + auth.RoleOrgAdmin + " is missing: run SyncRoles")
	}
	return o, tx.Commit(ctx)
}

func uniqueErr(err error) error {
	switch constraint, _ := dbx.UniqueViolation(err); constraint {
	case uniqueSlugIndex:
		return ErrSlugTaken
	case uniqueEmailIndex:
		return ErrEmailTaken
	}
	return err
}

// List returns organisations visible through the gate, optionally by status.
func (s Store) List(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Organisation], error) {
	where := []string{auth.GateSQL("o.id", 1), `($2 = '' OR o.status = $2)`}
	return listq.Fetch(ctx, s.Pool, q, OrgList, "SELECT "+orgColumns, "FROM organisations o",
		where, []any{gate, status}, scanOrg)
}

func (s Store) Get(ctx context.Context, id, gate string) (Organisation, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+orgColumns+` FROM organisations o
		 WHERE o.id = $1::uuid AND NOT o.is_deleted AND `+auth.GateSQL("o.id", 2), id, gate)
	if err != nil {
		return Organisation{}, err
	}
	o, err := pgx.CollectExactlyOneRow(rows, scanOrg)
	return o, notFound(err)
}

func (s Store) ListDocuments(ctx context.Context, q listq.Query, orgID string) (listq.Page[Document], error) {
	return listq.Fetch(ctx, s.Pool, q, DocList, "SELECT "+docColumns, "FROM documents d",
		[]string{`d.organisation_id = $1::uuid`}, []any{orgID}, scanDoc)
}

func (s Store) GetDocument(ctx context.Context, id, gate string) (Document, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT `+docColumns+` FROM documents d
		 WHERE d.id = $1::uuid AND NOT d.is_deleted AND `+auth.GateSQL("d.organisation_id", 2), id, gate)
	if err != nil {
		return Document{}, err
	}
	d, err := pgx.CollectExactlyOneRow(rows, scanDoc)
	return d, notFound(err)
}

// AddDocument records an uploaded file. A rejected organisation that submits
// a new document goes back to pending; statusChanged reports that.
func (s Store) AddDocument(ctx context.Context, d Document) (doc Document, statusChanged bool, err error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Document{}, false, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx,
		`INSERT INTO documents AS d (organisation_id, type, file_name, content_type, size_bytes,
		                             sha256, storage_key, uploaded_by)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid)
		 RETURNING `+docColumns,
		d.OrganisationID, d.Type, d.FileName, d.ContentType, d.SizeBytes, d.SHA256, d.StorageKey, d.UploadedBy)
	if err != nil {
		return Document{}, false, err
	}
	if doc, err = pgx.CollectExactlyOneRow(rows, scanDoc); err != nil {
		return Document{}, false, err
	}

	tag, err := tx.Exec(ctx,
		`UPDATE organisations SET status = $2, reviewed_by = NULL, reviewed_at = NULL, review_reason = NULL
		 WHERE id = $1::uuid AND status = $3 AND NOT is_deleted`,
		d.OrganisationID, StatusPending, StatusRejected)
	if err != nil {
		return Document{}, false, err
	}
	return doc, tag.RowsAffected() == 1, tx.Commit(ctx)
}

// NewKey is a freshly generated signing key, ready to store.
type NewKey struct {
	Address string // public
	Sealed  string // the private key, sealed
}

// Review records a reviewer's decision. Verifying needs at least one document.
//
// The first time an organisation is verified it receives key as its signing
// key; if it already has one, key is ignored, so its on-chain identity never
// changes. Every change the chain must know about is queued in the same
// transaction.
func (s Store) Review(ctx context.Context, id string, rv Review, reviewerID string, key NewKey) (Organisation, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Organisation{}, err
	}
	defer tx.Rollback(ctx)

	// Lock the organisation so a concurrent upload or review cannot interleave.
	var (
		was     string
		address *string
	)
	err = tx.QueryRow(ctx,
		`SELECT status, signing_address FROM organisations WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, id,
	).Scan(&was, &address)
	if err != nil {
		return Organisation{}, notFound(err)
	}
	hasKey := address != nil

	if rv.Decision == StatusVerified {
		var docs int
		err = tx.QueryRow(ctx,
			`SELECT count(*) FROM documents WHERE organisation_id = $1::uuid AND NOT is_deleted`, id).Scan(&docs)
		if err != nil {
			return Organisation{}, err
		}
		if docs == 0 {
			return Organisation{}, ErrNoDocuments
		}
		if !hasKey {
			if _, err := tx.Exec(ctx,
				`UPDATE organisations SET signing_address = $2, signing_key = $3 WHERE id = $1::uuid`,
				id, key.Address, key.Sealed); err != nil {
				return Organisation{}, err
			}
			if err := chain.Enqueue(ctx, tx, chain.KindOrgRegister, id, chain.OrgRegister{OrgID: id, Signer: key.Address}); err != nil {
				return Organisation{}, err
			}
		}
	}

	// Tell the chain when the verified flag actually flips.
	nowVerified, wasVerified := rv.Decision == StatusVerified, was == StatusVerified
	if nowVerified != wasVerified && (nowVerified || hasKey) {
		if err := chain.Enqueue(ctx, tx, chain.KindOrgVerified, id, chain.OrgVerified{OrgID: id, Verified: nowVerified}); err != nil {
			return Organisation{}, err
		}
	}

	rows, err := tx.Query(ctx,
		`UPDATE organisations AS o
		 SET status = $2, reviewed_by = $3::uuid, reviewed_at = now(), review_reason = NULLIF($4, '')
		 WHERE o.id = $1::uuid RETURNING `+orgColumns, id, rv.Decision, reviewerID, rv.Reason)
	if err != nil {
		return Organisation{}, err
	}
	o, err := pgx.CollectExactlyOneRow(rows, scanOrg)
	if err != nil {
		return Organisation{}, err
	}
	return o, tx.Commit(ctx)
}

// Accredit allows (or stops) an organisation issuing certificates, and queues
// the change for the chain. Only a verified certifier or regulator qualifies.
func (s Store) Accredit(ctx context.Context, id string, accredited bool) (Organisation, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Organisation{}, err
	}
	defer tx.Rollback(ctx)

	var (
		kind, status string
		was          bool
	)
	err = tx.QueryRow(ctx,
		`SELECT kind, status, accredited FROM organisations WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, id,
	).Scan(&kind, &status, &was)
	if err != nil {
		return Organisation{}, notFound(err)
	}
	if accredited && (status != StatusVerified || !slices.Contains(IssuerKinds, kind)) {
		return Organisation{}, ErrNotIssuer
	}
	// Tell the chain only when the flag actually flips.
	if accredited != was {
		if err := chain.Enqueue(ctx, tx, chain.KindOrgAccredited, id, chain.OrgAccredited{OrgID: id, Accredited: accredited}); err != nil {
			return Organisation{}, err
		}
	}
	rows, err := tx.Query(ctx,
		`UPDATE organisations AS o SET accredited = $2 WHERE o.id = $1::uuid RETURNING `+orgColumns, id, accredited)
	if err != nil {
		return Organisation{}, err
	}
	o, err := pgx.CollectExactlyOneRow(rows, scanOrg)
	if err != nil {
		return Organisation{}, err
	}
	return o, tx.Commit(ctx)
}
