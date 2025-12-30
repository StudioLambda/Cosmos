package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type AES struct {
	key []byte
}

var ErrMissmatchedAESNonceSize = errors.New("missmatched nonce size")

// NewAES creates a new AES encrypter with the provided key.
// Key should be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func NewAES(key []byte) (*AES, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, aes.KeySizeError(len(key))
	}

	return &AES{key: key}, nil
}

// Encrypt encrypts the plaintext using AES-GCM.
// Returns ciphertext with nonce prepended.
func (e *AES) Encrypt(value []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Prepend nonce to ciphertext
	return gcm.Seal(nonce, nonce, value, nil), nil
}

// Decrypt decrypts the ciphertext using AES-GCM.
// Expects nonce to be prepended to ciphertext.
func (e *AES) Decrypt(value []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	if len(value) < nonceSize {
		return nil, ErrMissmatchedAESNonceSize
	}

	nonce, ciphertext := value[:nonceSize], value[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
