package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// ID is the 32-byte form of a database uuid: its 16 bytes, right-aligned.
// The same row therefore has the same id in the database and on-chain.
func ID(uuid string) ([32]byte, error) {
	var out [32]byte
	raw, err := hex.DecodeString(strings.ReplaceAll(uuid, "-", ""))
	if err != nil || len(raw) != 16 {
		return out, fmt.Errorf("%q is not a uuid", uuid)
	}
	copy(out[16:], raw)
	return out, nil
}

// MustID is ID for uuids that came from the database.
func MustID(uuid string) [32]byte {
	id, err := ID(uuid)
	if err != nil {
		panic(err)
	}
	return id
}

// Key is the 32-byte form of a name such as a commodity ("coffee") or an
// event kind ("EXPORT"): its Keccak-256 hash.
func Key(name string) [32]byte {
	return crypto.Keccak256Hash([]byte(name))
}

// DataHash is the fingerprint of an off-chain record: SHA-256 of its
// canonical JSON. Canonical means object keys are sorted and there is no
// insignificant whitespace, so the same record always gives the same hash
// whoever computes it.
func DataHash(record any) ([32]byte, error) {
	canonical, err := CanonicalJSON(record)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

// CanonicalJSON encodes a record with sorted object keys and exact numbers.
func CanonicalJSON(record any) ([]byte, error) {
	first, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	// Decoding into generic maps and encoding again sorts every object's keys.
	var generic any
	dec := json.NewDecoder(bytes.NewReader(first))
	dec.UseNumber() // keep numbers exactly as written
	if err := dec.Decode(&generic); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(generic); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// Hex writes 32 bytes as 0x-prefixed hex, the form stored in the database.
func Hex(b [32]byte) string { return "0x" + hex.EncodeToString(b[:]) }

// FromHex reads a value written by Hex.
func FromHex(s string) ([32]byte, error) {
	var out [32]byte
	raw, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil || len(raw) != len(out) {
		return out, fmt.Errorf("%q is not a 32-byte hex value", s)
	}
	copy(out[:], raw)
	return out, nil
}
