package contract_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestErrDatabaseNoRowsMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"no database rows were found",
		contract.ErrDatabaseNoRows.Error(),
	)
}

func TestErrDatabaseNoRowsIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.ErrDatabaseNoRows)
}

func TestErrDatabaseNestedTransactionMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"nested transactions are not supported",
		contract.ErrDatabaseNestedTransaction.Error(),
	)
}

func TestErrDatabaseNestedTransactionIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.ErrDatabaseNestedTransaction)
}
