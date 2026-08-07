package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisPatternHashIncludesBaseEvent(t *testing.T) {
	t.Parallel()

	result := redisPattern("logs.#")

	require.Equal(t, "logs*", result)
}

func TestRedisPatternHashOnlyMatchesAllEvents(t *testing.T) {
	t.Parallel()

	result := redisPattern("#")

	require.Equal(t, "*", result)
}

func TestRedisPatternExactEventUnchanged(t *testing.T) {
	t.Parallel()

	result := redisPattern("logs.error")

	require.Equal(t, "logs.error", result)
}
