package argon2_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework/hash/argon2"
)

func TestItCanHashPasswords(t *testing.T) {
	h := argon2.NewHasher()
	content := []byte("hello, world")

	r, err := h.Hash(content)

	require.NoError(t, err)
	require.Greater(t, len(r), 0)
}

func TestItCanCheckHashedHashes(t *testing.T) {
	h := argon2.NewHasher()
	content := []byte("hello, world")

	r, _ := h.Hash(content)
	ok, err := h.Check(content, r)

	require.NoError(t, err)
	require.True(t, ok)
}
