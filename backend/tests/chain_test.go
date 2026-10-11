package tests

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
)

// The handover message is declared once, in a file the Hardhat tests read too.
const vectorPath = "../../contracts/test/vectors/handover.json"

type handoverVector struct {
	Domain struct {
		Name, Version     string
		ChainID           int64  `json:"chainId"`
		VerifyingContract string `json:"verifyingContract"`
	}
	Types struct {
		Handover []struct{ Name, Type string }
	}
	Message struct {
		PassportID string `json:"passportId"`
		FromOrg    string `json:"fromOrg"`
		ToOrg      string `json:"toOrg"`
		Quantity   string
		DataHash   string `json:"dataHash"`
		Nonce      string
		Deadline   string
	}
	PrivateKey, Signer, Digest, Signature string
}

func loadVector(t *testing.T) (handoverVector, chain.Domain, chain.Handover) {
	t.Helper()
	raw, err := os.ReadFile(vectorPath)
	if err != nil {
		t.Fatal(err)
	}
	var v handoverVector
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	b32 := func(s string) [32]byte {
		out, err := chain.FromHex(s)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	num := func(s string) *big.Int {
		n, ok := new(big.Int).SetString(s, 10)
		if !ok {
			t.Fatalf("bad number %q", s)
		}
		return n
	}
	domain := chain.Domain{ChainID: big.NewInt(v.Domain.ChainID), Contract: common.HexToAddress(v.Domain.VerifyingContract)}
	msg := chain.Handover{
		PassportID: b32(v.Message.PassportID), FromOrg: b32(v.Message.FromOrg), ToOrg: b32(v.Message.ToOrg),
		Quantity: num(v.Message.Quantity), DataHash: b32(v.Message.DataHash),
		Nonce: num(v.Message.Nonce), Deadline: num(v.Message.Deadline),
	}
	return v, domain, msg
}

// Go must sign exactly what the contract checks. If either side changes a
// field, its order or its type, this test or its Hardhat twin fails.
func TestHandoverSigningMatchesTheContract(t *testing.T) {
	v, domain, msg := loadVector(t)

	if v.Domain.Name != chain.DomainName || v.Domain.Version != chain.DomainVersion {
		t.Fatalf("domain is %s/%s in the vector, %s/%s in Go", v.Domain.Name, v.Domain.Version, chain.DomainName, chain.DomainVersion)
	}
	if len(v.Types.Handover) != len(chain.HandoverFields) {
		t.Fatalf("vector has %d fields, Go has %d", len(v.Types.Handover), len(chain.HandoverFields))
	}
	for i, f := range v.Types.Handover {
		if got := chain.HandoverFields[i]; got.Name != f.Name || got.Type != f.Type {
			t.Errorf("field %d: vector %s %s, Go %s %s", i, f.Type, f.Name, got.Type, got.Name)
		}
	}

	digest := domain.Digest(msg)
	if chain.Hex(digest) != v.Digest {
		t.Fatalf("digest = %s, want %s", chain.Hex(digest), v.Digest)
	}

	key, err := chain.ParseKey(v.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := chain.Sign(digest, key)
	if err != nil {
		t.Fatal(err)
	}
	if got := "0x" + hex.EncodeToString(sig); got != v.Signature {
		t.Fatalf("signature = %s, want %s", got, v.Signature)
	}
	signer, err := chain.Recover(digest, sig)
	if err != nil || signer != common.HexToAddress(v.Signer) {
		t.Fatalf("recovered %s (%v), want %s", signer, err, v.Signer)
	}
}

func TestSignaturesCannotBeReusedOrForged(t *testing.T) {
	_, domain, msg := loadVector(t)
	digest := domain.Digest(msg)

	hexA, addrA, err := chain.NewKey()
	if err != nil {
		t.Fatal(err)
	}
	_, addrB, _ := chain.NewKey()
	keyA, err := chain.ParseKey(hexA)
	if err != nil {
		t.Fatal(err)
	}
	sig, _ := chain.Sign(digest, keyA)

	if got, err := chain.Recover(digest, sig); err != nil || got != addrA || got == addrB {
		t.Fatalf("recovered %s (%v), want %s", got, err, addrA)
	}

	// Any change to the message, or moving it to another contract or chain,
	// makes the signature point at someone else.
	changes := map[string]func(*chain.Domain, *chain.Handover){
		"quantity":       func(_ *chain.Domain, h *chain.Handover) { h.Quantity = big.NewInt(1) },
		"receiver":       func(_ *chain.Domain, h *chain.Handover) { h.ToOrg[31] ^= 1 },
		"record":         func(_ *chain.Domain, h *chain.Handover) { h.DataHash[0] ^= 1 },
		"handover count": func(_ *chain.Domain, h *chain.Handover) { h.Nonce = new(big.Int).Add(h.Nonce, big.NewInt(1)) },
		"deadline":       func(_ *chain.Domain, h *chain.Handover) { h.Deadline = big.NewInt(1) },
		"other chain":    func(d *chain.Domain, _ *chain.Handover) { d.ChainID = big.NewInt(80002) },
		"other contract": func(d *chain.Domain, _ *chain.Handover) { d.Contract = common.HexToAddress("0x01") },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			d, h := domain, msg
			change(&d, &h)
			if got, _ := chain.Recover(d.Digest(h), sig); got == addrA {
				t.Fatal("signature still valid after the change")
			}
		})
	}

	// Malformed signatures are an error, never a panic or a zero address.
	badRecovery := append(append([]byte{}, sig[:64]...), 99)
	for name, bad := range map[string][]byte{
		"empty": nil, "too short": sig[:64], "bad last byte": badRecovery, "all zero": make([]byte, len(sig)),
	} {
		if _, err := chain.Recover(digest, bad); !errors.Is(err, chain.ErrBadSignature) {
			t.Errorf("%s signature: err = %v", name, err)
		}
	}
}

func TestIDsAndKeys(t *testing.T) {
	id, err := chain.ID(fakeOrgID)
	if err != nil {
		t.Fatal(err)
	}
	// The uuid's bytes sit in the low half; the high half is zero.
	if chain.Hex(id) != "0x"+strings.Repeat("0", 32)+strings.ReplaceAll(fakeOrgID, "-", "") {
		t.Fatalf("id = %s", chain.Hex(id))
	}
	if back, err := chain.FromHex(chain.Hex(id)); err != nil || back != id {
		t.Fatalf("hex round trip failed: %v", err)
	}
	for _, bad := range []string{"", "not-a-uuid", fakeOrgID + "00"} {
		if _, err := chain.ID(bad); err == nil {
			t.Errorf("%q accepted as a uuid", bad)
		}
	}
	if _, err := chain.FromHex("0x1234"); err == nil {
		t.Error("short hex accepted")
	}
	// A name's key is its Keccak-256 hash, as ethers.id(name) gives in the contract tests.
	if got := chain.Key("coffee"); got != [32]byte(crypto.Keccak256([]byte("coffee"))) {
		t.Fatalf("key = %s", chain.Hex(got))
	}
	if chain.Key("coffee") == chain.Key("cocoa") {
		t.Fatal("different names gave the same key")
	}
}

func TestDataHashIsCanonical(t *testing.T) {
	type record struct {
		B string `json:"b"`
		A int64  `json:"a"`
	}
	fromStruct, err := chain.DataHash(record{B: "x<y", A: 9007199254740993}) // beyond float precision
	if err != nil {
		t.Fatal(err)
	}
	fromMap, _ := chain.DataHash(map[string]any{"a": json.Number("9007199254740993"), "b": "x<y"})
	if fromStruct != fromMap {
		t.Fatal("the same record hashed differently depending on field order")
	}
	canonical, _ := chain.CanonicalJSON(record{B: "x<y", A: 9007199254740993})
	if string(canonical) != `{"a":9007199254740993,"b":"x<y"}` {
		t.Fatalf("canonical form = %s", canonical)
	}
	changed, _ := chain.DataHash(record{B: "x<y", A: 9007199254740992})
	if changed == fromStruct {
		t.Fatal("a changed record kept its hash")
	}
	if _, err := chain.DataHash(make(chan int)); err == nil {
		t.Fatal("unencodable record accepted")
	}
}
