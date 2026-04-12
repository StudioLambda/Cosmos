package hash_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/hash"
)

func TestItCanHashArgon2Passwords(t *testing.T) {
	t.Parallel()

	h := hash.NewArgon2()
	content := []byte("hello, world")

	r, err := h.Hash(content)

	require.NoError(t, err)
	require.Greater(t, len(r), 0)
}

func TestItCanCheckHashedArgon2Hashes(t *testing.T) {
	t.Parallel()

	h := hash.NewArgon2()

	r, err := h.Hash([]byte("hello, world"))

	require.NoError(t, err)

	ok, err := h.Check([]byte("hello, world"), r)

	require.NoError(t, err)
	require.True(t, ok)
}

func TestArgon2WithCustomConfig(t *testing.T) {
	t.Parallel()

	config := hash.Argon2Config{
		HashLength:  32,
		SaltLength:  16,
		TimeCost:    1,
		MemoryCost:  64 * 1024,
		Parallelism: 2,
		Mode:        2,
		Version:     0,
	}

	h := hash.NewArgon2With(config)

	r, err := h.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.Greater(t, len(r), 0)
}

func TestArgon2CheckWrongPasswordReturnsFalse(t *testing.T) {
	t.Parallel()

	h := hash.NewArgon2()

	hashed, err := h.Hash([]byte("correct-password"))

	require.NoError(t, err)

	ok, err := h.Check([]byte("wrong-password"), hashed)

	require.NoError(t, err)
	require.False(t, ok)
}

func TestArgon2HashZerosInputPassword(t *testing.T) {
	t.Parallel()

	h := hash.NewArgon2()
	password := []byte("sensitive-data")

	_, err := h.Hash(password)

	require.NoError(t, err)

	allZero := true
	for _, b := range password {
		if b != 0 {
			allZero = false
			break
		}
	}

	require.True(t, allZero)
}

func TestArgon2CheckZerosInputPassword(t *testing.T) {
	t.Parallel()

	h := hash.NewArgon2()

	hashed, err := h.Hash([]byte("hello"))

	require.NoError(t, err)

	password := []byte("hello")

	_, err = h.Check(password, hashed)

	require.NoError(t, err)

	allZero := true
	for _, b := range password {
		if b != 0 {
			allZero = false
			break
		}
	}

	require.True(t, allZero)
}
