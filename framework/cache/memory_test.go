package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/cache"

	"github.com/stretchr/testify/require"
)

func TestMemoryGetReturnsStoredValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", []byte("value"), 5*time.Minute)
	require.NoError(t, err)

	val, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, []byte("value"), val)
}

func TestMemoryGetReturnsNotFoundForMissingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	_, err := mem.Get(ctx, "missing")

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryPutOverwritesExistingValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", []byte("old"), 5*time.Minute)
	require.NoError(t, err)

	err = mem.Put(ctx, "key", []byte("new"), 5*time.Minute)
	require.NoError(t, err)

	val, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, []byte("new"), val)
}

func TestMemoryDeleteRemovesKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", []byte("value"), 5*time.Minute)
	require.NoError(t, err)

	err = mem.Delete(ctx, "key")
	require.NoError(t, err)

	_, err = mem.Get(ctx, "key")

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryDeleteNonExistentKeyIsNoOp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Delete(ctx, "nonexistent")

	require.NoError(t, err)
}

func TestMemoryHasReturnsTrueForExistingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", []byte("value"), 5*time.Minute)
	require.NoError(t, err)

	found, err := mem.Has(ctx, "key")

	require.NoError(t, err)
	require.True(t, found)
}

func TestMemoryHasReturnsFalseForMissingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	found, err := mem.Has(ctx, "missing")

	require.NoError(t, err)
	require.False(t, found)
}

func TestMemoryIncrementIncreasesValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	// go-cache Increment requires the value to be stored as int64 directly
	mem.Store().Set("counter", int64(10), 5*time.Minute)

	result, err := mem.Increment(ctx, "counter", 5)

	require.NoError(t, err)
	require.Equal(t, int64(15), result)
}

func TestMemoryIncrementReturnsErrorForMissingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	_, err := mem.Increment(ctx, "missing", 1)

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryDecrementDecreasesValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	mem.Store().Set("counter", int64(10), 5*time.Minute)

	result, err := mem.Decrement(ctx, "counter", 3)

	require.NoError(t, err)
	require.Equal(t, int64(7), result)
}

func TestMemoryDecrementReturnsErrorForMissingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	_, err := mem.Decrement(ctx, "missing", 1)

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryPutZeroTTLUsesDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", []byte("value"), 0)
	require.NoError(t, err)

	val, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, []byte("value"), val)
}

func TestCacheWrapperGetDecodesJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	c := contract.NewCache(mem)

	err := c.Put(ctx, "key", "hello", 5*time.Minute)
	require.NoError(t, err)

	result, err := c.Get[string](ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "hello", result)
}

func TestCacheWrapperPullReturnsAndRemoves(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	c := contract.NewCache(mem)

	err := c.Put(ctx, "key", "value", 5*time.Minute)
	require.NoError(t, err)

	result, err := c.Pull[string](ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "value", result)

	_, err = c.Get[string](ctx, "key")
	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestCacheWrapperForever(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	c := contract.NewCache(mem)

	err := c.Forever(ctx, "key", "permanent")
	require.NoError(t, err)

	result, err := c.Get[string](ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "permanent", result)
}

func TestCacheWrapperRememberReturnsCachedValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	c := contract.NewCache(mem)

	err := c.Put(ctx, "key", "cached", 5*time.Minute)
	require.NoError(t, err)

	called := false

	result, err := c.Remember(ctx, "key", 5*time.Minute, func() (string, error) {
		called = true
		return "computed", nil
	})

	require.NoError(t, err)
	require.Equal(t, "cached", result)
	require.False(t, called)
}

func TestCacheWrapperRememberComputesOnMiss(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	c := contract.NewCache(mem)

	result, err := c.Remember(ctx, "key", 5*time.Minute, func() (string, error) {
		return "computed", nil
	})

	require.NoError(t, err)
	require.Equal(t, "computed", result)
}

func TestCacheWrapperIncrementDelegatesToDriver(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	c := contract.NewCache(mem)

	mem.Store().Set("counter", int64(10), 5*time.Minute)

	result, err := c.Increment(ctx, "counter", 5)

	require.NoError(t, err)
	require.Equal(t, int64(15), result)
}
