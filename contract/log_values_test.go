package contract_test

import (
	"context"
	"testing"

	"github.com/studiolambda/cosmos/contract"

	"github.com/stretchr/testify/require"
)

func TestWithLogValuesComposesValues(t *testing.T) {
	t.Parallel()

	ctx := contract.WithLogValues(t.Context(), map[string]any{"request_id": "first"})
	ctx = contract.WithLogValues(ctx, map[string]any{
		"request_id":     "second",
		"correlation_id": "correlation",
	})
	values, ok := contract.LogValuesFrom(ctx)

	require.True(t, ok)
	require.Equal(t, map[string]any{
		"request_id":     "second",
		"correlation_id": "correlation",
	}, values.Values())
}

func TestLogValuesCopiesInput(t *testing.T) {
	t.Parallel()

	input := map[string]any{"request_id": "original"}
	ctx := contract.WithLogValues(context.Background(), input)
	input["request_id"] = "modified"
	values, ok := contract.LogValuesFrom(ctx)

	require.True(t, ok)
	require.Equal(t, "original", values.Values()["request_id"])
}
