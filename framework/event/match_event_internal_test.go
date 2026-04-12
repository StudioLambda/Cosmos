package event

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchEventExactMatch(t *testing.T) {
	t.Parallel()

	result := matchEvent("user.created", "user.created")

	require.True(t, result)
}

func TestMatchEventNoMatch(t *testing.T) {
	t.Parallel()

	result := matchEvent("user.created", "order.created")

	require.False(t, result)
}

func TestMatchEventStarMatchesSingleToken(t *testing.T) {
	t.Parallel()

	result := matchEvent("user.*.created", "user.123.created")

	require.True(t, result)
}

func TestMatchEventStarDoesNotMatchMultipleTokens(t *testing.T) {
	t.Parallel()

	result := matchEvent("user.*.created", "user.1.2.created")

	require.False(t, result)
}

func TestMatchEventHashMatchesZeroTokens(t *testing.T) {
	t.Parallel()

	result := matchEvent("logs.#", "logs")

	require.True(t, result)
}

func TestMatchEventHashMatchesOneToken(t *testing.T) {
	t.Parallel()

	result := matchEvent("logs.#", "logs.error")

	require.True(t, result)
}

func TestMatchEventHashMatchesMultipleTokens(t *testing.T) {
	t.Parallel()

	result := matchEvent("logs.#", "logs.error.db")

	require.True(t, result)
}

func TestMatchEventHashAloneMatchesMultipleTokens(t *testing.T) {
	t.Parallel()

	result := matchEvent("#", "anything.here")

	require.True(t, result)
}

func TestMatchEventHashAloneMatchesSingleToken(t *testing.T) {
	t.Parallel()

	result := matchEvent("#", "single")

	require.True(t, result)
}

func TestMatchEventHashAloneMatchesEmptyString(t *testing.T) {
	t.Parallel()

	result := matchEvent("#", "")

	require.True(t, result)
}

func TestMatchEventStarAloneMatchesSingleToken(t *testing.T) {
	t.Parallel()

	result := matchEvent("*", "anything")

	require.True(t, result)
}

func TestMatchEventStarAloneDoesNotMatchDotted(t *testing.T) {
	t.Parallel()

	result := matchEvent("*", "a.b")

	require.False(t, result)
}

func TestMatchEventEmptyPatternMatchesEmptyEvent(t *testing.T) {
	t.Parallel()

	result := matchEvent("", "")

	require.True(t, result)
}

func TestMatchEventNoWildcardsDifferentValues(t *testing.T) {
	t.Parallel()

	result := matchEvent("a.b.c", "a.b.d")

	require.False(t, result)
}

func TestMatchEventPartsBothEmptySlices(t *testing.T) {
	t.Parallel()

	result := matchEventParts([]string{}, []string{})

	require.True(t, result)
}

func TestMatchEventPartsEmptyPatternNonEmptyEvent(t *testing.T) {
	t.Parallel()

	result := matchEventParts([]string{}, []string{"a"})

	require.False(t, result)
}

func TestMatchEventPartsHashPatternEmptyEvent(t *testing.T) {
	t.Parallel()

	result := matchEventParts([]string{"#"}, []string{})

	require.True(t, result)
}
