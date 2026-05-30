package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/mock"
	"github.com/studiolambda/cosmos/framework/session"

	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCacheDriverGetReturnsSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)

	sess := contract.NewSessionFrom("abc123", time.Now(), time.Now().Add(time.Hour), map[string]any{"user": "test"})

	// The CacheDriver stores JSON; simulate what Save would have stored.
	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.abc123",
	).Return([]byte(`{"id":"abc123","created_at":"2025-01-01T00:00:00Z","expires_at":"2025-01-01T01:00:00Z","storage":{"user":"test"}}`), nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	result, err := driver.Get(ctx, "abc123")

	require.NoError(t, err)
	require.Equal(t, "abc123", result.SessionID())

	val, ok := result.Get("user")
	require.True(t, ok)
	require.Equal(t, "test", val)

	_ = sess // reference to avoid unused
}

func TestCacheDriverGetReturnsErrorWhenCacheFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)
	cacheErr := errors.New("cache failure")

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.abc123",
	).Return([]byte(nil), cacheErr).Once()

	driver := session.NewCacheDriver(cacheMock)
	_, err := driver.Get(ctx, "abc123")

	require.ErrorIs(t, err, cacheErr)
}

func TestCacheDriverGetReturnsErrorForInvalidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.abc123",
	).Return([]byte("not-json"), nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	_, err := driver.Get(ctx, "abc123")

	require.Error(t, err)
}

func TestCacheDriverSavePersistsSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)
	ttl := 2 * time.Hour

	sess := contract.NewSessionFrom("session-id-123", time.Now(), time.Now().Add(ttl), map[string]any{"key": "val"})

	cacheMock.On(
		"Put",
		tmock.Anything,
		"cosmos.sessions.session-id-123",
		tmock.AnythingOfType("[]uint8"),
		ttl,
	).Return(nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	err := driver.Save(ctx, sess, ttl)

	require.NoError(t, err)
}

func TestCacheDriverDeleteRemovesSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)

	cacheMock.On(
		"Delete", tmock.Anything, "cosmos.sessions.abc123",
	).Return(nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	err := driver.Delete(ctx, "abc123")

	require.NoError(t, err)
}

func TestCacheDriverWithUsesEmptyPrefixWhenDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)

	cacheMock.On(
		"Get", tmock.Anything, ".abc123",
	).Return([]byte(`{"id":"abc123","created_at":"2025-01-01T00:00:00Z","expires_at":"2025-01-01T01:00:00Z","storage":{}}`), nil).Once()

	driver := session.NewCacheDriverWith(
		cacheMock, session.CacheDriverOptions{},
	)
	result, err := driver.Get(ctx, "abc123")

	require.NoError(t, err)
	require.Equal(t, "abc123", result.SessionID())
}

func TestCacheDriverGetReturnsNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheDriverMock(t)

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.missing",
	).Return([]byte(nil), contract.ErrCacheKeyNotFound).Once()

	driver := session.NewCacheDriver(cacheMock)
	_, err := driver.Get(ctx, "missing")

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}
