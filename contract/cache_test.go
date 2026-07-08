package contract_test

import (
	"context"
	"encoding/json/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	cmock "github.com/studiolambda/cosmos/contract/mock"
)

func TestErrCacheKeyNotFoundMessage(t *testing.T) {
	t.Parallel()
	require.Equal(t, "cache key not found", contract.ErrCacheKeyNotFound.Error())
}

func TestErrCacheKeyNotFoundIsNonNil(t *testing.T) {
	t.Parallel()
	require.NotNil(t, contract.ErrCacheKeyNotFound)
}

func TestErrCacheUnsupportedOperationMessage(t *testing.T) {
	t.Parallel()
	require.Equal(t, "cache unsupported operation", contract.ErrCacheUnsupportedOperation.Error())
}

func TestErrCacheUnsupportedOperationIsNonNil(t *testing.T) {
	t.Parallel()
	require.NotNil(t, contract.ErrCacheUnsupportedOperation)
}

func TestCacheGet(t *testing.T) {
	t.Parallel()

	m := new(cmock.CacheDriverMock)
	c := contract.NewCache(m)

	ctx := context.Background()

	val := map[string]string{"hello": "world"}
	raw, _ := json.Marshal(val)

	m.EXPECT().Get(ctx, "mykey").Return(raw, nil)

	res, err := c.Get[map[string]string](ctx, "mykey")
	require.NoError(t, err)
	require.Equal(t, val, res)
}

func TestCachePut(t *testing.T) {
	t.Parallel()

	m := new(cmock.CacheDriverMock)
	c := contract.NewCache(m)

	ctx := context.Background()
	val := map[string]string{"foo": "bar"}
	raw, _ := json.Marshal(val)

	m.EXPECT().Put(ctx, "mykey", raw, time.Minute).Return(nil)

	err := c.Put(ctx, "mykey", val, time.Minute)
	require.NoError(t, err)
}

func TestCacheRemember(t *testing.T) {
	t.Parallel()

	m := new(cmock.CacheDriverMock)
	c := contract.NewCache(m)

	ctx := context.Background()

	// Test case where it's not found, so it computes and puts
	m.EXPECT().Get(ctx, "mykey").Return(nil, contract.ErrCacheKeyNotFound)

	val := map[string]string{"foo": "bar"}
	raw, _ := json.Marshal(val)

	m.EXPECT().Put(ctx, "mykey", raw, time.Minute).Return(nil)

	res, err := c.Remember(ctx, "mykey", time.Minute, func() (map[string]string, error) {
		return val, nil
	})

	require.NoError(t, err)
	require.Equal(t, val, res)
}

func TestCachePull(t *testing.T) {
	t.Parallel()

	m := new(cmock.CacheDriverMock)
	c := contract.NewCache(m)

	ctx := context.Background()
	val := map[string]string{"pulled": "value"}
	raw, _ := json.Marshal(val)

	m.EXPECT().Get(ctx, "mykey").Return(raw, nil)
	m.EXPECT().Delete(ctx, "mykey").Return(nil)

	res, err := c.Pull[map[string]string](ctx, "mykey")
	require.NoError(t, err)
	require.Equal(t, val, res)
}

func TestCacheForever(t *testing.T) {
	t.Parallel()

	m := new(cmock.CacheDriverMock)
	c := contract.NewCache(m)

	ctx := context.Background()
	val := "forever_value"
	raw, _ := json.Marshal(val)

	m.EXPECT().Put(ctx, "mykey", raw, time.Duration(0)).Return(nil)

	err := c.Forever(ctx, "mykey", val)
	require.NoError(t, err)
}
