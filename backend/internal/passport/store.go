package passport

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
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
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrFacilityNotFound  = errors.New("facility not found")
	ErrFacilitySuspended = errors.New("facility suspended")
	ErrFacilityPending   = errors.New("facility pending approval")
	ErrLicenceExpired    = errors.New("facility licence expired")
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
	ID           string    `json:"id"`
	PassportID   string    `json:"passport_id"`
	FromOrg      string    `json:"from_org"`
	ToOrg        string    `json:"to_org"`
	Stage        string    `json:"stage"` // the template stage the lot enters
	Quantity     int64     `json:"quantity"`
	Note         string    `json:"note"`
	DataHash     string    `json:"data_hash"`
	Deadline     time.Time `json:"deadline"`
	Status       string    `json:"status"`
	SignedByFrom bool      `json:"signed_by_sender"`
	SignedByTo   bool      `json:"signed_by_receiver"`
	RejectReason *string   `json:"reject_reason"` // the receiver's note
	// RejectReasons are why the receiver refused, from RejectReasons.
	RejectReasons []string   `json:"reject_reasons"`
	RespondedAt   *time.Time `json:"responded_at"`
	TxHash        *string    `json:"tx_hash"`
	AnchoredAt    *time.Time `json:"anchored_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// SplitFrom is set when part of a lot was handed over: the passport it was
	// cut from. PassportID (and ChildPassportID) is then the new passport.
	SplitFrom       *string `json:"split_from_passport_id"`
	ChildPassportID *string `json:"child_passport_id"`
	// Location is where the sender handed the goods over.
	Location *Location `json:"location"`
	// ReceivedQuantity is what the receiver weighed on arrival, and where.
	ReceivedQuantity *int64    `json:"received_quantity"`
	ReceiverLocation *Location `json:"receiver_location"`
	Photos           []Photo   `json:"photos"`
	// TemplateVersion is the version of the commodity's template it was judged by.
	TemplateVersion *int `json:"template_version"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func location(lat, lon *float64) *Location {
	if lat == nil || lon == nil {
		return nil
	}
	return &Location{*lat, *lon}
}

// Photo is a picture taken at a handover by its sender or its receiver.
type Photo struct {
	ID          string    `json:"id"`
	Side        string    `json:"side"` // PhotoBySender or PhotoByReceiver
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	SHA256      string    `json:"sha256"`
	StorageKey  string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
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
	           true, h.sig_to IS NOT NULL, h.reject_reason, h.reject_reasons, h.responded_at, h.tx_hash, h.anchored_at,
	           h.created_at, h.updated_at, h.split_from::text, h.lat, h.lon, h.received_quantity,
	           h.receiver_lat, h.receiver_lon, h.template_version`

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
	var (
		h                                  Handover
		lat, lon, receiverLat, receiverLon *float64
	)
	err := row.Scan(&h.ID, &h.PassportID, &h.FromOrg, &h.ToOrg, &h.Stage, &h.Quantity, &h.Note, &h.DataHash,
		&h.Deadline, &h.Status, &h.SignedByFrom, &h.SignedByTo, &h.RejectReason, &h.RejectReasons, &h.RespondedAt,
		&h.TxHash, &h.AnchoredAt, &h.CreatedAt, &h.UpdatedAt, &h.SplitFrom, &lat, &lon, &h.ReceivedQuantity,
		&receiverLat, &receiverLon, &h.TemplateVersion)
	if h.SplitFrom != nil {
		h.ChildPassportID = &h.PassportID
	}
	h.Location, h.ReceiverLocation = location(lat, lon), location(receiverLat, receiverLon)
	h.Photos = []Photo{}
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
	var licenceValidTo *time.Time
	err = tx.QueryRow(ctx,
		`SELECT kind, status, capacity, capacity_unit, capacity_period, licence_valid_to FROM facilities
		 WHERE id = $1::uuid AND organisation_id = $2::uuid AND NOT is_deleted FOR UPDATE`,
		in.FacilityID, orgID).Scan(&fKind, &fStatus, &capacity, &fUnit, &fPeriod, &licenceValidTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return Passport{}, ErrFacilityNotFound
	}
	if err != nil {
		return Passport{}, err
	}
	switch {
	case fStatus == facility.StatusPending:
		return Passport{}, ErrFacilityPending
	case fStatus != facility.StatusActive:
		return Passport{}, ErrFacilitySuspended
	// The licence is valid through its last day.
	case licenceValidTo != nil && now.UTC().Format(facility.DateLayout) > licenceValidTo.Format(facility.DateLayout):
		return Passport{}, ErrLicenceExpired
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
	// Where the goods were handed over, and the SHA-256 of each sender photo:
	// left out when absent, so older handovers keep their hash.
	Lat    *float64 `json:"lat,omitempty"`
	Lon    *float64 `json:"lon,omitempty"`
	Photos []string `json:"photos,omitempty"`
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
//
// Handing over less than the passport holds first cuts that part into its own
// passport (a split at yield 1, in this transaction), and that passport is
// handed over; the rest stays on the original. If the handover is then
// rejected, cancelled or lapses, the new passport simply stays with the sender.
func (s Store) StartHandover(ctx context.Context, passportID, fromOrg, userID string, in HandoverInput,
	photos []upload.File, now time.Time) (Handover, error) {

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Handover{}, err
	}
	defer tx.Rollback(ctx)

	// Lock the passport: only one handover can be started at a time.
	var holder, status, commodity, unit string
	var remaining, count int64
	err = tx.QueryRow(ctx,
		`SELECT organisation_id::text, status, commodity, unit, remaining, handover_count FROM passports
		 WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, passportID).Scan(&holder, &status, &commodity, &unit, &remaining, &count)
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
	quantity := remaining
	if in.Quantity != nil {
		if *in.Quantity > remaining {
			return Handover{}, InsufficientError{PassportID: passportID, Requested: *in.Quantity, Remaining: remaining, Unit: unit}
		}
		quantity = *in.Quantity
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

	// Part of the lot: cut it into its own passport, which is what moves.
	handed, nonce := passportID, count
	var splitFrom *string
	if quantity < remaining {
		t, err := transformTx(ctx, tx, fromOrg, userID, TransformInput{
			Stage:   template.StageSplit,
			Inputs:  []TransformSource{{PassportID: passportID, Quantity: quantity}},
			Outputs: []TransformOutput{{Quantity: quantity}},
		}, now, true)
		if err != nil {
			return Handover{}, err
		}
		handed, nonce, splitFrom = t.Outputs[0].ID, 0, &passportID
	}

	id, err := newID(ctx, tx)
	if err != nil {
		return Handover{}, err
	}
	shas := make([]string, len(photos))
	for i, f := range photos {
		shas[i] = f.SHA256
	}
	hash, err := chain.DataHash(handoverRecord{ID: id, PassportID: handed, FromOrg: fromOrg, ToOrg: in.ToOrganisationID,
		Quantity: quantity, Note: in.Note, Stage: in.Stage, Lat: in.Lat, Lon: in.Lon, Photos: shas})
	if err != nil {
		return Handover{}, err
	}
	dataHash, deadline := chain.Hex(hash), now.Add(HandoverTTL).Truncate(time.Second)

	msg, err := message(id, handed, fromOrg, in.ToOrganisationID, quantity, dataHash, nonce, deadline)
	if err != nil {
		return Handover{}, err
	}
	sigFrom, err := s.signAs(ctx, tx, fromOrg, s.Domain.Digest(msg))
	if err != nil {
		return Handover{}, err
	}

	h, err := oneHandover(tx.Query(ctx,
		`INSERT INTO handovers AS h (id, passport_id, from_org, to_org, quantity, note, data_hash, nonce,
		                             deadline, sig_from, status, created_by, stage, split_from, lat, lon)
		 VALUES ($3::uuid, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9, $10, $11, $12, $1, $13::uuid, $14, $15::uuid, $16, $17)
		 RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, id, handed, fromOrg, in.ToOrganisationID, quantity, in.Note,
		dataHash, nonce, deadline, sigFrom, userID, in.Stage, splitFrom, in.Lat, in.Lon))
	if err != nil {
		return Handover{}, err
	}
	if h.Photos, err = addPhotos(ctx, tx, id, PhotoBySender, userID, photos); err != nil {
		return Handover{}, err
	}
	if err := notify.Add(ctx, tx, in.ToOrganisationID, notify.KindHandoverStarted, id, MsgNotifyHandoverStarted); err != nil {
		return Handover{}, err
	}
	return h, tx.Commit(ctx)
}

// addPhotos records photos already stored in the bucket against a handover.
func addPhotos(ctx context.Context, tx pgx.Tx, handoverID, side, userID string, photos []upload.File) ([]Photo, error) {
	out := []Photo{}
	for _, f := range photos {
		p := Photo{Side: side, FileName: f.Name, ContentType: f.ContentType, SizeBytes: f.Size, SHA256: f.SHA256, StorageKey: f.Key}
		err := tx.QueryRow(ctx,
			`INSERT INTO handover_photos (handover_id, side, file_name, content_type, size_bytes, sha256, storage_key, uploaded_by)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::uuid) RETURNING id::text, created_at`,
			handoverID, side, f.Name, f.ContentType, f.Size, f.SHA256, f.Key, userID).Scan(&p.ID, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// querier is a pool or a transaction.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// attachPhotos fills in the photos of each handover.
func attachPhotos(ctx context.Context, q querier, hs []Handover) error {
	if len(hs) == 0 {
		return nil
	}
	ids := make([]string, len(hs))
	for i, h := range hs {
		ids[i] = h.ID
	}
	by := map[string][]Photo{}
	rows, err := q.Query(ctx,
		`SELECT handover_id::text, id::text, side, file_name, content_type, size_bytes, sha256, storage_key, created_at
		 FROM handover_photos WHERE handover_id = ANY($1::uuid[]) AND NOT is_deleted ORDER BY created_at, id`, ids)
	if err != nil {
		return err
	}
	for rows.Next() {
		var hid string
		var p Photo
		if err := rows.Scan(&hid, &p.ID, &p.Side, &p.FileName, &p.ContentType, &p.SizeBytes, &p.SHA256, &p.StorageKey, &p.CreatedAt); err != nil {
			rows.Close()
			return err
		}
		by[hid] = append(by[hid], p)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range hs {
		if p := by[hs[i].ID]; p != nil {
			hs[i].Photos = p
		}
	}
	return nil
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

// ToleranceError says the weight a receiver recorded differs from what was
// sent by more than the stage allows.
type ToleranceError struct {
	Sent, Received int64
	Unit           string
	Tolerance      float64 // percent
}

// Difference is how far, in percent of what was sent, the two weights differ.
func (e ToleranceError) Difference() float64 {
	return math.Abs(float64(e.Sent-e.Received)) * 100 / float64(e.Sent)
}

func (e ToleranceError) Error() string {
	return fmt.Sprintf("received %d %s but %d %s was sent: %.1f%% off, and this stage accepts at most %.1f%%. Reject the handover, or raise a flag",
		e.Received, e.Unit, e.Sent, e.Unit, e.Difference(), e.Tolerance)
}

// tolerance is the weight difference, in percent, the handover's stage accepts.
func tolerance(ctx context.Context, tx pgx.Tx, handoverID string) (pct float64, unit string, err error) {
	var stage, commodity string
	err = tx.QueryRow(ctx,
		`SELECT h.stage, p.commodity, p.unit FROM handovers h JOIN passports p ON p.id = h.passport_id
		 WHERE h.id = $1::uuid`, handoverID).Scan(&stage, &commodity, &unit)
	if err != nil {
		return 0, "", err
	}
	_, stages, err := liveStages(ctx, tx, commodity)
	if err != nil {
		return 0, "", err
	}
	for _, st := range stages {
		if st.Key == stage {
			return st.Tolerance(), unit, nil
		}
	}
	return template.DefaultHandoverTolerance, unit, nil // a handover from before stages
}

// Accept is the receiver's half. They record what they weighed; within the
// stage's tolerance they sign the same message, and only now, with both
// signatures, does the passport change holder. The change and the queued
// chain transaction are saved together. Their own photos and location are
// kept with the handover, outside the signed record (the sender signed first).
//
// A difference above RuleFlagShare of the tolerance is accepted, but the
// rules raise a flag about it in the same transaction.
func (s Store) Accept(ctx context.Context, id, toOrg, userID string, in AcceptInput, photos []upload.File,
	raise RaiseFlag) (Handover, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Handover{}, err
	}
	defer tx.Rollback(ctx)

	h, nonce, sigFrom, err := lockAwaiting(ctx, tx, id, sideReceiver, toOrg)
	if err != nil {
		return Handover{}, err
	}
	pct, unit, err := tolerance(ctx, tx, id)
	if err != nil {
		return Handover{}, err
	}
	off := ToleranceError{Sent: h.Quantity, Received: *in.ReceivedQuantity, Unit: unit, Tolerance: pct}
	if off.Difference() > pct {
		return Handover{}, off
	}
	if raise != nil && off.Difference() > pct*RuleFlagShare {
		reason := fmt.Sprintf(MsgRuleToleranceFmt, off.Difference(), pct)
		if err := raise(ctx, tx, toOrg, userID, toOrg, h.PassportID, reason, SourceRules); err != nil {
			return Handover{}, err
		}
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
		`UPDATE handovers AS h SET status = $3, sig_to = $4, responded_by = $5::uuid, responded_at = now(),
		        received_quantity = $7, receiver_lat = $8, receiver_lon = $9
		 WHERE h.id = $6::uuid RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, HandoverCompleted, sigTo, userID, id, *in.ReceivedQuantity, in.Lat, in.Lon))
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
	if _, err := addPhotos(ctx, tx, id, PhotoByReceiver, userID, photos); err != nil {
		return Handover{}, err
	}
	if err := notify.Add(ctx, tx, h.FromOrg, notify.KindHandoverAccepted, id, MsgNotifyHandoverAccepted); err != nil {
		return Handover{}, err
	}
	hs := []Handover{out}
	if err := attachPhotos(ctx, tx, hs); err != nil {
		return Handover{}, err
	}
	return hs[0], tx.Commit(ctx)
}

// RaiseFlag raises a flag about a passport in the caller's transaction
// (flag.RaiseFlag, passed in: the flag package builds on this one). source
// says who raised it: a person, or the rules (see flag.Sources).
type RaiseFlag func(ctx context.Context, tx pgx.Tx, orgID, userID, holder, passportID, reason, source string) error

// close ends an awaiting handover without moving the passport. after, if
// set, runs in the same transaction with the closed handover.
func (s Store) close(ctx context.Context, id, side, orgID, userID, status, note string, reasons []string,
	after func(pgx.Tx, Handover) error) (Handover, error) {

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Handover{}, err
	}
	defer tx.Rollback(ctx)

	if _, _, _, err := lockAwaiting(ctx, tx, id, side, orgID); err != nil {
		return Handover{}, err
	}
	if reasons == nil {
		reasons = []string{}
	}
	h, err := oneHandover(tx.Query(ctx,
		`UPDATE handovers AS h SET status = $3, reject_reason = NULLIF($4, ''), reject_reasons = $7,
		        responded_by = $5::uuid, responded_at = now()
		 WHERE h.id = $6::uuid RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, status, note, userID, id, reasons))
	if err != nil {
		return Handover{}, err
	}
	if status == HandoverRejected {
		if err := notify.Add(ctx, tx, h.FromOrg, notify.KindHandoverRejected, id, MsgNotifyHandoverRejected); err != nil {
			return Handover{}, err
		}
	}
	if after != nil {
		if err := after(tx, h); err != nil {
			return Handover{}, err
		}
	}
	hs := []Handover{h}
	if err := attachPhotos(ctx, tx, hs); err != nil {
		return Handover{}, err
	}
	return hs[0], tx.Commit(ctx)
}

// Reject is the receiver refusing the goods, saying why. The holder does not
// change. With AlsoFlag, a flag about the passport is raised in the same
// transaction, so the regulator looks into it.
func (s Store) Reject(ctx context.Context, id, toOrg, userID string, in RejectInput, raise RaiseFlag) (Handover, error) {
	var after func(pgx.Tx, Handover) error
	if in.AlsoFlag {
		after = func(tx pgx.Tx, h Handover) error {
			reason := MsgFlagFromReject + ": " + strings.Join(in.Reasons, ", ")
			if in.Note != "" {
				reason += ". " + in.Note
			}
			return raise(ctx, tx, toOrg, userID, h.FromOrg, h.PassportID, reason, SourceUser)
		}
	}
	return s.close(ctx, id, sideReceiver, toOrg, userID, HandoverRejected, in.Note, in.Reasons, after)
}

// Cancel is the sender withdrawing a handover before it is answered.
func (s Store) Cancel(ctx context.Context, id, fromOrg, userID string) (Handover, error) {
	return s.close(ctx, id, sideSender, fromOrg, userID, HandoverCancelled, "", nil, nil)
}

// FindHandovers returns handovers the gate's organisation sent or is to receive.
func (s Store) FindHandovers(ctx context.Context, q listq.Query, gate, status string) (listq.Page[Handover], error) {
	// The status filter is applied to the status as read, so "expired" works.
	const shown = `CASE WHEN h.status = $1 AND h.deadline < now() THEN $2 ELSE h.status END`
	page, err := listq.Fetch(ctx, s.Pool, q, HandoverList, "SELECT "+handoverColumns(1, 2), "FROM handovers h",
		[]string{party(3), `($4 = '' OR ` + shown + ` = $4)`},
		[]any{HandoverAwaiting, HandoverExpired, gate, status}, scanHandover)
	if err == nil {
		err = attachPhotos(ctx, s.Pool, page.Data)
	}
	return page, err
}

func (s Store) GetHandover(ctx context.Context, id, gate string) (Handover, error) {
	h, err := oneHandover(s.Pool.Query(ctx,
		`SELECT `+handoverColumns(1, 2)+` FROM handovers h
		 WHERE h.id = $3::uuid AND NOT h.is_deleted AND `+party(4),
		HandoverAwaiting, HandoverExpired, id, gate))
	if err != nil {
		return Handover{}, err
	}
	hs := []Handover{h}
	if err := attachPhotos(ctx, s.Pool, hs); err != nil {
		return Handover{}, err
	}
	return hs[0], nil
}

// HandoverPhoto returns one photo of a handover the gate's organisation is a party to.
func (s Store) HandoverPhoto(ctx context.Context, handoverID, photoID, gate string) (Photo, error) {
	var p Photo
	err := s.Pool.QueryRow(ctx,
		`SELECT ph.id::text, ph.side, ph.file_name, ph.content_type, ph.size_bytes, ph.sha256, ph.storage_key, ph.created_at
		 FROM handover_photos ph JOIN handovers h ON h.id = ph.handover_id
		 WHERE ph.id = $1::uuid AND ph.handover_id = $2::uuid AND NOT ph.is_deleted AND NOT h.is_deleted AND `+party(3),
		photoID, handoverID, gate).Scan(&p.ID, &p.Side, &p.FileName, &p.ContentType, &p.SizeBytes, &p.SHA256, &p.StorageKey, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Photo{}, ErrNotFound
	}
	return p, err
}
