package chacha20

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

type Encrypter struct {
	aead cipher.AEAD
}

var ErrMissmatchedNonceSize = errors.New("missmatched nonce size")

// NewEncrypter creates a new ChaCha20-Poly1305 encrypter.
// Key must be exactly 32 bytes.
func NewEncrypter(key []byte) (*Encrypter, error) {
	aead, err := chacha20poly1305.New(key)

	if err != nil {
		return nil, err
	}

	return &Encrypter{aead: aead}, nil
}

// Encrypt encrypts the plaintext using ChaCha20-Poly1305.
// Returns ciphertext with nonce prepended.
func (e *Encrypter) Encrypt(value []byte) ([]byte, error) {
	nonce := make([]byte, e.aead.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Prepend nonce to ciphertext
	return e.aead.Seal(nonce, nonce, value, nil), nil
}

// Decrypt decrypts the ciphertext using ChaCha20-Poly1305.
// Expects nonce to be prepended to ciphertext.
func (e *Encrypter) Decrypt(value []byte) ([]byte, error) {
	nonceSize := e.aead.NonceSize()

	if len(value) < nonceSize {
		return nil, ErrMissmatchedNonceSize
	}

	nonce, ciphertext := value[:nonceSize], value[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
