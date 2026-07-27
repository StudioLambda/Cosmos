package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithDefaultsPreservesExplicitValues(t *testing.T) {
	t.Parallel()

	config := RateLimitConfig{
		Name:   "custom",
		Limit:  42,
		Window: 5 * time.Second,
	}.withDefaults()

	require.Equal(t, "custom", config.Name)
	require.Equal(t, 42, config.Limit)
	require.Equal(t, 5*time.Second, config.Window)
}

func TestWithDefaultsFillsName(t *testing.T) {
	t.Parallel()

	config := RateLimitConfig{}.withDefaults()

	require.Equal(t, DefaultRateLimitConfig.Name, config.Name)
}

func TestWithDefaultsFillsLimit(t *testing.T) {
	t.Parallel()

	config := RateLimitConfig{}.withDefaults()

	require.Equal(t, DefaultRateLimitConfig.Limit, config.Limit)
}

func TestWithDefaultsFillsWindow(t *testing.T) {
	t.Parallel()

	config := RateLimitConfig{}.withDefaults()

	require.Equal(t, DefaultRateLimitConfig.Window, config.Window)
}

func TestRateLimitWithPanicsWithoutCache(t *testing.T) {
	t.Parallel()

	require.PanicsWithValue(t, "rate limit middleware: cache must not be nil", func() {
		_ = RateLimitWith(nil, RateLimitConfig{}, nil)
	})
}
