package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// AES implements contract.Encrypter using AES-GCM (Galois/Counter Mode)
// authenticated encryption. The nonce is generated randomly for each
// Encrypt call and prepended to the ciphertext.
type AES struct {
	key []byte
}

// ErrMismatchedAESNonceSize is returned when the ciphertext provided
// to Decrypt is shorter than the expected GCM nonce size, indicating
// truncated or corrupted data.
var ErrMismatchedAESNonceSize = errors.New("mismatched nonce size")

// NewAES creates an AES encrypter with the provided key.
// The key must be 16, 24, or 32 bytes for AES-128, AES-192,
// or AES-256 respectively.
func NewAES(key []byte) (*AES, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, aes.KeySizeError(len(key))
	}

	return &AES{key: key}, nil
}

// Encrypt encrypts the plaintext using AES-GCM with a random nonce.
// The returned slice contains the nonce followed by the ciphertext
// and authentication tag.
func (encrypter *AES) Encrypt(value []byte) ([]byte, error) {
	block, err := aes.NewCipher(encrypter.key)

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

	return gcm.Seal(nonce, nonce, value, nil), nil
}

// Decrypt decrypts AES-GCM ciphertext that has the nonce prepended.
// Returns ErrMismatchedAESNonceSize if the input is too short to
// contain a valid nonce.
func (encrypter *AES) Decrypt(value []byte) ([]byte, error) {
	block, err := aes.NewCipher(encrypter.key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	if len(value) < nonceSize {
		return nil, ErrMismatchedAESNonceSize
	}

	nonce, ciphertext := value[:nonceSize], value[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
