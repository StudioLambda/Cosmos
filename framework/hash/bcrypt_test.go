package hash_test

import (
	"testing"

	"github.com/studiolambda/cosmos/framework/hash"

	"github.com/stretchr/testify/require"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHashProducesOutput(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()
	content := []byte("hello, world")

	hashed, err := hasher.Hash(content)

	require.NoError(t, err)
	require.Greater(t, len(hashed), 0)
}

func TestBcryptCheckMatchesCorrectPassword(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)

	ok, err := hasher.Check([]byte("hello, world"), hashed)

	require.NoError(t, err)
	require.True(t, ok)
}

func TestBcryptWithDefaultOptions(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcryptWith(hash.BcryptOptions{})

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.Greater(t, len(hashed), 0)

	ok, err := hasher.Check([]byte("hello, world"), hashed)

	require.NoError(t, err)
	require.True(t, ok)
}

func TestBcryptCheckWrongPasswordReturnsFalse(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()

	hashed, err := hasher.Hash([]byte("correct-password"))

	require.NoError(t, err)

	ok, err := hasher.Check([]byte("wrong-password"), hashed)

	require.NoError(t, err)
	require.False(t, ok)
}

func TestBcryptCheckCorruptedHashReturnsError(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()

	ok, err := hasher.Check([]byte("password"), []byte("not-a-hash"))

	require.Error(t, err)
	require.False(t, ok)
}

func TestBcryptHashZerosInputPassword(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()
	password := []byte("sensitive-data")

	_, err := hasher.Hash(password)

	require.NoError(t, err)
	require.Equal(t, make([]byte, len(password)), password)
}

func TestBcryptCheckZerosInputPassword(t *testing.T) {
	t.Parallel()

	hasher := hash.NewBcrypt()

	hashed, err := hasher.Hash([]byte("hello"))

	require.NoError(t, err)

	password := []byte("hello")

	_, err = hasher.Check(password, hashed)

	require.NoError(t, err)
	require.Equal(t, make([]byte, len(password)), password)
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
