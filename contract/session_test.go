package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestSessionKeyIsNonNil(t *testing.T) {
	require.NotNil(t, contract.SessionKey)
}

func TestSessionKeyIsDistinctType(t *testing.T) {
	var other any = struct{}{}

	require.NotEqual(t, other, contract.SessionKey)
}
