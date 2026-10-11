package chain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// HandoverFields is the handover message, field by field, in signing order.
// It mirrors the Handover struct in CustodyManager.sol.
var HandoverFields = []struct{ Name, Type string }{
	{"passportId", "bytes32"},
	{"fromOrg", "bytes32"},
	{"toOrg", "bytes32"},
	{"quantity", "uint256"},
	{"dataHash", "bytes32"},
	{"nonce", "uint256"},
	{"deadline", "uint256"},
}

// HandoverType is the EIP-712 type string built from HandoverFields.
func HandoverType() string {
	parts := make([]string, len(HandoverFields))
	for i, f := range HandoverFields {
		parts[i] = f.Type + " " + f.Name
	}
	return "Handover(" + strings.Join(parts, ",") + ")"
}

const domainType = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"

// Handover is the message both organisations sign.
type Handover struct {
	PassportID [32]byte
	FromOrg    [32]byte
	ToOrg      [32]byte
	Quantity   *big.Int
	DataHash   [32]byte
	Nonce      *big.Int // the passport's handover count
	Deadline   *big.Int // unix seconds
}

// Domain binds a signature to one CustodyManager on one chain, so a
// signature made for a test network is worthless on the real one.
type Domain struct {
	ChainID  *big.Int
	Contract common.Address
}

func word(n *big.Int) []byte { return common.LeftPadBytes(n.Bytes(), 32) }

func (d Domain) separator() []byte {
	return crypto.Keccak256(
		crypto.Keccak256([]byte(domainType)),
		crypto.Keccak256([]byte(DomainName)),
		crypto.Keccak256([]byte(DomainVersion)),
		word(d.ChainID),
		common.LeftPadBytes(d.Contract.Bytes(), 32),
	)
}

// Digest is the exact 32 bytes each party signs: EIP-712's hash of the
// message under this domain. CustodyManager.hashHandover computes the same.
func (d Domain) Digest(h Handover) [32]byte {
	structHash := crypto.Keccak256(
		crypto.Keccak256([]byte(HandoverType())),
		h.PassportID[:], h.FromOrg[:], h.ToOrg[:],
		word(h.Quantity),
		h.DataHash[:],
		word(h.Nonce), word(h.Deadline),
	)
	return crypto.Keccak256Hash([]byte{0x19, 0x01}, d.separator(), structHash)
}

// The last byte of a signature says which of two possible keys signed.
// Contracts expect it as 27 or 28; the signing library produces 0 or 1.
const recoveryOffset = 27

var ErrBadSignature = errors.New("signature is malformed")

// Sign signs a digest. The result is 65 bytes, in the form the contract accepts.
func Sign(digest [32]byte, key *ecdsa.PrivateKey) ([]byte, error) {
	sig, err := crypto.Sign(digest[:], key)
	if err != nil {
		return nil, err
	}
	sig[crypto.RecoveryIDOffset] += recoveryOffset
	return sig, nil
}

// Recover returns the address whose key produced the signature over digest.
func Recover(digest [32]byte, sig []byte) (common.Address, error) {
	if len(sig) != crypto.SignatureLength || sig[crypto.RecoveryIDOffset] < recoveryOffset {
		return common.Address{}, ErrBadSignature
	}
	raw := make([]byte, len(sig))
	copy(raw, sig)
	raw[crypto.RecoveryIDOffset] -= recoveryOffset
	pub, err := crypto.SigToPub(digest[:], raw)
	if err != nil {
		return common.Address{}, ErrBadSignature
	}
	return crypto.PubkeyToAddress(*pub), nil
}

// NewKey creates a signing key for an organisation and returns it as hex
// together with the address that identifies the organisation on-chain.
func NewKey() (hexKey string, address common.Address, err error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", common.Address{}, err
	}
	return hex.EncodeToString(crypto.FromECDSA(key)), crypto.PubkeyToAddress(key.PublicKey), nil
}

// ParseKey reads a key written by NewKey; a leading 0x is accepted.
func ParseKey(hexKey string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(strings.TrimPrefix(hexKey, "0x"))
}
