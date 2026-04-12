package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/cache"
)

func TestMemoryGetReturnsStoredValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", "value", 5*time.Minute)
	require.NoError(t, err)

	val, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "value", val)
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

	err := mem.Put(ctx, "key", "old", 5*time.Minute)
	require.NoError(t, err)

	err = mem.Put(ctx, "key", "new", 5*time.Minute)
	require.NoError(t, err)

	val, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "new", val)
}

func TestMemoryDeleteRemovesKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", "value", 5*time.Minute)
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

	err := mem.Put(ctx, "key", "value", 5*time.Minute)
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

func TestMemoryPullReturnsAndRemovesValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", "value", 5*time.Minute)
	require.NoError(t, err)

	val, err := mem.Pull(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "value", val)

	_, err = mem.Get(ctx, "key")

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryPullReturnsErrorForMissingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	_, err := mem.Pull(ctx, "missing")

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryForeverStoresPermanently(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Forever(ctx, "key", "permanent")
	require.NoError(t, err)

	val, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "permanent", val)
}

func TestMemoryIncrementIncreasesValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "counter", int64(10), 5*time.Minute)
	require.NoError(t, err)

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

	err := mem.Put(ctx, "counter", int64(10), 5*time.Minute)
	require.NoError(t, err)

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

func TestMemoryRememberReturnsCachedValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", "cached", 5*time.Minute)
	require.NoError(t, err)

	called := false

	val, err := mem.Remember(
		ctx, "key", 5*time.Minute, func() (any, error) {
			called = true
			return "computed", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "cached", val)
	require.False(t, called)
}

func TestMemoryRememberComputesOnMiss(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	val, err := mem.Remember(
		ctx, "key", 5*time.Minute, func() (any, error) {
			return "computed", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "computed", val)

	val, err = mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "computed", val)
}

func TestMemoryRememberReturnsComputeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	computeErr := errors.New("compute failed")

	_, err := mem.Remember(
		ctx, "key", 5*time.Minute, func() (any, error) {
			return nil, computeErr
		},
	)

	require.ErrorIs(t, err, computeErr)
}

func TestMemoryRememberForeverReturnsCachedValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Forever(ctx, "key", "cached")
	require.NoError(t, err)

	called := false

	val, err := mem.RememberForever(
		ctx, "key", func() (any, error) {
			called = true
			return "computed", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "cached", val)
	require.False(t, called)
}

func TestMemoryRememberForeverComputesOnMiss(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	val, err := mem.RememberForever(
		ctx, "key", func() (any, error) {
			return "computed", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "computed", val)

	val, err = mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "computed", val)
}

func TestMemoryRememberForeverReturnsComputeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)
	computeErr := errors.New("compute failed")

	_, err := mem.RememberForever(
		ctx, "key", func() (any, error) {
			return nil, computeErr
		},
	)

	require.ErrorIs(t, err, computeErr)
}

func TestMemoryPutStoresVariousTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "int", 42, 5*time.Minute)
	require.NoError(t, err)

	err = mem.Put(ctx, "bool", true, 5*time.Minute)
	require.NoError(t, err)

	err = mem.Put(
		ctx, "slice", []string{"a", "b"}, 5*time.Minute,
	)
	require.NoError(t, err)

	intVal, err := mem.Get(ctx, "int")
	require.NoError(t, err)
	require.Equal(t, 42, intVal)

	boolVal, err := mem.Get(ctx, "bool")
	require.NoError(t, err)
	require.Equal(t, true, boolVal)

	sliceVal, err := mem.Get(ctx, "slice")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, sliceVal)
}

func TestMemoryIncrementWithoutExternalMutex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "counter", int64(0), 5*time.Minute)
	require.NoError(t, err)

	result, err := mem.Increment(ctx, "counter", 10)

	require.NoError(t, err)
	require.Equal(t, int64(10), result)

	result, err = mem.Increment(ctx, "counter", 5)

	require.NoError(t, err)
	require.Equal(t, int64(15), result)

	val, err := mem.Get(ctx, "counter")

	require.NoError(t, err)
	require.Equal(t, int64(15), val)
}

func TestMemoryDecrementWithoutExternalMutex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "counter", int64(20), 5*time.Minute)
	require.NoError(t, err)

	result, err := mem.Decrement(ctx, "counter", 7)

	require.NoError(t, err)
	require.Equal(t, int64(13), result)

	result, err = mem.Decrement(ctx, "counter", 3)

	require.NoError(t, err)
	require.Equal(t, int64(10), result)

	val, err := mem.Get(ctx, "counter")

	require.NoError(t, err)
	require.Equal(t, int64(10), val)
}

func TestMemoryRememberDistinguishesKeyNotFoundFromOtherErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	// When key is missing, compute is called and result is stored.
	val, err := mem.Remember(
		ctx, "key", 5*time.Minute, func() (any, error) {
			return "computed", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "computed", val)

	// Verify the value was stored.
	stored, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "computed", stored)
}

func TestMemoryRememberForeverDistinguishesKeyNotFoundFromOtherErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	// When key is missing, compute is called and result is stored permanently.
	val, err := mem.RememberForever(
		ctx, "key", func() (any, error) {
			return "computed", nil
		},
	)

	require.NoError(t, err)
	require.Equal(t, "computed", val)

	// Verify the value was stored.
	stored, err := mem.Get(ctx, "key")

	require.NoError(t, err)
	require.Equal(t, "computed", stored)
}

func TestMemoryIncrementReturnsErrorForNonIntegerValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", "not-a-number", 5*time.Minute)
	require.NoError(t, err)

	_, err = mem.Increment(ctx, "key", 1)

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}

func TestMemoryDecrementReturnsErrorForNonIntegerValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mem := cache.NewMemory(5*time.Minute, 10*time.Minute)

	err := mem.Put(ctx, "key", "not-a-number", 5*time.Minute)
	require.NoError(t, err)

	_, err = mem.Decrement(ctx, "key", 1)

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}
