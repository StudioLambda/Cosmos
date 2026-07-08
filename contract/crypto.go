package contract

import "errors"

// ErrEncrypterClosed is returned when Encrypt or Decrypt is called
// after [Encrypter.Close] has been called.
var ErrEncrypterClosed = errors.New("encrypter is closed")

// Encrypter defines the interface for encrypting and decrypting data.
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
//	if err := encrypter.Close(); err != nil {
//		return err
//	}
//	_ = plaintext
type Encrypter interface {
	// Encrypt takes a byte slice and returns an encrypted version of it.
	// It returns an error if the encryption operation fails.
	Encrypt(value []byte) ([]byte, error)

	// Decrypt takes an encrypted byte slice and returns the decrypted original value.
	// It returns an error if the decryption operation fails.
	Decrypt(value []byte) ([]byte, error)

	// Close releases encrypter resources and clears key material when applicable.
	Close() error
}
