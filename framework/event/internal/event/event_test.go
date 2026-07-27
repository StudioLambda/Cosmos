package event

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRejectsNonFinalHashWildcard(t *testing.T) {
	t.Parallel()

	err := Validate("orders.#.created")

	require.ErrorIs(t, err, ErrInvalidEvent)
}

func TestMatchHashRequiresTrailingToken(t *testing.T) {
	t.Parallel()

	require.False(t, Match("logs.#", "logs"))
	require.True(t, Match("logs.#", "logs.error"))
	require.True(t, Match("logs.#", "logs.error.database"))
}

func TestMatchSingleTokenWildcard(t *testing.T) {
	t.Parallel()

	require.True(t, Match("users.*.created", "users.42.created"))
	require.False(t, Match("users.*.created", "users.42.audit.created"))
}
