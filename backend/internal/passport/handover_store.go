package passport

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

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
	Location   *Location       `json:"location"`
	SealNumber *string         `json:"seal_number,omitempty"`
	Checklist  map[string]bool `json:"checklist,omitempty"`
	// ReceivedQuantity is what the receiver weighed on arrival, and where.
	ReceivedQuantity *int64    `json:"received_quantity"`
	ReceiverLocation *Location `json:"receiver_location"`
	Photos           []Photo   `json:"photos"`
	// TemplateVersion is the version of the commodity's template it was judged by.
	TemplateVersion *int `json:"template_version"`
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

func scanHandover(row pgx.CollectableRow) (Handover, error) {
	var (
		h                                  Handover
		lat, lon, receiverLat, receiverLon *float64
		checklistJSON                      []byte
	)
	err := row.Scan(&h.ID, &h.PassportID, &h.FromOrg, &h.ToOrg, &h.Stage, &h.Quantity, &h.Note, &h.DataHash,
		&h.Deadline, &h.Status, &h.SignedByFrom, &h.SignedByTo, &h.RejectReason, &h.RejectReasons, &h.RespondedAt,
		&h.TxHash, &h.AnchoredAt, &h.CreatedAt, &h.UpdatedAt, &h.SplitFrom, &lat, &lon, &h.ReceivedQuantity,
		&receiverLat, &receiverLon, &h.TemplateVersion, &h.SealNumber, &checklistJSON)
	if len(checklistJSON) > 0 {
		_ = json.Unmarshal(checklistJSON, &h.Checklist)
	}
	if h.SplitFrom != nil {
		h.ChildPassportID = &h.PassportID
	}
	h.Location, h.ReceiverLocation = location(lat, lon), location(receiverLat, receiverLon)
	h.Photos = []Photo{}
	return h, err
}

func oneHandover(rows pgx.Rows, err error) (Handover, error) { return one(rows, err, scanHandover) }

func party(n int) string { return fmt.Sprintf(partyFmt, n) }

func handoverColumns(awaitingArg, expiredArg int) string {
	return fmt.Sprintf(handoverColumnsFmt, awaitingArg, expiredArg)
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
func message(passportID, fromOrg, toOrg string, quantity int64, dataHash string, nonce int64, deadline time.Time) (chain.Handover, error) {
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

	lot, err := lockForHandover(ctx, tx, passportID, fromOrg, in.Quantity)
	if err != nil {
		return Handover{}, err
	}
	if err := checkHandoverStage(ctx, tx, lot.commodity, in.Stage, passportID); err != nil {
		return Handover{}, err
	}
	if err := checkReceiver(ctx, tx, in.ToOrganisationID, fromOrg); err != nil {
		return Handover{}, err
	}
	handed, nonce, splitFrom, err := splitIfPart(ctx, tx, lot, passportID, fromOrg, userID, now)
	if err != nil {
		return Handover{}, err
	}
	sh := sentHandover{handed: handed, fromOrg: fromOrg, userID: userID, quantity: lot.quantity,
		nonce: nonce, splitFrom: splitFrom, deadline: now.Add(HandoverTTL).Truncate(time.Second)}
	h, err := s.insertHandover(ctx, tx, sh, in, photos)
	if err != nil {
		return Handover{}, err
	}
	if err := notify.Add(ctx, tx, in.ToOrganisationID, notify.KindHandoverStarted, h.ID, MsgNotifyHandoverStarted); err != nil {
		return Handover{}, err
	}
	return h, tx.Commit(ctx)
}

type heldLot struct {
	commodity           string
	remaining, quantity int64
	count               int64 // handovers so far: the next nonce
}

// lockForHandover locks the passport so only one handover starts at a time,
// and works out how much moves: all of it, or the requested part.
func lockForHandover(ctx context.Context, tx pgx.Tx, passportID, fromOrg string, want *int64) (heldLot, error) {
	var l heldLot
	var holder, status, unit string
	err := tx.QueryRow(ctx,
		`SELECT organisation_id::text, status, commodity, unit, remaining, handover_count FROM passports
		 WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, passportID).Scan(&holder, &status, &l.commodity, &unit, &l.remaining, &l.count)
	switch {
	case errors.Is(err, pgx.ErrNoRows) || (err == nil && holder != fromOrg):
		return l, ErrNotHolder
	case err != nil:
		return l, err
	case status != StatusActive || l.remaining <= 0:
		return l, ErrNotActive
	}
	if err := noOpenHandover(ctx, tx, passportID); err != nil {
		return l, err
	}
	l.quantity = l.remaining
	if want != nil {
		if *want > l.remaining {
			return l, InsufficientError{PassportID: passportID, Requested: *want, Remaining: l.remaining, Unit: unit}
		}
		l.quantity = *want
	}
	return l, nil
}

// checkHandoverStage: the stage the lot enters must take handovers, and its
// evidence must be on the passport.
func checkHandoverStage(ctx context.Context, tx pgx.Tx, commodity, stage, passportID string) error {
	_, stages, err := liveStages(ctx, tx, commodity)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(stages, func(st template.Stage) bool {
		return st.Key == stage && slices.Contains(st.Events, EventHandover)
	})
	if i < 0 {
		return ErrHandoverStage
	}
	return requireEvidence(ctx, tx, stages[i], attestation.SubjectPassport, passportID)
}

// checkReceiver needs a verified organisation with a signing key, other than the sender.
func checkReceiver(ctx context.Context, tx pgx.Tx, toOrg, fromOrg string) error {
	var ok bool
	err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM organisations WHERE id = $1::uuid AND status = $2
		                AND signing_address IS NOT NULL AND NOT is_deleted)`,
		toOrg, auth.OrgVerified).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok || strings.EqualFold(toOrg, fromOrg) {
		return ErrReceiverInvalid
	}
	return nil
}

// splitIfPart cuts a part of the lot into its own passport, which is then what
// moves (with a fresh nonce). The whole lot moves as itself.
func splitIfPart(ctx context.Context, tx pgx.Tx, l heldLot, passportID, fromOrg, userID string, now time.Time) (handed string, nonce int64, splitFrom *string, err error) {
	if l.quantity == l.remaining {
		return passportID, l.count, nil, nil
	}
	t, err := transformTx(ctx, tx, fromOrg, userID, TransformInput{
		Stage:   template.StageSplit,
		Inputs:  []TransformSource{{PassportID: passportID, Quantity: l.quantity}},
		Outputs: []TransformOutput{{Quantity: l.quantity}},
	}, now, true)
	if err != nil {
		return "", 0, nil, err
	}
	return t.Outputs[0].ID, 0, &passportID, nil
}

type sentHandover struct {
	handed, fromOrg, userID string
	quantity, nonce         int64
	splitFrom               *string
	deadline                time.Time
}

// insertHandover hashes the handover, signs it as the sender and stores it with its photos.
func (s Store) insertHandover(ctx context.Context, tx pgx.Tx, sh sentHandover, in HandoverInput, photos []upload.File) (Handover, error) {
	id, err := newID(ctx, tx)
	if err != nil {
		return Handover{}, err
	}
	dataHash, sigFrom, err := s.signSent(ctx, tx, id, sh, in, photos)
	if err != nil {
		return Handover{}, err
	}
	checklistJSON, _ := json.Marshal(in.Checklist)
	if len(checklistJSON) == 0 {
		checklistJSON = []byte(`{}`)
	}
	h, err := oneHandover(tx.Query(ctx,
		`INSERT INTO handovers AS h (id, passport_id, from_org, to_org, quantity, note, data_hash, nonce,
		                             deadline, sig_from, status, created_by, stage, split_from, lat, lon, seal_number, checklist)
		 VALUES ($3::uuid, $4::uuid, $5::uuid, $6::uuid, $7, $8, $9, $10, $11, $12, $1, $13::uuid, $14, $15::uuid, $16, $17, NULLIF($18,''), $19::jsonb)
		 RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, id, sh.handed, sh.fromOrg, in.ToOrganisationID, sh.quantity, in.Note,
		dataHash, sh.nonce, sh.deadline, sigFrom, sh.userID, in.Stage, sh.splitFrom, in.Lat, in.Lon, in.SealNumber, string(checklistJSON)))
	if err != nil {
		return Handover{}, err
	}
	h.Photos, err = addPhotos(ctx, tx, id, PhotoBySender, sh.userID, photos)
	return h, err
}

func (s Store) signSent(ctx context.Context, tx pgx.Tx, id string, sh sentHandover, in HandoverInput, photos []upload.File) (dataHash, sig string, err error) {
	shas := make([]string, len(photos))
	for i, f := range photos {
		shas[i] = f.SHA256
	}
	hash, err := chain.DataHash(handoverRecord{ID: id, PassportID: sh.handed, FromOrg: sh.fromOrg, ToOrg: in.ToOrganisationID,
		Quantity: sh.quantity, Note: in.Note, Stage: in.Stage, Lat: in.Lat, Lon: in.Lon, Photos: shas})
	if err != nil {
		return "", "", err
	}
	dataHash = chain.Hex(hash)
	msg, err := message(sh.handed, sh.fromOrg, in.ToOrganisationID, sh.quantity, dataHash, sh.nonce, sh.deadline)
	if err != nil {
		return "", "", err
	}
	sig, err = s.signAs(ctx, tx, sh.fromOrg, s.Domain.Digest(msg))
	return dataHash, sig, err
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
	if err := checkReceived(ctx, tx, h, *in.ReceivedQuantity, toOrg, userID, raise); err != nil {
		return Handover{}, err
	}
	if err := checkUnchanged(ctx, tx, h, nonce); err != nil {
		return Handover{}, err
	}
	msg, err := message(h.PassportID, h.FromOrg, h.ToOrg, h.Quantity, h.DataHash, nonce, h.Deadline)
	if err != nil {
		return Handover{}, err
	}
	sigTo, err := s.signAs(ctx, tx, toOrg, s.Domain.Digest(msg))
	if err != nil {
		return Handover{}, err
	}
	out, err := completeHandover(ctx, tx, h, userID, sigTo, in, photos)
	if err != nil {
		return Handover{}, err
	}
	if err := moveCustody(ctx, tx, h, nonce, sigFrom, sigTo); err != nil {
		return Handover{}, err
	}
	if err := notify.Add(ctx, tx, h.FromOrg, notify.KindHandoverAccepted, id, MsgNotifyHandoverAccepted); err != nil {
		return Handover{}, err
	}
	return out, tx.Commit(ctx)
}

// completeHandover records the receiver's answer and photos, and returns the
// handover with every photo from both sides.
func completeHandover(ctx context.Context, tx pgx.Tx, h Handover, userID, sigTo string, in AcceptInput, photos []upload.File) (Handover, error) {
	out, err := oneHandover(tx.Query(ctx,
		`UPDATE handovers AS h SET status = $3, sig_to = $4, responded_by = $5::uuid, responded_at = now(),
		        received_quantity = $7, receiver_lat = $8, receiver_lon = $9
		 WHERE h.id = $6::uuid RETURNING `+handoverColumns(1, 2),
		HandoverAwaiting, HandoverExpired, HandoverCompleted, sigTo, userID, h.ID, *in.ReceivedQuantity, in.Lat, in.Lon))
	if err != nil {
		return Handover{}, err
	}
	if _, err := addPhotos(ctx, tx, h.ID, PhotoByReceiver, userID, photos); err != nil {
		return Handover{}, err
	}
	hs := []Handover{out}
	err = attachPhotos(ctx, tx, hs)
	return hs[0], err
}

// checkReceived refuses a weight off by more than the stage allows, and flags
// one that comes close (RuleFlagShare of the tolerance).
func checkReceived(ctx context.Context, tx pgx.Tx, h Handover, received int64, toOrg, userID string, raise RaiseFlag) error {
	pct, unit, err := tolerance(ctx, tx, h.ID)
	if err != nil {
		return err
	}
	off := ToleranceError{Sent: h.Quantity, Received: received, Unit: unit, Tolerance: pct}
	if off.Difference() > pct {
		return off
	}
	if raise == nil || off.Difference() <= pct*RuleFlagShare {
		return nil
	}
	reason := fmt.Sprintf(MsgRuleToleranceFmt, off.Difference(), pct)
	return raise(ctx, tx, toOrg, userID, toOrg, h.PassportID, reason, SourceRules)
}

// checkUnchanged: the passport must be exactly as it was when the sender signed.
func checkUnchanged(ctx context.Context, tx pgx.Tx, h Handover, nonce int64) error {
	var holder, status string
	var remaining, count int64
	err := tx.QueryRow(ctx,
		`SELECT organisation_id::text, status, remaining, handover_count FROM passports
		 WHERE id = $1::uuid AND NOT is_deleted FOR UPDATE`, h.PassportID).Scan(&holder, &status, &remaining, &count)
	if err != nil {
		return err
	}
	if holder != h.FromOrg || status != StatusActive || remaining != h.Quantity || count != nonce {
		return ErrHandoverStale
	}
	return nil
}

// moveCustody gives the passport to the receiver, records the trace event and
// queues both signatures for the chain.
func moveCustody(ctx context.Context, tx pgx.Tx, h Handover, nonce int64, sigFrom, sigTo string) error {
	if _, err := tx.Exec(ctx,
		`UPDATE passports SET organisation_id = $2::uuid, handover_count = handover_count + 1 WHERE id = $1::uuid`,
		h.PassportID, h.ToOrg); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO trace_events (passport_id, ref_id, kind, from_org, to_org, quantity, data_hash)
		 VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)`,
		h.PassportID, h.ID, EventHandover, h.FromOrg, h.ToOrg, h.Quantity, h.DataHash); err != nil {
		return err
	}
	return chain.Enqueue(ctx, tx, chain.KindHandover, h.ID, chain.HandoverSigned{
		HandoverID: h.ID, PassportID: h.PassportID, FromOrg: h.FromOrg, ToOrg: h.ToOrg,
		Quantity: h.Quantity, DataHash: h.DataHash, Nonce: nonce, Deadline: h.Deadline.Unix(),
		SigFrom: sigFrom, SigTo: sigTo,
	})
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
func (s Store) FindHandovers(ctx context.Context, q listq.Query, gate, status, box string) (listq.Page[Handover], error) {
	// The status filter is applied to the status as read, so "expired" works.
	const shown = `CASE WHEN h.status = $1 AND h.deadline < now() THEN $2 ELSE h.status END`
	inBox := `($5 = ''
		OR ($5 = '` + BoxIncoming + `' AND h.to_org = NULLIF($3, '')::uuid AND ` + shown + ` = $1)
		OR ($5 = '` + BoxOutgoing + `' AND h.from_org = NULLIF($3, '')::uuid AND ` + shown + ` = $1)
		OR ($5 = '` + BoxDone + `' AND ` + shown + ` <> $1))`
	page, err := listq.Fetch(ctx, s.Pool, q, HandoverList, "SELECT "+handoverColumns(1, 2), "FROM handovers h",
		[]string{party(3), `($4 = '' OR ` + shown + ` = $4)`, inBox},
		[]any{HandoverAwaiting, HandoverExpired, gate, status, box}, scanHandover)
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
