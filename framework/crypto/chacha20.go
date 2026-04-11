package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// ChaCha20 implements contract.Encrypter using ChaCha20-Poly1305
// authenticated encryption. The AEAD cipher is created once at
// construction time and reused for every Encrypt/Decrypt call.
type ChaCha20 struct {
	// aead is the underlying ChaCha20-Poly1305 AEAD cipher.
	aead cipher.AEAD

	// key is the raw key material, retained so that Close
	// can zero it from memory.
	key []byte

	// AdditionalData is optional additional authenticated data (AAD)
	// passed to the AEAD Seal and Open operations. AAD is
	// authenticated but not encrypted, which allows binding the
	// ciphertext to a particular context (e.g. a user ID, resource
	// path, or version) so that ciphertext cannot be transplanted
	// between contexts. When set, the same AAD must be provided for
	// both encryption and decryption or authentication will fail.
	AdditionalData []byte
}

// ErrMismatchedChaCha20NonceSize is returned when the ciphertext
// provided to Decrypt is shorter than the expected nonce size,
// indicating truncated or corrupted data.
var ErrMismatchedChaCha20NonceSize = errors.New("mismatched nonce size")

// NewChaCha20 creates a ChaCha20-Poly1305 encrypter with the given
// key. The key must be exactly 32 bytes.
func NewChaCha20(key []byte) (*ChaCha20, error) {
	aead, err := chacha20poly1305.New(key)

	if err != nil {
		return nil, err
	}

	return &ChaCha20{aead: aead, key: key}, nil
}

// Encrypt encrypts the plaintext using ChaCha20-Poly1305 with a
// random nonce. The returned slice contains the nonce followed by
// the ciphertext and authentication tag.
func (encrypter *ChaCha20) Encrypt(value []byte) ([]byte, error) {
	nonce := make([]byte, encrypter.aead.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return encrypter.aead.Seal(nonce, nonce, value, encrypter.AdditionalData), nil
}

// Decrypt decrypts ChaCha20-Poly1305 ciphertext that has the nonce
// prepended. Returns ErrMismatchedChaCha20NonceSize if the input is
// too short to contain a valid nonce.
func (encrypter *ChaCha20) Decrypt(value []byte) ([]byte, error) {
	nonceSize := encrypter.aead.NonceSize()

	if len(value) < nonceSize {
		return nil, ErrMismatchedChaCha20NonceSize
	}

	nonce, ciphertext := value[:nonceSize], value[nonceSize:]
	plaintext, err := encrypter.aead.Open(nil, nonce, ciphertext, encrypter.AdditionalData)

	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Close zeros the key material from memory. Callers should
// defer Close immediately after creating the encrypter to
// ensure key material does not linger in process memory.
func (encrypter *ChaCha20) Close() {
	for i := range encrypter.key {
		encrypter.key[i] = 0
	}
}
