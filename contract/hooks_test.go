package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestHooksKeyIsNonNil(t *testing.T) {
	require.NotNil(t, contract.HooksKey)
}

func TestHooksKeyIsDistinctType(t *testing.T) {
	var other any = struct{}{}

	require.NotEqual(t, other, contract.HooksKey)
}
