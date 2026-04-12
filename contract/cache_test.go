package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestErrCacheKeyNotFoundMessage(t *testing.T) {
	t.Parallel()

	require.Equal(t, "cache key not found", contract.ErrCacheKeyNotFound.Error())
}

func TestErrCacheKeyNotFoundIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.ErrCacheKeyNotFound)
}

func TestErrCacheUnsupportedOperationMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"cache unsupported operation",
		contract.ErrCacheUnsupportedOperation.Error(),
	)
}

func TestErrCacheUnsupportedOperationIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.ErrCacheUnsupportedOperation)
}
