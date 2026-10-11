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
	Orgs       *bindings.OrgRegistry
	Facilities *bindings.FacilityRegistry
	Passports  *bindings.PassportRegistry
	Custody    *bindings.CustodyManager
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
	c.Custody, err = bindings.NewCustodyManager(a.Custody, backend)
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
	Signer    *bind.TransactOpts // the platform key: pays for and sends every transaction
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
	}
	return nil, errRefused{"unknown kind " + it.Kind}
}

// Which table's row gets the transaction hash once an item is anchored.
var anchoredTable = map[string]string{
	KindPassportIssue: "passports",
	KindHandover:      "handovers",
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
	if table, ok := anchoredTable[it.Kind]; ok {
		// table comes from the map above, never from data.
		if _, err := tx.Exec(ctx,
			`UPDATE `+table+` SET tx_hash = $2, anchored_at = now() WHERE id = $1::uuid`, it.SubjectID, hash.Hex()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE trace_events SET tx_hash = $2 WHERE ref_id = $1::uuid`, it.SubjectID, hash.Hex()); err != nil {
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
