package hash_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/hash"

	"golang.org/x/crypto/bcrypt"
)

func TestItCanHashBcryptPasswords(t *testing.T) {
	t.Parallel()

	h := hash.NewBcrypt()
	content := []byte("hello, world")

	r, err := h.Hash(content)

	require.NoError(t, err)
	require.Greater(t, len(r), 0)
}

func TestItCanCheckHashedBcryptHashes(t *testing.T) {
	t.Parallel()

	h := hash.NewBcrypt()

	r, err := h.Hash([]byte("hello, world"))

	require.NoError(t, err)

	ok, err := h.Check([]byte("hello, world"), r)

	require.NoError(t, err)
	require.True(t, ok)
}

func TestBcryptWithDefaultOptions(t *testing.T) {
	t.Parallel()

	h := hash.NewBcryptWith(hash.BcryptOptions{})

	r, err := h.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.Greater(t, len(r), 0)

	ok, err := h.Check([]byte("hello, world"), r)

	require.NoError(t, err)
	require.True(t, ok)
}

func TestBcryptCheckWrongPasswordReturnsFalse(t *testing.T) {
	t.Parallel()

	h := hash.NewBcrypt()

	hashed, err := h.Hash([]byte("correct-password"))

	require.NoError(t, err)

	ok, err := h.Check([]byte("wrong-password"), hashed)

	require.NoError(t, err)
	require.False(t, ok)
}

func TestBcryptCheckCorruptedHashReturnsError(t *testing.T) {
	t.Parallel()

	h := hash.NewBcrypt()

	ok, err := h.Check([]byte("password"), []byte("not-a-hash"))

	require.Error(t, err)
	require.False(t, ok)
}

func TestBcryptHashZerosInputPassword(t *testing.T) {
	t.Parallel()

	h := hash.NewBcrypt()
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

func TestBcryptCheckZerosInputPassword(t *testing.T) {
	t.Parallel()

	h := hash.NewBcrypt()

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

func TestBcryptNewBcryptWithCustomCost(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcryptWith(hash.BcryptOptions{Cost: 14})

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)

	cost, err := bcrypt.Cost(hashed)

	require.NoError(t, err)
	require.Equal(t, 14, cost)
}

func TestBcryptNeedsRehashReturnsTrueForDifferentCost(t *testing.T) {
	t.Parallel()

	hasher12 := hash.NewBcryptWith(hash.BcryptOptions{Cost: 12})
	hasher14 := hash.NewBcryptWith(hash.BcryptOptions{Cost: 14})

	hashed, err := hasher12.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.True(t, hasher14.NeedsRehash(hashed))
}

func TestBcryptNeedsRehashReturnsFalseForSameCost(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcryptWith(hash.BcryptOptions{Cost: 12})

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.False(t, hasher.NeedsRehash(hashed))
}

func TestBcryptNeedsRehashReturnsTrueForInvalidHash(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()

	require.True(t, hasher.NeedsRehash([]byte("not-a-valid-hash")))
}
