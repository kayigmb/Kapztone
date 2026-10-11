package chain

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

// The payloads below are what each kind of queued item carries. Ids are
// database uuids and hashes are 0x-hex; the relayer converts them.

type OrgRegister struct {
	OrgID  string `json:"org_id"`
	Signer string `json:"signer"` // the organisation's signing address
}

type OrgVerified struct {
	OrgID    string `json:"org_id"`
	Verified bool   `json:"verified"`
}

type FacilityRegister struct {
	FacilityID string `json:"facility_id"`
	OrgID      string `json:"org_id"`
	DataHash   string `json:"data_hash"`
	Capacity   int64  `json:"capacity"`
}

type FacilitySuspended struct {
	FacilityID string `json:"facility_id"`
	Suspended  bool   `json:"suspended"`
}

type PassportIssue struct {
	PassportID string `json:"passport_id"`
	Commodity  string `json:"commodity"`
	FacilityID string `json:"facility_id"`
	Period     int64  `json:"period"`
	Quantity   int64  `json:"quantity"`
	DataHash   string `json:"data_hash"`
}

type HandoverSigned struct {
	HandoverID string `json:"handover_id"`
	PassportID string `json:"passport_id"`
	FromOrg    string `json:"from_org"`
	ToOrg      string `json:"to_org"`
	Quantity   int64  `json:"quantity"`
	DataHash   string `json:"data_hash"`
	Nonce      int64  `json:"nonce"`
	Deadline   int64  `json:"deadline"` // unix seconds
	SigFrom    string `json:"sig_from"` // 0x-hex
	SigTo      string `json:"sig_to"`
}

type MaxYieldSet struct {
	Commodity       string `json:"commodity"`
	Stage           string `json:"stage"`
	PartsPerMillion int64  `json:"parts_per_million"`
}

type OrgAccredited struct {
	OrgID      string `json:"org_id"`
	Accredited bool   `json:"accredited"`
}

type TransformOutput struct {
	PassportID string `json:"passport_id"`
	Quantity   int64  `json:"quantity"`
	DataHash   string `json:"data_hash"`
}

type Transform struct {
	TransformID string            `json:"transform_id"`
	Inputs      []string          `json:"inputs"` // passport ids
	InputsUsed  []int64           `json:"inputs_used"`
	Stage       string            `json:"stage"`
	Outputs     []TransformOutput `json:"outputs"`
	DataHash    string            `json:"data_hash"`
}

type AttestationIssue struct {
	AttestationID string `json:"attestation_id"`
	IssuerOrg     string `json:"issuer_org"`
	SubjectID     string `json:"subject_id"`
	DocHash       string `json:"doc_hash"`
	ValidFrom     int64  `json:"valid_from"` // unix seconds
	ValidTo       int64  `json:"valid_to"`
}

type AttestationRevoke struct {
	AttestationID string `json:"attestation_id"`
}

// PassportReason is the payload of a flag and of a revocation: the passport
// and the hash of the reason record.
type PassportReason struct {
	PassportID string `json:"passport_id"`
	ReasonHash string `json:"reason_hash"`
}

// Enqueue queues a change for the chain. It must be called with the database
// transaction that makes the change itself, so the two are saved together or
// not at all: nothing is ever changed without being queued, or queued without
// having happened.
func Enqueue(ctx context.Context, tx pgx.Tx, kind, subjectID string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO chain_outbox (kind, subject_id, payload, status) VALUES ($1, $2::uuid, $3::jsonb, $4)`,
		kind, subjectID, string(body), StatusPending)
	return err
}
