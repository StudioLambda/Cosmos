package hash_test

import (
	"testing"

	"github.com/studiolambda/cosmos/framework/hash"

	"github.com/stretchr/testify/require"
)

func TestArgon2HashProducesOutput(t *testing.T) {
	t.Parallel()

	hasher := hash.NewArgon2()
	content := []byte("hello, world")

	hashed, err := hasher.Hash(content)

	require.NoError(t, err)
	require.Greater(t, len(hashed), 0)
}

func TestArgon2CheckMatchesCorrectPassword(t *testing.T) {
	t.Parallel()

	hasher := hash.NewArgon2()

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)

	ok, err := hasher.Check([]byte("hello, world"), hashed)

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

	hasher := hash.NewArgon2With(config)

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.Greater(t, len(hashed), 0)
}

func TestArgon2CheckWrongPasswordReturnsFalse(t *testing.T) {
	t.Parallel()

	hasher := hash.NewArgon2()

	hashed, err := hasher.Hash([]byte("correct-password"))

	require.NoError(t, err)

	ok, err := hasher.Check([]byte("wrong-password"), hashed)

	require.NoError(t, err)
	require.False(t, ok)
}

func TestArgon2HashZerosInputPassword(t *testing.T) {
	t.Parallel()

	hasher := hash.NewArgon2()
	password := []byte("sensitive-data")

	_, err := hasher.Hash(password)

	require.NoError(t, err)
	require.Equal(t, make([]byte, len(password)), password)
}

func TestArgon2CheckZerosInputPassword(t *testing.T) {
	t.Parallel()

	hasher := hash.NewArgon2()

	hashed, err := hasher.Hash([]byte("hello"))

	require.NoError(t, err)

	password := []byte("hello")

	_, err = hasher.Check(password, hashed)

	require.NoError(t, err)
	require.Equal(t, make([]byte, len(password)), password)
}

func TestArgon2NeedsRehashReturnsTrueForDifferentParams(t *testing.T) {
	t.Parallel()

	config1 := hash.Argon2Config{
		HashLength:  32,
		SaltLength:  16,
		TimeCost:    1,
		MemoryCost:  64 * 1024,
		Parallelism: 2,
		Mode:        2,
		Version:     0x13,
	}

	config2 := hash.Argon2Config{
		HashLength:  32,
		SaltLength:  16,
		TimeCost:    3,
		MemoryCost:  64 * 1024,
		Parallelism: 2,
		Mode:        2,
		Version:     0x13,
	}

	hasher1 := hash.NewArgon2With(config1)
	hasher2 := hash.NewArgon2With(config2)

	hashed, err := hasher1.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.True(t, hasher2.NeedsRehash(hashed))
}

func TestArgon2NeedsRehashReturnsFalseForSameParams(t *testing.T) {
	t.Parallel()

	config := hash.Argon2Config{
		HashLength:  32,
		SaltLength:  16,
		TimeCost:    1,
		MemoryCost:  64 * 1024,
		Parallelism: 2,
		Mode:        2,
		Version:     0x13,
	}

	hasher := hash.NewArgon2With(config)

	hashed, err := hasher.Hash([]byte("hello, world"))

	require.NoError(t, err)
	require.False(t, hasher.NeedsRehash(hashed))
}

func TestArgon2NeedsRehashReturnsTrueForInvalidHash(t *testing.T) {
	t.Parallel()

	hasher := hash.NewArgon2()

	require.True(t, hasher.NeedsRehash([]byte("not-a-valid-hash")))
}
