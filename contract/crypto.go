package contract

import (
	"encoding/json/v2"
	"errors"
	"fmt"
)

// ErrEncrypterClosed is returned when Encrypt or Decrypt is called
// after the concrete encrypter has been closed.
var ErrEncrypterClosed = errors.New("encrypter is closed")

// EncrypterDriver defines the interface for encrypting and decrypting raw data.
// Implementations of Encrypter are responsible for securing data through
// encryption and recovering the original data through decryption.
//
// Example:
//
//	ciphertext, err := encrypter.Encrypt([]byte("secret"))
//	if err != nil {
//		return err
//	}
//	plaintext, err := encrypter.Decrypt(ciphertext)
//	if err != nil {
//		return err
//	}
//	_ = plaintext
type EncrypterDriver interface {
	// Encrypt takes a byte slice and returns an encrypted version of it.
	// It returns an error if the encryption operation fails.
	Encrypt(value []byte) ([]byte, error)

	// Decrypt takes an encrypted byte slice and returns the decrypted original value.
	// It returns an error if the decryption operation fails.
	Decrypt(value []byte) ([]byte, error)
}

// Encrypter provides typed encryption over an [EncrypterDriver].
type Encrypter struct {
	driver EncrypterDriver
}

// NewEncrypter creates a new Encrypter that delegates to driver.
func NewEncrypter(driver EncrypterDriver) *Encrypter {
	return &Encrypter{driver: driver}
}

// Driver returns the underlying [EncrypterDriver].
func (encrypter *Encrypter) Driver() EncrypterDriver {
	return encrypter.driver
}

// Encrypt encrypts a JSON-encoded value.
func (encrypter *Encrypter) Encrypt[T any](value T) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode encrypted value: %w", err)
	}

	return encrypter.driver.Encrypt(encoded)
}

// Decrypt decrypts value and JSON-decodes it into T.
func (encrypter *Encrypter) Decrypt[T any](value []byte) (res T, err error) {
	plaintext, err := encrypter.driver.Decrypt(value)
	if err != nil {
		return res, err
	}

	if err := json.Unmarshal(plaintext, &res); err != nil {
		return res, fmt.Errorf("decode decrypted value: %w", err)
	}

	return res, nil
}

// EncryptRaw encrypts raw bytes.
func (encrypter *Encrypter) EncryptRaw(value []byte) ([]byte, error) {
	return encrypter.driver.Encrypt(value)
}

// DecryptRaw decrypts raw bytes.
func (encrypter *Encrypter) DecryptRaw(value []byte) ([]byte, error) {
	return encrypter.driver.Decrypt(value)
}
