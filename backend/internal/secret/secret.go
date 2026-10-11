// Package secret encrypts small secrets (2FA seeds, later signing keys) before
// they are stored, using AES-256-GCM with a master key from the environment.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// KeyBytes is the master key length: 32 bytes, given as 64 hex characters.
const KeyBytes = 32

type Box struct {
	aead cipher.AEAD
}

// New builds a Box from a hex-encoded master key, e.g. `openssl rand -hex 32`.
func New(hexKey string) (*Box, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil || len(key) != KeyBytes {
		return nil, fmt.Errorf("encryption key must be %d hex characters", KeyBytes*2)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{aead: aead}, nil
}

// Seal encrypts plaintext. The result carries its own random nonce, so sealing
// the same value twice gives different output.
func (b *Box) Seal(plaintext string) (string, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.aead.Seal(nonce, nonce, []byte(plaintext), nil)), nil
}

// Open decrypts a value from Seal. It fails if the value was altered or was
// sealed with a different key.
func (b *Box) Open(sealed string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil || len(raw) < b.aead.NonceSize() {
		return "", errors.New("sealed value is malformed")
	}
	nonce, ciphertext := raw[:b.aead.NonceSize()], raw[b.aead.NonceSize():]
	plain, err := b.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("sealed value cannot be opened")
	}
	return string(plain), nil
}
