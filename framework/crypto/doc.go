// Package crypto provides authenticated encryption implementations.
//
// Both AES-GCM and ChaCha20-Poly1305 satisfy contract encryption interfaces and
// support optional additional authenticated data (AAD).
//
// # Memory hygiene
//
// Types expose Close methods that clear key material; callers should defer
// cleanup when lifecycle allows.
//
// Example
//
//	enc, err := crypto.NewAES(key)
//	if err != nil {
//		return err
//	}
//	defer enc.Close()
//
//	enc.AdditionalData = []byte("tenant:acme")
//	ciphertext, err := enc.Encrypt(plaintext)
//	_ = ciphertext
//	_ = err
package crypto
