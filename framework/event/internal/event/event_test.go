package event

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePatternRejectsNonFinalHashWildcard(t *testing.T) {
	t.Parallel()

	err := ValidatePattern("orders.#.created")

	require.ErrorIs(t, err, ErrInvalidEvent)
}

func TestValidateNameRejectsWildcards(t *testing.T) {
	t.Parallel()

	err := ValidateName("orders.*")

	require.ErrorIs(t, err, ErrInvalidEvent)
}

func TestMatchHashMatchesZeroTrailingTokens(t *testing.T) {
	t.Parallel()

	require.True(t, Match("logs.#", "logs"))
	require.True(t, Match("logs.#", "logs.error"))
	require.True(t, Match("logs.#", "logs.error.database"))
}

func TestMatchSingleTokenWildcard(t *testing.T) {
	t.Parallel()

	require.True(t, Match("users.*.created", "users.42.created"))
	require.False(t, Match("users.*.created", "users.42.audit.created"))
}
