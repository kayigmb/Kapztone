package chain

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kayigmb/chaintrack/backend/internal/chain/bindings"
)

// A sender turns one outbox payload into its contract call. A payload that
// cannot be read is refused: retrying would never fix it.
type sender func(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error)

var senders = map[string]sender{
	KindOrgRegister:       sendOrgRegister,
	KindOrgVerified:       sendOrgVerified,
	KindOrgAccredited:     sendOrgAccredited,
	KindFacilityRegister:  sendFacilityRegister,
	KindFacilitySuspended: sendFacilitySuspended,
	KindPassportIssue:     sendPassportIssue,
	KindHandover:          sendHandover,
	KindMaxYieldSet:       sendMaxYield,
	KindTransform:         sendTransform,
	KindAttestationIssue:  sendAttestationIssue,
	KindAttestationRevoke: sendAttestationRevoke,
	KindFlag:              sendFlag,
	KindPassportRevoke:    sendPassportRevoke,
	KindPassportRetire:    sendPassportRetire,
}

func decode[T any](payload []byte) (T, error) {
	var p T
	if err := json.Unmarshal(payload, &p); err != nil {
		return p, errRefused{err.Error()}
	}
	return p, nil
}

func hashOf(s string) ([32]byte, error) {
	h, err := FromHex(s)
	if err != nil {
		return h, errRefused{err.Error()}
	}
	return h, nil
}

func sendOrgRegister(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[OrgRegister](payload)
	if err != nil {
		return nil, err
	}
	return c.Orgs.Register(opts, MustID(p.OrgID), common.HexToAddress(p.Signer))
}

func sendOrgVerified(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[OrgVerified](payload)
	if err != nil {
		return nil, err
	}
	return c.Orgs.SetVerified(opts, MustID(p.OrgID), p.Verified)
}

func sendOrgAccredited(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[OrgAccredited](payload)
	if err != nil {
		return nil, err
	}
	return c.Orgs.SetAccreditation(opts, MustID(p.OrgID), Key(p.Type), uint64(p.ValidFrom), uint64(p.ValidTo))
}

func sendFacilityRegister(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[FacilityRegister](payload)
	if err != nil {
		return nil, err
	}
	hash, err := hashOf(p.DataHash)
	if err != nil {
		return nil, err
	}
	return c.Facilities.Register(opts, MustID(p.FacilityID), MustID(p.OrgID), hash, big.NewInt(p.Capacity))
}

func sendFacilitySuspended(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[FacilitySuspended](payload)
	if err != nil {
		return nil, err
	}
	return c.Facilities.SetSuspended(opts, MustID(p.FacilityID), p.Suspended)
}

func sendPassportIssue(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[PassportIssue](payload)
	if err != nil {
		return nil, err
	}
	hash, err := hashOf(p.DataHash)
	if err != nil {
		return nil, err
	}
	return c.Passports.Issue(opts, MustID(p.PassportID), Key(p.Commodity), MustID(p.FacilityID),
		big.NewInt(p.Period), big.NewInt(p.Quantity), hash)
}

func sendHandover(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[HandoverSigned](payload)
	if err != nil {
		return nil, err
	}
	hash, err := hashOf(p.DataHash)
	if err != nil {
		return nil, err
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

func sendMaxYield(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[MaxYieldSet](payload)
	if err != nil {
		return nil, err
	}
	return c.Passports.SetMaxYield(opts, Key(p.Commodity), Key(p.Stage), big.NewInt(p.PartsPerMillion))
}

func sendTransform(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[Transform](payload)
	if err != nil {
		return nil, err
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
		outHash, err := hashOf(o.DataHash)
		if err != nil {
			return nil, err
		}
		outputs[i] = bindings.PassportRegistryOutput{Id: MustID(o.PassportID), Quantity: big.NewInt(o.Quantity), DataHash: outHash}
	}
	return c.Passports.Transform(opts, inputs, used, Key(p.Stage), outputs, hash)
}

func sendAttestationIssue(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[AttestationIssue](payload)
	if err != nil {
		return nil, err
	}
	hash, err := hashOf(p.DocHash)
	if err != nil {
		return nil, err
	}
	return c.Attestations.Issue(opts, MustID(p.AttestationID), MustID(p.IssuerOrg), MustID(p.SubjectID),
		Key(p.Type), hash, uint64(p.ValidFrom), uint64(p.ValidTo))
}

func sendAttestationRevoke(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[AttestationRevoke](payload)
	if err != nil {
		return nil, err
	}
	return c.Attestations.Revoke(opts, MustID(p.AttestationID))
}

func sendFlag(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[PassportReason](payload)
	if err != nil {
		return nil, err
	}
	hash, err := hashOf(p.ReasonHash)
	if err != nil {
		return nil, err
	}
	return c.Passports.Flag(opts, MustID(p.PassportID), hash)
}

func sendPassportRevoke(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[PassportReason](payload)
	if err != nil {
		return nil, err
	}
	hash, err := hashOf(p.ReasonHash)
	if err != nil {
		return nil, err
	}
	return c.Passports.Revoke(opts, MustID(p.PassportID), hash)
}

func sendPassportRetire(c Contracts, opts *bind.TransactOpts, payload []byte) (*types.Transaction, error) {
	p, err := decode[PassportRetire](payload)
	if err != nil {
		return nil, err
	}
	return c.Passports.Retire(opts, MustID(p.PassportID))
}
