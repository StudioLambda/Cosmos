package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/crypto"
)

func TestItCanCreateAESEncrypter(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	_, err := crypto.NewAES(key)

	require.NoError(t, err)
}

func TestItCanEncryptAES(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	_, err = e.Encrypt(plain)

	require.NoError(t, err)
}

func TestItCanDecryptAES(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestAESNewWithInvalidKeySize(t *testing.T) {
	t.Parallel()

	_, err := crypto.NewAES([]byte("short"))

	require.Error(t, err)
}

func TestAESNewWith16ByteKey(t *testing.T) {
	t.Parallel()

	key := []byte("1234567890123456")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestAESNewWith24ByteKey(t *testing.T) {
	t.Parallel()

	key := []byte("123456789012345678901234")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestAESDecryptWithShortCiphertext(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	_, err = e.Decrypt([]byte("short"))

	require.ErrorIs(t, err, crypto.ErrMismatchedAESNonceSize)
}

func TestAESDecryptWithCorruptedCiphertext(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	cypher[len(cypher)-1] ^= 0xFF

	_, err = e.Decrypt(cypher)

	require.Error(t, err)
}

func TestAESCloseZerosKeyMaterial(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	copy(key, "12345678901234567890123456789012")

	e, err := crypto.NewAES(key)

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

func TestAESAdditionalDataMustMatchForDecrypt(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")

	encrypter, err := crypto.NewAES(key)
	require.NoError(t, err)

	encrypter.AdditionalData = []byte("context-v1")

	plain := []byte("Hello, World!")
	cypher, err := encrypter.Encrypt(plain)

	require.NoError(t, err)

	decrypter, err := crypto.NewAES(key)
	require.NoError(t, err)

	decrypter.AdditionalData = []byte("context-v2")

	_, err = decrypter.Decrypt(cypher)

	require.Error(t, err)
}

func TestAESAdditionalDataRoundTrip(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	e.AdditionalData = []byte("user-42")

	plain := []byte("Hello, World!")
	cypher, err := e.Encrypt(plain)

	require.NoError(t, err)

	res, err := e.Decrypt(cypher)

	require.NoError(t, err)
	require.Equal(t, plain, res)
}

func TestAESEncryptProducesDifferentCiphertexts(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	plain := []byte("Hello, World!")

	c1, err := e.Encrypt(plain)
	require.NoError(t, err)

	c2, err := e.Encrypt(plain)
	require.NoError(t, err)

	require.NotEqual(t, c1, c2)
}

func TestAESDecryptEmptyInput(t *testing.T) {
	t.Parallel()

	key := []byte("12345678901234567890123456789012")
	e, err := crypto.NewAES(key)

	require.NoError(t, err)

	_, err = e.Decrypt([]byte{})

	require.ErrorIs(t, err, crypto.ErrMismatchedAESNonceSize)
}
