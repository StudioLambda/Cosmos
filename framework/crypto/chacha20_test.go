package crypto_test

import (
	"testing"

	"github.com/studiolambda/cosmos/framework/crypto"

	"github.com/stretchr/testify/require"
)

func TestChaCha20NewCreatesEncrypter(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	_, err := crypto.NewChaCha20(key)

	require.NoError(t, err)
}

func TestChaCha20EncryptSucceeds(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = encrypter.Encrypt(plain)

	require.NoError(t, err)
}

func TestChaCha20EncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	ciphertext, err := encrypter.Encrypt(plain)

	require.NoError(t, err)

	res, err := encrypter.Decrypt(ciphertext)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestChaCha20NewWithInvalidKeySize(t *testing.T) {
	t.Parallel()

	_, err := crypto.NewChaCha20([]byte("short"))

	require.Error(t, err)
}

func TestChaCha20DecryptWithShortCiphertext(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	_, err = encrypter.Decrypt([]byte("short"))

	require.ErrorIs(t, err, crypto.ErrMismatchedChaCha20NonceSize)
}

func TestChaCha20DecryptWithCorruptedCiphertext(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	ciphertext, err := encrypter.Encrypt(plain)

	require.NoError(t, err)

	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = encrypter.Decrypt(ciphertext)

	require.Error(t, err)
}

func TestChaCha20CloseZerosKeyMaterial(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	copy(key, "12345678901234567890123456789012")

	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	encrypter.Close()

	allZero := true
	for _, b := range key {
		if b != 0 {
			allZero = false
			break
		}
	}

	require.True(t, allZero)
}

func TestChaCha20AdditionalDataMustMatchForDecrypt(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")

	encrypter, err := crypto.NewChaCha20(key)
	require.NoError(t, err)

	encrypter.AdditionalData = []byte("context-v1")

	plain := []byte("Hello, World!")
	ciphertext, err := encrypter.Encrypt(plain)

	require.NoError(t, err)

	decrypter, err := crypto.NewChaCha20(key)
	require.NoError(t, err)

	decrypter.AdditionalData = []byte("context-v2")

	_, err = decrypter.Decrypt(ciphertext)

	require.Error(t, err)
}

func TestChaCha20AdditionalDataRoundTrip(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	encrypter.AdditionalData = []byte("user-42")

	plain := []byte("Hello, World!")
	ciphertext, err := encrypter.Encrypt(plain)

	require.NoError(t, err)

	res, err := encrypter.Decrypt(ciphertext)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestChaCha20EncryptProducesDifferentCiphertexts(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")

	c1, err := encrypter.Encrypt(plain)
	require.NoError(t, err)

	c2, err := encrypter.Encrypt(plain)
	require.NoError(t, err)

	require.NotEqual(t, c1, c2)
}

func TestChaCha20DecryptEmptyInput(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	encrypter, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	_, err = encrypter.Decrypt([]byte{})

	require.ErrorIs(t, err, crypto.ErrMismatchedChaCha20NonceSize)
}
