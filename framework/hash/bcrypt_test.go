package hash_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/hash"
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
