package passport

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrFacilityNotFound  = errors.New("facility not found")
	ErrFacilitySuspended = errors.New("facility suspended")
	ErrTemplateNotFound  = errors.New("template not found")
	ErrKindMismatch      = errors.New("facility kind mismatch")
	ErrUnitMismatch      = errors.New("unit mismatch")
	ErrNotHolder         = errors.New("not holder")
	ErrNotActive         = errors.New("passport not active")
	ErrOpenHandover      = errors.New("open handover exists")
	ErrReceiverInvalid   = errors.New("receiver invalid")
	ErrHandoverClosed    = errors.New("handover closed")
	ErrHandoverStale     = errors.New("handover stale")
	ErrHandoverStage     = errors.New("handover stage invalid")
)

// EvidenceError lists the evidence a stage asks for that has no valid
// certificate yet.
type EvidenceError struct {
	Stage   string
	Missing []string
}

func (e EvidenceError) Error() string {
	return fmt.Sprintf("stage %s needs evidence that is missing: %s. Attach a valid certificate of each first",
		e.Stage, strings.Join(e.Missing, ", "))
}

// CapacityError says by how much an issue would exceed the facility's
// capacity for the period, and the most that can still be issued.
type CapacityError struct {
	Requested, Available int64
	Unit                 string
}

func (e CapacityError) Error() string {
	return fmt.Sprintf("capacity exceeded by %d %s; at most %d %s can still be issued from this facility in this period",
		e.Requested-e.Available, e.Unit, e.Available, e.Unit)
}

type Passport struct {
	ID string `json:"id"`
	// Code is the readable lot code (ACME-2026-0412); it works wherever the id does.
	Code           string `json:"code"`
	OrganisationID string `json:"organisation_id"` // current holder
	// FacilityID and Period are set on a passport issued at an origin
	// facility; TransformID on one made by a transform. Never both.
	FacilityID    *string    `json:"facility_id"`
	TransformID   *string    `json:"transform_id"`
	TemplateID    string     `json:"template_id"`
	Commodity     string     `json:"commodity"`
	Unit          string     `json:"unit"`
	Quantity      int64      `json:"quantity"`
	Remaining     int64      `json:"remaining"`
	Period        *int64     `json:"period"`
	Status        string     `json:"status"`
	DataHash      string     `json:"data_hash"`
	HandoverCount int64      `json:"handover_count"`
	TxHash        *string    `json:"tx_hash"`
	AnchoredAt    *time.Time `json:"anchored_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// Events is the journey so far; filled when one passport is fetched.
	Events []Event `json:"events,omitempty"`
}

type Event struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	FromOrg   *string   `json:"from_org"`
	ToOrg     *string   `json:"to_org"`
	Quantity  int64     `json:"quantity"`
	DataHash  string    `json:"data_hash"`
	TxHash    *string   `json:"tx_hash"`
	CreatedAt time.Time `json:"created_at"`
}

type Handover struct {
	ID           string     `json:"id"`
	PassportID   string     `json:"passport_id"`
	FromOrg      string     `json:"from_org"`
	ToOrg        string     `json:"to_org"`
	Stage        string     `json:"stage"` // the template stage the lot enters
	Quantity     int64      `json:"quantity"`
	Note         string     `json:"note"`
	DataHash     string     `json:"data_hash"`
	Deadline     time.Time  `json:"deadline"`
	Status       string     `json:"status"`
	SignedByFrom bool       `json:"signed_by_sender"`
	SignedByTo   bool       `json:"signed_by_receiver"`
	RejectReason *string    `json:"reject_reason"`
	RespondedAt  *time.Time `json:"responded_at"`
	TxHash       *string    `json:"tx_hash"`
	AnchoredAt   *time.Time `json:"anchored_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// What each list endpoint lets callers search and sort by.
var (
	List = listq.Config{
		Alias:  "p",
		Search: []string{"commodity", "status"},
		Sort:   []string{"commodity", "quantity", "status", listq.ColCreatedAt},
	}
	HandoverList = listq.Config{
		Alias:  "h",
		Search: []string{"note", "status"},
		Sort:   []string{"status", "deadline", listq.ColCreatedAt},
	}
)

const (
	columns = `p.id::text, p.code, p.organisation_id::text, p.facility_id::text, p.transform_id::text, p.template_id::text, p.commodity,
	           p.unit, p.quantity, p.remaining, p.period, p.status, p.data_hash, p.handover_count,
	           p.tx_hash, p.anchored_at, p.created_at, p.updated_at`

	// An awaiting handover past its deadline reads as expired ($%[1]d = awaiting, $%[2]d = expired).
	handoverColumnsFmt = `h.id::text, h.passport_id::text, h.from_org::text, h.to_org::text, h.stage, h.quantity, h.note,
	           h.data_hash, h.deadline,
	           CASE WHEN h.status = $%[1]d AND h.deadline < now() THEN $%[2]d ELSE h.status END,
	           true, h.sig_to IS NOT NULL, h.reject_reason, h.responded_at, h.tx_hash, h.anchored_at,
	           h.created_at, h.updated_at`

	// visibleFmt: who may see a passport. Its holder, anyone it passed through,
	// and the receiver of a handover still awaiting an answer (they must be
	// able to inspect what they are asked to accept). $%[1]d is the gate.
	visibleFmt = `($%[1]d = '' OR p.organisation_id = NULLIF($%[1]d, '')::uuid
	  OR EXISTS (SELECT 1 FROM trace_events e WHERE e.passport_id = p.id AND NOT e.is_deleted
	             AND (e.from_org = NULLIF($%[1]d, '')::uuid OR e.to_org = NULLIF($%[1]d, '')::uuid))
	  OR EXISTS (SELECT 1 FROM handovers h WHERE h.passport_id = p.id AND NOT h.is_deleted
	             AND h.to_org = NULLIF($%[1]d, '')::uuid))`

	// partyFmt: who may see a handover: its sender and its receiver.
	partyFmt = `($%[1]d = '' OR h.from_org = NULLIF($%[1]d, '')::uuid OR h.to_org = NULLIF($%[1]d, '')::uuid)`
)

func scan(row pgx.CollectableRow) (Passport, error) {
	var p Passport
	err := row.Scan(&p.ID, &p.Code, &p.OrganisationID, &p.FacilityID, &p.TransformID, &p.TemplateID, &p.Commodity, &p.Unit,
		&p.Quantity, &p.Remaining, &p.Period, &p.Status, &p.DataHash, &p.HandoverCount,
		&p.TxHash, &p.AnchoredAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func scanHandover(row pgx.CollectableRow) (Handover, error) {
	var h Handover
	err := row.Scan(&h.ID, &h.PassportID, &h.FromOrg, &h.ToOrg, &h.Stage, &h.Quantity, &h.Note, &h.DataHash,
		&h.Deadline, &h.Status, &h.SignedByFrom, &h.SignedByTo, &h.RejectReason, &h.RespondedAt,
		&h.TxHash, &h.AnchoredAt, &h.CreatedAt, &h.UpdatedAt)
	return h, err
}

func scanEvent(row pgx.CollectableRow) (Event, error) {
	var e Event
	err := row.Scan(&e.ID, &e.Kind, &e.FromOrg, &e.ToOrg, &e.Quantity, &e.DataHash, &e.TxHash, &e.CreatedAt)
	return e, err
}

func one[T any](rows pgx.Rows, err error, scan pgx.RowToFunc[T]) (T, error) {
	var zero T
	if err != nil {
		return zero, err
	}
	v, err := pgx.CollectExactlyOneRow(rows, scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return zero, ErrNotFound
	}
	return v, err
}

func onePassport(rows pgx.Rows, err error) (Passport, error) { return one(rows, err, scan) }
func oneHandover(rows pgx.Rows, err error) (Handover, error) { return one(rows, err, scanHandover) }

func newID(ctx context.Context, tx pgx.Tx) (id string, err error) {
	err = tx.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&id)
	return id, err
}

type Store struct {
	Pool *pgxpool.Pool
	// Secrets opens organisations' signing keys.
	Secrets *secret.Box
	// Domain binds handover signatures to the deployed CustodyManager.
	Domain chain.Domain
}

// issueRecord is what a passport's data hash covers.
type issueRecord struct {
	ID             string `json:"id"`
	OrganisationID string `json:"organisation_id"`
	FacilityID     string `json:"facility_id"`
	TemplateID     string `json:"template_id"`
	Commodity      string `json:"commodity"`
	Unit           string `json:"unit"`
	Quantity       int64  `json:"quantity"`
	Period         int64  `json:"period"`
}

// Issue creates an origin passport after checking every rule the contract
// will check again: the facility is the organisation's and active, it suits
// the commodity, and its capacity for the period is not exceeded.
func (s Store) Issue(ctx context.Context, orgID, userID string, in IssueInput, now time.Time) (Passport, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Passport{}, err
	}
	defer tx.Rollback(ctx)

	// Locking the facility makes concurrent issues queue up, so two of them
	// cannot both fit into the last of the capacity.
	var fKind, fStatus, fUnit, fPeriod string
	var capacity int64
	err = tx.QueryRow(ctx,
		`SELECT kind, status, capacity, capacity_unit, capacity_period FROM facilities
		 WHERE id = $1::uuid AND organisation_id = $2::uuid AND NOT is_deleted FOR UPDATE`,
		in.FacilityID, orgID).Scan(&fKind, &fStatus, &capacity, &fUnit, &fPeriod)
	if errors.Is(err, pgx.ErrNoRows) {
		return Passport{}, ErrFacilityNotFound
	}
	if err != nil {
		return Passport{}, err
	}
	if fStatus != facility.StatusActive {
		return Passport{}, ErrFacilitySuspended
	}

	var commodity, unit, originKind, stagesJSON string
	err = tx.QueryRow(ctx,
		`SELECT commodity, unit, origin_facility_kind, stages::text FROM commodity_templates
		 WHERE id = $1::uuid AND NOT is_deleted`, in.TemplateID).Scan(&commodity, &unit, &originKind, &stagesJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return Passport{}, ErrTemplateNotFound
	}
	if err != nil {
		return Passport{}, err
	}
	if originKind != fKind {
		return Passport{}, ErrKindMismatch
	}
	if unit != fUnit {
		return Passport{}, ErrUnitMismatch
	}
	stages, err := parseStages(stagesJSON)
	if err != nil {
		return Passport{}, err
	}
	// The origin stage's evidence is about the facility the goods come from.
	for _, st := range stages {
		if slices.Contains(st.Events, EventIssue) {
			if err := requireEvidence(ctx, tx, st, attestation.SubjectFacility, in.FacilityID); err != nil {
				return Passport{}, err
			}
		}
	}

	period := facility.PeriodFor(fPeriod, now)
	used, err := facility.CapacityUsed(ctx, tx, []string{in.FacilityID}, []int64{period})
	if err != nil {
		return Passport{}, err
	}
	if available := capacity - used[in.FacilityID]; in.Quantity > available {
		return Passport{}, CapacityError{Requested: in.Quantity, Available: available, Unit: unit}
	}

	id, err := newID(ctx, tx)
	if err != nil {
		return Passport{}, err
	}
	hash, err := chain.DataHash(issueRecord{id, orgID, in.FacilityID, in.TemplateID, commodity, unit, in.Quantity, period})
	if err != nil {
		return Passport{}, err
	}
	dataHash := chain.Hex(hash)
	code, err := nextCode(ctx, tx, orgID, now)
	if err != nil {
		return Passport{}, err
	}

	p, err := onePassport(tx.Query(ctx,
		`INSERT INTO passports AS p (id, organisation_id, facility_id, template_id, commodity, unit,
		                             quantity, remaining, period, status, data_hash, issued_by, code)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $7, $8, $9, $10, $11::uuid, $12)
		 RETURNING `+columns,
		id, orgID, in.FacilityID, in.TemplateID, commodity, unit, in.Quantity, period, StatusActive, dataHash, userID, code))
	if err != nil {
		return Passport{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, to_org, quantity, data_hash)
		 VALUES ($1::uuid, $1::uuid, $2, $3::uuid, $4, $5)`, id, EventIssue, orgID, in.Quantity, dataHash); err != nil {
		return Passport{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindPassportIssue, id, chain.PassportIssue{
		PassportID: id, Commodity: commodity, FacilityID: in.FacilityID,
		Period: period, Quantity: in.Quantity, DataHash: dataHash,
	})
	if err != nil {
		return Passport{}, err
	}
	return p, tx.Commit(ctx)
}

func visible(n int) string { return fmt.Sprintf(visibleFmt, n) }
func party(n int) string   { return fmt.Sprintf(partyFmt, n) }

func handoverColumns(awaitingArg, expiredArg int) string {
	return fmt.Sprintf(handoverColumnsFmt, awaitingArg, expiredArg)
}

func (s Store) Find(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Passport], error) {
	return listq.Fetch(ctx, s.Pool, q, List, "SELECT "+columns, "FROM passports p",
		[]string{visible(1), `($2 = '' OR p.status = $2)`}, []any{gate, status}, scan)
}

// Get returns one passport with its journey.
func (s Store) Get(ctx context.Context, id, gate string) (Passport, error) {
	p, err := onePassport(s.Pool.Query(ctx,
		`SELECT `+columns+` FROM passports p WHERE p.id = $1::uuid AND NOT p.is_deleted AND `+visible(2), id, gate))
	if err != nil {
		return Passport{}, err
	}
	rows, err := s.Pool.Query(ctx,
		`SELECT id::text, kind, from_org::text, to_org::text, quantity, data_hash, tx_hash, created_at
		 FROM trace_events WHERE passport_id = $1::uuid AND NOT is_deleted ORDER BY created_at, id`, id)
	if err != nil {
		return Passport{}, err
	}
	p.Events, err = pgx.CollectRows(rows, scanEvent)
	return p, err
}

// handoverRecord is what a handover's data hash covers.
type handoverRecord struct {
	ID         string `json:"id"`
	PassportID string `json:"passport_id"`
	FromOrg    string `json:"from_org"`
	ToOrg      string `json:"to_org"`
	Quantity   int64  `json:"quantity"`
	Note       string `json:"note"`
	// Stage is left out when empty, so handovers made before stages keep their hash.
	Stage string `json:"stage,omitempty"`
}

// message builds the typed message both organisations sign for a handover.
func message(id, passportID, fromOrg, toOrg string, quantity int64, dataHash string, nonce int64, deadline time.Time) (chain.Handover, error) {
	hash, err := chain.FromHex(dataHash)
	if err != nil {
		return chain.Handover{}, err
	}
	return chain.Handover{
		PassportID: chain.MustID(passportID),
		FromOrg:    chain.MustID(fromOrg),
		ToOrg:      chain.MustID(toOrg),
		Quantity:   big.NewInt(quantity),
		DataHash:   hash,
		Nonce:      big.NewInt(nonce),
		Deadline:   big.NewInt(deadline.Unix()),
	}, nil
}

// signAs signs a handover digest with an organisation's key and checks the
// signature really is that organisation's before it is trusted.
func (s Store) signAs(ctx context.Context, tx pgx.Tx, orgID string, digest [32]byte) (string, error) {
	var sealed, address *string
	err := tx.QueryRow(ctx,
		`SELECT signing_key, signing_address FROM organisations WHERE id = $1::uuid AND NOT is_deleted`, orgID,
	).Scan(&sealed, &address)
	if err != nil {
		return "", err
	}
	if sealed == nil || address == nil {
		return "", fmt.Errorf("organisation %s has no signing key", orgID)
	}
	hexKey, err := s.Secrets.Open(*sealed)
	if err != nil {
		return "", err
	}
	key, err := chain.ParseKey(hexKey)
	if err != nil {
		return "", err
	}
	sig, err := chain.Sign(digest, key)
	if err != nil {
		return "", err
	}
	if signer, err := chain.Recover(digest, sig); err != nil || signer != common.HexToAddress(*address) {
		return "", fmt.Errorf("organisation %s: stored key does not match its signing address", orgID)
	}
	return "0x" + hex.EncodeToString(sig), nil
}

// StartHandover is the sender's half: it checks the rules, signs the
// handover with the holder's key and leaves it awaiting the receiver.
// The holder does not change yet.
func (s Store) StartHandover(ctx context.Context, passportID, fromOrg, userID string, in HandoverInput, now time.Time) (Handover, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Handover{}, err
	}
	defer tx.Rollback(ctx)

	// Lock the passport: only one handover can be started at a time.
	var holder, status, commodity string
	var remaining, count int64
	err = tx.QueryRow(ctx,
		`SELECT organisation_id::text, status, commodity, remaining, handover_count FROM passports
		 WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, passportID).Scan(&holder, &status, &commodity, &remaining, &count)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && holder != fromOrg) {
		return Handover{}, ErrNotHolder
	}
	if err != nil {
		return Handover{}, err
	}
	if status != StatusActive || remaining <= 0 {
		return Handover{}, ErrNotActive
	}
	if err := noOpenHandover(ctx, tx, passportID); err != nil {
		return Handover{}, err
	}

	// The stage the lot enters must take handovers, and its evidence must be on the passport.
	_, stages, err := liveStages(ctx, tx, commodity)
	if err != nil {
		return Handover{}, err
	}
	i := slices.IndexFunc(stages, func(st template.Stage) bool {
		return st.Key == in.Stage && slices.Contains(st.Events, EventHandover)
	})
	if i < 0 {
		return Handover{}, ErrHandoverStage
	}
	if err := requireEvidence(ctx, tx, stages[i], attestation.SubjectPassport, passportID); err != nil {
		return Handover{}, err
	}

	var receiverOK bool
	err = tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM organisations WHERE id = $1::uuid AND status = $2
		                AND signing_address IS NOT NULL AND NOT is_deleted)`,
		in.ToOrganisationID, auth.OrgVerified).Scan(&receiverOK)
	if err != nil {
		return Handover{}, err
	}
	if !receiverOK || strings.EqualFold(in.ToOrganisationID, fromOrg) {
		return Handover{}, ErrReceiverInvalid
	}

	id, err := newID(ctx, tx)
	if err != nil {
		return Handover{}, err
	}
	hash, err := chain.DataHash(handoverRecord{id, passportID, fromOrg, in.ToOrganisationID, remaining, in.Note, in.Stage})
	if err != nil {
		return Handover{}, err
	}
	dataHash, deadline := chain.Hex(hash), now.Add(HandoverTTL).Truncate(time.Second)

	msg, err := message(id, passportID, fromOrg, in.ToOrganisationID, remaining, dataHash, count, deadline)
	if err != nil {
		return Handover{}, err
	}
	sigFrom, err := s.signAs(ctx, tx, fromOrg, s.Domain.Digest(msg))
	if err != nil {
		return Handover{}, err
	}

	h, err := oneHandover(tx.Query(ctx,
		`INSERT INTO handovers AS h (id, passport_id, from_org, to_org, quantity, note, data_hash, nonce,
		                             deadline, sig_from, status, created_by, stage)
		 VALUES ($3::uuid, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9, $10, $11, $12, $1, $13::uuid, $14)
		 RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, id, passportID, fromOrg, in.ToOrganisationID, remaining, in.Note,
		dataHash, count, deadline, sigFrom, userID, in.Stage))
	if err != nil {
		return Handover{}, err
	}
	if err := notify.Add(ctx, tx, in.ToOrganisationID, notify.KindHandoverStarted, id, MsgNotifyHandoverStarted); err != nil {
		return Handover{}, err
	}
	return h, tx.Commit(ctx)
}

// lockAwaiting locks a handover the acting organisation may answer and
// checks it is still awaiting. side is the column that must equal orgID:
// the receiver answers, the sender cancels. Whether the deadline has passed is
// judged by the database clock, the same clock that reports a handover as expired.
func lockAwaiting(ctx context.Context, tx pgx.Tx, id, side, orgID string) (h Handover, nonce int64, sigFrom string, err error) {
	var (
		status string
		lapsed bool
	)
	err = tx.QueryRow(ctx,
		// side is "to_org" or "from_org", chosen in code.
		`SELECT passport_id::text, from_org::text, to_org::text, quantity, data_hash, nonce, deadline, sig_from,
		        status, deadline < now()
		 FROM handovers WHERE id = $1::uuid AND `+side+` = $2::uuid AND NOT is_deleted FOR UPDATE`, id, orgID,
	).Scan(&h.PassportID, &h.FromOrg, &h.ToOrg, &h.Quantity, &h.DataHash, &nonce, &h.Deadline, &sigFrom, &status, &lapsed)
	if errors.Is(err, pgx.ErrNoRows) {
		return h, 0, "", ErrNotFound
	}
	if err != nil {
		return h, 0, "", err
	}
	if status != HandoverAwaiting || lapsed {
		return h, 0, "", ErrHandoverClosed
	}
	h.ID = id
	return h, nonce, sigFrom, nil
}

const (
	sideReceiver = "to_org"
	sideSender   = "from_org"
)

// Accept is the receiver's half. It signs the same message, and only now,
// with both signatures, does the passport change holder. The change and the
// queued chain transaction are saved together.
func (s Store) Accept(ctx context.Context, id, toOrg, userID string) (Handover, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Handover{}, err
	}
	defer tx.Rollback(ctx)

	h, nonce, sigFrom, err := lockAwaiting(ctx, tx, id, sideReceiver, toOrg)
	if err != nil {
		return Handover{}, err
	}

	// The passport must be exactly as it was when the sender signed.
	var holder, status string
	var remaining, count int64
	err = tx.QueryRow(ctx,
		`SELECT organisation_id::text, status, remaining, handover_count FROM passports
		 WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, h.PassportID).Scan(&holder, &status, &remaining, &count)
	if err != nil {
		return Handover{}, err
	}
	if holder != h.FromOrg || status != StatusActive || remaining != h.Quantity || count != nonce {
		return Handover{}, ErrHandoverStale
	}

	msg, err := message(id, h.PassportID, h.FromOrg, h.ToOrg, h.Quantity, h.DataHash, nonce, h.Deadline)
	if err != nil {
		return Handover{}, err
	}
	sigTo, err := s.signAs(ctx, tx, toOrg, s.Domain.Digest(msg))
	if err != nil {
		return Handover{}, err
	}

	out, err := oneHandover(tx.Query(ctx,
		`UPDATE handovers AS h SET status = $3, sig_to = $4, responded_by = $5::uuid, responded_at = now()
		 WHERE h.id = $6::uuid RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, HandoverCompleted, sigTo, userID, id))
	if err != nil {
		return Handover{}, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE passports SET organisation_id = $2::uuid, handover_count = handover_count + 1 WHERE id = $1::uuid`,
		h.PassportID, toOrg); err != nil {
		return Handover{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, from_org, to_org, quantity, data_hash)
		 VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)`,
		h.PassportID, id, EventHandover, h.FromOrg, h.ToOrg, h.Quantity, h.DataHash); err != nil {
		return Handover{}, err
	}
	err = chain.Enqueue(ctx, tx, chain.KindHandover, id, chain.HandoverSigned{
		HandoverID: id, PassportID: h.PassportID, FromOrg: h.FromOrg, ToOrg: h.ToOrg,
		Quantity: h.Quantity, DataHash: h.DataHash, Nonce: nonce, Deadline: h.Deadline.Unix(),
		SigFrom: sigFrom, SigTo: sigTo,
	})
	if err != nil {
		return Handover{}, err
	}
	if err := notify.Add(ctx, tx, h.FromOrg, notify.KindHandoverAccepted, id, MsgNotifyHandoverAccepted); err != nil {
		return Handover{}, err
	}
	return out, tx.Commit(ctx)
}

// close ends an awaiting handover without moving the passport.
func (s Store) close(ctx context.Context, id, side, orgID, userID, status, reason string) (Handover, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Handover{}, err
	}
	defer tx.Rollback(ctx)

	if _, _, _, err := lockAwaiting(ctx, tx, id, side, orgID); err != nil {
		return Handover{}, err
	}
	h, err := oneHandover(tx.Query(ctx,
		`UPDATE handovers AS h SET status = $3, reject_reason = NULLIF($4, ''), responded_by = $5::uuid, responded_at = now()
		 WHERE h.id = $6::uuid RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, status, reason, userID, id))
	if err != nil {
		return Handover{}, err
	}
	if status == HandoverRejected {
		if err := notify.Add(ctx, tx, h.FromOrg, notify.KindHandoverRejected, id, MsgNotifyHandoverRejected); err != nil {
			return Handover{}, err
		}
	}
	return h, tx.Commit(ctx)
}

// Reject is the receiver refusing the goods. The holder does not change.
func (s Store) Reject(ctx context.Context, id, toOrg, userID, reason string) (Handover, error) {
	return s.close(ctx, id, sideReceiver, toOrg, userID, HandoverRejected, reason)
}

// Cancel is the sender withdrawing a handover before it is answered.
func (s Store) Cancel(ctx context.Context, id, fromOrg, userID string) (Handover, error) {
	return s.close(ctx, id, sideSender, fromOrg, userID, HandoverCancelled, "")
}

// FindHandovers returns handovers the gate's organisation sent or is to receive.
func (s Store) FindHandovers(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Handover], error) {
	// The status filter is applied to the status as read, so "expired" works.
	const shown = `CASE WHEN h.status = $1 AND h.deadline < now() THEN $2 ELSE h.status END`
	return listq.Fetch(ctx, s.Pool, q, HandoverList, "SELECT "+handoverColumns(1, 2), "FROM handovers h",
		[]string{party(3), `($4 = '' OR ` + shown + ` = $4)`},
		[]any{HandoverAwaiting, HandoverExpired, gate, status}, scanHandover)
}

func (s Store) GetHandover(ctx context.Context, id, gate string) (Handover, error) {
	return oneHandover(s.Pool.Query(ctx,
		`SELECT `+handoverColumns(1, 2)+` FROM handovers h
		 WHERE h.id = $3::uuid AND NOT h.is_deleted AND `+party(4),
		HandoverAwaiting, HandoverExpired, id, gate))
}
