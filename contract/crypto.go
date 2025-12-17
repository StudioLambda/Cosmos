package contract

// Encrypter defines the interface for encrypting and decrypting data.
// Implementations of Encrypter are responsible for securing data through
// encryption and recovering the original data through decryption.
type Encrypter interface {
	// Encrypt takes a byte slice and returns an encrypted version of it.
	// It returns an error if the encryption operation fails.
	Encrypt(value []byte) ([]byte, error)

	// Decrypt takes an encrypted byte slice and returns the decrypted original value.
	// It returns an error if the decryption operation fails.
	Decrypt(value []byte) ([]byte, error)
}
