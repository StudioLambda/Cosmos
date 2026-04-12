package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func TestItCanCreateChaCha20Encrypter(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	_, err := crypto.NewChaCha20(key)

	require.NoError(t, err)
}

func TestItCanEncryptChaCha20(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecryptChaCha20(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

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
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	_, err = e.Decrypt([]byte("short"))

	require.ErrorIs(t, err, crypto.ErrMismatchedChaCha20NonceSize)
}

func TestChaCha20DecryptWithCorruptedCiphertext(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	cypher[len(cypher)-1] ^= 0xFF

	_, err = e.Decrypt(cypher)

	require.Error(t, err)
}

func TestChaCha20CloseZerosKeyMaterial(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	copy(key, "12345678901234567890123456789012")

	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	e.Close()

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
	cypher, err := encrypter.Encrypt(plain)

	require.NoError(t, err)

	decrypter, err := crypto.NewChaCha20(key)
	require.NoError(t, err)

	decrypter.AdditionalData = []byte("context-v2")

	_, err = decrypter.Decrypt(cypher)

	require.Error(t, err)
}

func TestChaCha20AdditionalDataRoundTrip(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	e.AdditionalData = []byte("user-42")

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestChaCha20EncryptProducesDifferentCiphertexts(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")

	c1, err := e.Encrypt(plain)
	require.NoError(t, err)

	c2, err := e.Encrypt(plain)
	require.NoError(t, err)

	require.NotEqual(t, c1, c2)
}

func TestChaCha20DecryptEmptyInput(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewChaCha20(key)

	require.NoError(t, err)

	_, err = e.Decrypt([]byte{})

	require.ErrorIs(t, err, crypto.ErrMismatchedChaCha20NonceSize)
}
