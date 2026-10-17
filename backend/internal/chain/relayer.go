package chain

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/chain/bindings"
)

// Backend is the chain connection the relayer needs: sending transactions
// and reading their receipts. An RPC client and the in-process test chain
// both satisfy it.
type Backend interface {
	bind.ContractBackend
	bind.DeployBackend
}

// Contracts are the deployed contracts the relayer writes to.
type Contracts struct {
	Orgs         *bindings.OrgRegistry
	Facilities   *bindings.FacilityRegistry
	Passports    *bindings.PassportRegistry
	Custody      *bindings.CustodyManager
	Attestations *bindings.AttestationRegistry
}

// Addresses are where the contracts live on the target chain.
type Addresses struct {
	Orgs, Facilities, Passports, Custody, Attestations common.Address
}

// Bind attaches to the deployed contracts.
func Bind(a Addresses, backend bind.ContractBackend) (Contracts, error) {
	var (
		c   Contracts
		err error
	)
	if c.Orgs, err = bindings.NewOrgRegistry(a.Orgs, backend); err != nil {
		return c, err
	}
	if c.Facilities, err = bindings.NewFacilityRegistry(a.Facilities, backend); err != nil {
		return c, err
	}
	if c.Passports, err = bindings.NewPassportRegistry(a.Passports, backend); err != nil {
		return c, err
	}
	if c.Custody, err = bindings.NewCustodyManager(a.Custody, backend); err != nil {
		return c, err
	}
	c.Attestations, err = bindings.NewAttestationRegistry(a.Attestations, backend)
	return c, err
}

// Relayer carries queued changes to the chain, oldest first, one at a time.
//
// ponytail: one worker, strictly sequential. That keeps ordering trivially
// correct (an organisation is registered before its facility). If throughput
// ever matters, batch independent items, but keep per-subject order.
type Relayer struct {
	Pool      *pgxpool.Pool
	Backend   Backend
	Signer    *bind.TransactOpts // the relayer key (relayer + operator): pays for and sends every transaction
	Contracts Contracts
	// NotBefore, when set, makes the relayer ignore items queued earlier.
	// Tests use it so a test chain never processes someone else's items.
	NotBefore time.Time
}

type item struct {
	ID, Kind, SubjectID string
	Payload             []byte
	TxHash              *string
}

// errRefused marks a transaction the contract rejected. Retrying cannot help.
type errRefused struct{ reason string }

func (e errRefused) Error() string { return "refused by contract: " + e.reason }

// Run processes the queue until ctx ends. A chain that cannot be reached is
// not an error to give up on: items simply wait.
func (r *Relayer) Run(ctx context.Context) {
	for ctx.Err() == nil {
		worked, err := r.ProcessNext(ctx)
		if err != nil {
			log.Printf("relayer: %v (will retry)", err)
		}
		if err != nil || !worked {
			select {
			case <-ctx.Done():
			case <-time.After(PollInterval):
			}
		}
	}
}

// ProcessNext handles the oldest pending item. It reports whether there was
// one. A returned error means the chain could not be reached or the result
// could not be recorded; the item stays pending and is tried again.
func (r *Relayer) ProcessNext(ctx context.Context) (worked bool, err error) {
	var it item
	err = r.Pool.QueryRow(ctx,
		`SELECT id::text, kind, subject_id::text, payload::text, tx_hash FROM chain_outbox
		 WHERE status = $1 AND created_at >= $2 AND NOT is_deleted ORDER BY seq LIMIT 1`, StatusPending, r.NotBefore,
	).Scan(&it.ID, &it.Kind, &it.SubjectID, &it.Payload, &it.TxHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	hash, err := r.submit(ctx, it)
	var refused errRefused
	switch {
	case errors.As(err, &refused):
		log.Printf("relayer: %s %s %v", it.Kind, it.SubjectID, err)
		_, dbErr := r.Pool.Exec(ctx,
			`UPDATE chain_outbox SET status = $2, last_error = $3, attempts = attempts + 1 WHERE id = $1::uuid`,
			it.ID, StatusFailed, refused.reason)
		return true, dbErr
	case err != nil:
		r.Pool.Exec(ctx, `UPDATE chain_outbox SET last_error = $2, attempts = attempts + 1 WHERE id = $1::uuid`,
			it.ID, err.Error())
		return true, fmt.Errorf("%s %s: %w", it.Kind, it.SubjectID, err)
	}
	return true, r.confirm(ctx, it, hash)
}

// submit sends the item's transaction (unless an earlier attempt already did)
// and waits for it to be included.
func (r *Relayer) submit(ctx context.Context, it item) (common.Hash, error) {
	// An earlier attempt may have sent the transaction and then lost contact.
	// Sending it again would be refused as a duplicate, so look for it first.
	if it.TxHash != nil {
		hash := common.HexToHash(*it.TxHash)
		if receipt, err := r.Backend.TransactionReceipt(ctx, hash); err == nil {
			return hash, receiptErr(receipt)
		}
	}

	opts := *r.Signer
	opts.Context = ctx
	tx, err := r.send(&opts, it)
	if err != nil {
		if reason, ok := revertReason(err); ok {
			return common.Hash{}, errRefused{reason}
		}
		return common.Hash{}, err
	}
	if _, err := r.Pool.Exec(ctx, `UPDATE chain_outbox SET tx_hash = $2 WHERE id = $1::uuid`, it.ID, tx.Hash().Hex()); err != nil {
		return common.Hash{}, err
	}

	waitCtx, cancel := context.WithTimeout(ctx, minedTimeout)
	defer cancel()
	receipt, err := bind.WaitMined(waitCtx, r.Backend, tx)
	if err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), receiptErr(receipt)
}

func receiptErr(receipt *types.Receipt) error {
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errRefused{"transaction reverted"}
	}
	return nil
}

// send builds and sends the contract call for one item.
func (r *Relayer) send(opts *bind.TransactOpts, it item) (*types.Transaction, error) {
	decode := func(dst any) error { return json.Unmarshal(it.Payload, dst) }
	c := r.Contracts

	switch it.Kind {
	case KindOrgRegister:
		var p OrgRegister
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Orgs.Register(opts, MustID(p.OrgID), common.HexToAddress(p.Signer))

	case KindOrgVerified:
		var p OrgVerified
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Orgs.SetVerified(opts, MustID(p.OrgID), p.Verified)

	case KindFacilityRegister:
		var p FacilityRegister
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		hash, err := FromHex(p.DataHash)
		if err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Facilities.Register(opts, MustID(p.FacilityID), MustID(p.OrgID), hash, big.NewInt(p.Capacity))

	case KindFacilitySuspended:
		var p FacilitySuspended
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Facilities.SetSuspended(opts, MustID(p.FacilityID), p.Suspended)

	case KindPassportIssue:
		var p PassportIssue
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		hash, err := FromHex(p.DataHash)
		if err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Passports.Issue(opts, MustID(p.PassportID), Key(p.Commodity), MustID(p.FacilityID),
			big.NewInt(p.Period), big.NewInt(p.Quantity), hash)

	case KindHandover:
		var p HandoverSigned
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		hash, err := FromHex(p.DataHash)
		if err != nil {
			return nil, errRefused{err.Error()}
		}
		sigFrom, err1 := hex.DecodeString(strings.TrimPrefix(p.SigFrom, "0x"))
		sigTo, err2 := hex.DecodeString(strings.TrimPrefix(p.SigTo, "0x"))
		if err1 != nil || err2 != nil {
			return nil, errRefused{"signature is not hex"}
		}
		return c.Custody.RecordHandover(opts, bindings.CustodyManagerHandover{
			PassportId: MustID(p.PassportID),
			FromOrg:    MustID(p.FromOrg),
			ToOrg:      MustID(p.ToOrg),
			Quantity:   big.NewInt(p.Quantity),
			DataHash:   hash,
			Nonce:      big.NewInt(p.Nonce),
			Deadline:   big.NewInt(p.Deadline),
		}, sigFrom, sigTo)

	case KindMaxYieldSet:
		var p MaxYieldSet
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Passports.SetMaxYield(opts, Key(p.Commodity), Key(p.Stage), big.NewInt(p.PartsPerMillion))

	case KindOrgAccredited:
		var p OrgAccredited
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Orgs.SetAccreditation(opts, MustID(p.OrgID), Key(p.Type), uint64(p.ValidFrom), uint64(p.ValidTo))

	case KindTransform:
		var p Transform
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		hash, err := FromHex(p.DataHash)
		if err != nil || len(p.Inputs) != len(p.InputsUsed) {
			return nil, errRefused{"transform payload is malformed"}
		}
		inputs, used := make([][32]byte, len(p.Inputs)), make([]*big.Int, len(p.Inputs))
		for i := range p.Inputs {
			inputs[i], used[i] = MustID(p.Inputs[i]), big.NewInt(p.InputsUsed[i])
		}
		outputs := make([]bindings.PassportRegistryOutput, len(p.Outputs))
		for i, o := range p.Outputs {
			outHash, err := FromHex(o.DataHash)
			if err != nil {
				return nil, errRefused{err.Error()}
			}
			outputs[i] = bindings.PassportRegistryOutput{Id: MustID(o.PassportID), Quantity: big.NewInt(o.Quantity), DataHash: outHash}
		}
		return c.Passports.Transform(opts, inputs, used, Key(p.Stage), outputs, hash)

	case KindAttestationIssue:
		var p AttestationIssue
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		hash, err := FromHex(p.DocHash)
		if err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Attestations.Issue(opts, MustID(p.AttestationID), MustID(p.IssuerOrg), MustID(p.SubjectID),
			Key(p.Type), hash, uint64(p.ValidFrom), uint64(p.ValidTo))

	case KindAttestationRevoke:
		var p AttestationRevoke
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Attestations.Revoke(opts, MustID(p.AttestationID))

	case KindFlag, KindPassportRevoke:
		var p PassportReason
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		hash, err := FromHex(p.ReasonHash)
		if err != nil {
			return nil, errRefused{err.Error()}
		}
		if it.Kind == KindFlag {
			return c.Passports.Flag(opts, MustID(p.PassportID), hash)
		}
		return c.Passports.Revoke(opts, MustID(p.PassportID), hash)

	case KindPassportRetire:
		var p PassportRetire
		if err := decode(&p); err != nil {
			return nil, errRefused{err.Error()}
		}
		return c.Passports.Retire(opts, MustID(p.PassportID))
	}
	return nil, errRefused{"unknown kind " + it.Kind}
}

// The statements below take the item's subject id ($1) and the transaction hash ($2).
const (
	anchorRowFmt = `UPDATE %s SET tx_hash = $2, anchored_at = now() WHERE id = $1::uuid`
	// Journey events that came from the subject row.
	anchorEvents = `UPDATE trace_events SET tx_hash = $2 WHERE ref_id = $1::uuid`
)

func anchorRow(table string) string { return fmt.Sprintf(anchorRowFmt, table) }

// What is written back once an item of each kind is anchored. Kinds not
// listed change no row but their own.
var anchored = map[string][]string{
	KindPassportIssue: {anchorRow("passports"), anchorEvents},
	KindHandover:      {anchorRow("handovers"), anchorEvents},
	// A transform also anchors the passports it created.
	KindTransform: {anchorRow("transforms"), anchorEvents,
		`UPDATE passports SET tx_hash = $2, anchored_at = now() WHERE transform_id = $1::uuid`},
	KindOrgAccredited:     {anchorRow("accreditations")},
	KindAttestationIssue:  {anchorRow("attestations")},
	KindAttestationRevoke: {`UPDATE attestations SET revoke_tx_hash = $2 WHERE id = $1::uuid`},
	KindFlag:              {anchorRow("flags"), anchorEvents},
	// The subject of a revocation is its REVOKE journey event, so the
	// passport keeps the hash of the transaction that created it.
	KindPassportRevoke: {`UPDATE trace_events SET tx_hash = $2 WHERE id = $1::uuid`},
	// Likewise the subject of a retirement is its RETIRE journey event.
	KindPassportRetire: {`UPDATE trace_events SET tx_hash = $2 WHERE id = $1::uuid`},
}

// confirm records that the item is on-chain and writes the transaction hash
// back to the rows it describes.
func (r *Relayer) confirm(ctx context.Context, it item, hash common.Hash) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`UPDATE chain_outbox SET status = $2, tx_hash = $3, last_error = NULL, attempts = attempts + 1 WHERE id = $1::uuid`,
		it.ID, StatusConfirmed, hash.Hex()); err != nil {
		return err
	}
	for _, sql := range anchored[it.Kind] {
		if _, err := tx.Exec(ctx, sql, it.SubjectID, hash.Hex()); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// contractABIs are used to turn a contract's refusal into its error name.
var contractABIs = func() []abi.ABI {
	var out []abi.ABI
	for _, meta := range []*bind.MetaData{
		bindings.OrgRegistryMetaData, bindings.FacilityRegistryMetaData, bindings.PassportRegistryMetaData,
		bindings.CustodyManagerMetaData, bindings.AttestationRegistryMetaData,
	} {
		if parsed, err := meta.GetAbi(); err == nil && parsed != nil {
			out = append(out, *parsed)
		}
	}
	return out
}()

// revertReason reports whether err is the contract refusing the call, and
// names the contract error (e.g. "CapacityExceeded") when it can.
func revertReason(err error) (string, bool) {
	var dataErr interface{ ErrorData() any }
	if errors.As(err, &dataErr) {
		if raw, ok := dataErr.ErrorData().(string); ok {
			if data, decErr := hex.DecodeString(strings.TrimPrefix(raw, "0x")); decErr == nil && len(data) >= 4 {
				var selector [4]byte
				copy(selector[:], data[:4])
				for _, a := range contractABIs {
					if e, err := a.ErrorByID(selector); err == nil {
						return e.Name, true
					}
				}
			}
		}
		return err.Error(), true
	}
	if strings.Contains(err.Error(), "execution reverted") {
		return err.Error(), true
	}
	return "", false
}

// Refused reports whether err is a contract refusing a call (for a read:
// the thing asked about is not on-chain), as opposed to the chain being
// unreachable.
func Refused(err error) bool {
	_, ok := revertReason(err)
	return ok
}
