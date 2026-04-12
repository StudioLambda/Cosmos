package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegistryGetCreatesNewEntry(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(10, 20, time.Hour, time.Hour)
	t.Cleanup(registry.close)

	limiter := registry.get("192.168.1.1")

	require.NotNil(t, limiter)
	require.Equal(t, 1, registry.size())
}

func TestRegistryGetReusesExistingEntry(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(10, 20, time.Hour, time.Hour)
	t.Cleanup(registry.close)

	first := registry.get("192.168.1.1")
	second := registry.get("192.168.1.1")

	require.Same(t, first, second)
	require.Equal(t, 1, registry.size())
}

func TestRegistryGetUpdatesLastSeen(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(10, 20, time.Hour, time.Hour)
	t.Cleanup(registry.close)

	registry.get("192.168.1.1")

	registry.mu.Lock()
	firstSeen := registry.entries["192.168.1.1"].lastSeen
	registry.mu.Unlock()

	time.Sleep(5 * time.Millisecond)

	registry.get("192.168.1.1")

	registry.mu.Lock()
	secondSeen := registry.entries["192.168.1.1"].lastSeen
	registry.mu.Unlock()

	require.True(t, secondSeen.After(firstSeen))
}

func TestRegistrySizeReturnsEntryCount(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(10, 20, time.Hour, time.Hour)
	t.Cleanup(registry.close)

	require.Equal(t, 0, registry.size())

	registry.get("10.0.0.1")
	require.Equal(t, 1, registry.size())

	registry.get("10.0.0.2")
	require.Equal(t, 2, registry.size())

	registry.get("10.0.0.1")
	require.Equal(t, 2, registry.size())
}

func TestRegistryCleanupEvictsIdleEntries(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(
		10, 20,
		50*time.Millisecond,
		100*time.Millisecond,
	)
	t.Cleanup(registry.close)

	registry.get("10.0.0.1")
	registry.get("10.0.0.2")
	require.Equal(t, 2, registry.size())

	time.Sleep(200 * time.Millisecond)

	require.Equal(t, 0, registry.size())
}

func TestRegistryCleanupPreservesActiveEntries(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(
		10, 20,
		50*time.Millisecond,
		150*time.Millisecond,
	)
	t.Cleanup(registry.close)

	registry.get("stale")
	registry.get("active")

	time.Sleep(100 * time.Millisecond)

	registry.get("active")

	time.Sleep(100 * time.Millisecond)

	registry.mu.Lock()
	_, staleExists := registry.entries["stale"]
	_, activeExists := registry.entries["active"]
	registry.mu.Unlock()

	require.False(t, staleExists)
	require.True(t, activeExists)
}

func TestRegistryCloseStopsCleanupGoroutine(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(
		10, 20,
		50*time.Millisecond,
		50*time.Millisecond,
	)

	registry.get("10.0.0.1")
	registry.close()

	time.Sleep(150 * time.Millisecond)

	require.Equal(t, 1, registry.size())
}

func TestRegistryEvictedKeyGetsNewLimiterOnRevisit(t *testing.T) {
	t.Parallel()

	registry := newRateLimitRegistry(
		1, 1,
		50*time.Millisecond,
		100*time.Millisecond,
	)
	t.Cleanup(registry.close)

	first := registry.get("10.0.0.1")
	require.True(t, first.Allow())
	require.False(t, first.Allow())

	time.Sleep(200 * time.Millisecond)

	require.Equal(t, 0, registry.size())

	second := registry.get("10.0.0.1")
	require.True(t, second.Allow())
}

func TestWithDefaultsFillsCleanupInterval(t *testing.T) {
	t.Parallel()

	opts := RateLimitOptions{}.withDefaults()

	require.Equal(t, 1*time.Minute, opts.CleanupInterval)
}

func TestWithDefaultsFillsMaxIdleTime(t *testing.T) {
	t.Parallel()

	opts := RateLimitOptions{}.withDefaults()

	require.Equal(t, 5*time.Minute, opts.MaxIdleTime)
}

func TestWithDefaultsPreservesCustomCleanupInterval(t *testing.T) {
	t.Parallel()

	opts := RateLimitOptions{
		CleanupInterval: 30 * time.Second,
	}.withDefaults()

	require.Equal(t, 30*time.Second, opts.CleanupInterval)
}

func TestWithDefaultsPreservesCustomMaxIdleTime(t *testing.T) {
	t.Parallel()

	opts := RateLimitOptions{
		MaxIdleTime: 10 * time.Minute,
	}.withDefaults()

	require.Equal(t, 10*time.Minute, opts.MaxIdleTime)
}
