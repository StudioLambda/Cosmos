package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// AES implements contract.Encrypter using AES-GCM (Galois/Counter
// Mode) authenticated encryption. The cipher.AEAD is created once
// during construction and reused for every operation. This is safe
// because GCM instances are safe for concurrent use with different
// nonces. The nonce is generated randomly for each Encrypt call
// and prepended to the ciphertext.
type AES struct {
	// key is the raw AES key material (16, 24, or 32 bytes).
	key []byte

	// gcm holds the pre-computed AEAD cipher created from the
	// key during construction, avoiding repeated cipher setup
	// on every Encrypt and Decrypt call.
	gcm cipher.AEAD

	// AdditionalData is optional additional authenticated data (AAD)
	// passed to the GCM Seal and Open operations. AAD is authenticated
	// but not encrypted, which allows binding the ciphertext to a
	// particular context (e.g. a user ID, resource path, or version)
	// so that ciphertext cannot be transplanted between contexts.
	// When set, the same AAD must be provided for both encryption
	// and decryption or authentication will fail.
	AdditionalData []byte
}

// ErrMismatchedAESNonceSize is returned when the ciphertext provided
// to Decrypt is shorter than the expected GCM nonce size, indicating
// truncated or corrupted data.
var ErrMismatchedAESNonceSize = errors.New("mismatched nonce size")

// NewAES creates an AES encrypter with the provided key. The key
// must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
// respectively. The GCM cipher is constructed eagerly so that
// Encrypt and Decrypt avoid repeated setup overhead.
func NewAES(key []byte) (*AES, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, aes.KeySizeError(len(key))
	}

	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	return &AES{key: key, gcm: gcm}, nil
}

// Encrypt encrypts the plaintext using AES-GCM with a random
// nonce. The returned slice contains the nonce followed by the
// ciphertext and authentication tag.
func (encrypter *AES) Encrypt(value []byte) ([]byte, error) {
	nonce := make([]byte, encrypter.gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return encrypter.gcm.Seal(
		nonce, nonce, value, encrypter.AdditionalData,
	), nil
}

// Decrypt decrypts AES-GCM ciphertext that has the nonce
// prepended. Returns ErrMismatchedAESNonceSize if the input is
// too short to contain a valid nonce.
func (encrypter *AES) Decrypt(value []byte) ([]byte, error) {
	nonceSize := encrypter.gcm.NonceSize()

	if len(value) < nonceSize {
		return nil, ErrMismatchedAESNonceSize
	}

	nonce, ciphertext := value[:nonceSize], value[nonceSize:]
	plaintext, err := encrypter.gcm.Open(
		nil, nonce, ciphertext, encrypter.AdditionalData,
	)

	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Close zeros the key material from memory. Callers should
// defer Close immediately after creating the encrypter to
// ensure key material does not linger in process memory.
func (encrypter *AES) Close() {
	for i := range encrypter.key {
		encrypter.key[i] = 0
	}
}
