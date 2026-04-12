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
	cacheMock := mock.NewCacheMock(t)
	sessionMock := mock.NewSessionMock(t)

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.abc123",
	).Return(sessionMock, nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	sess, err := driver.Get(ctx, "abc123")

	require.NoError(t, err)
	require.Equal(t, sessionMock, sess)
}

func TestCacheDriverGetReturnsErrorWhenCacheFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheMock(t)
	cacheErr := errors.New("cache failure")

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.abc123",
	).Return(nil, cacheErr).Once()

	driver := session.NewCacheDriver(cacheMock)
	_, err := driver.Get(ctx, "abc123")

	require.ErrorIs(t, err, cacheErr)
}

func TestCacheDriverGetReturnsInvalidTypeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheMock(t)

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.abc123",
	).Return("not-a-session", nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	_, err := driver.Get(ctx, "abc123")

	require.ErrorIs(t, err, session.ErrCacheDriverInvalidType)
}

func TestCacheDriverSavePersistsSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheMock(t)
	sessionMock := mock.NewSessionMock(t)
	ttl := 2 * time.Hour

	sessionMock.On("SessionID").Return("session-id-123").Once()
	cacheMock.On(
		"Put",
		tmock.Anything,
		"cosmos.sessions.session-id-123",
		sessionMock,
		ttl,
	).Return(nil).Once()

	driver := session.NewCacheDriver(cacheMock)
	err := driver.Save(ctx, sessionMock, ttl)

	require.NoError(t, err)
}

func TestCacheDriverDeleteRemovesSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheMock(t)

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
	cacheMock := mock.NewCacheMock(t)
	sessionMock := mock.NewSessionMock(t)

	cacheMock.On(
		"Get", tmock.Anything, ".abc123",
	).Return(sessionMock, nil).Once()

	driver := session.NewCacheDriverWith(
		cacheMock, session.CacheDriverOptions{},
	)
	sess, err := driver.Get(ctx, "abc123")

	require.NoError(t, err)
	require.Equal(t, sessionMock, sess)
}

func TestCacheDriverGetReturnsNotFoundError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cacheMock := mock.NewCacheMock(t)

	cacheMock.On(
		"Get", tmock.Anything, "cosmos.sessions.missing",
	).Return(nil, contract.ErrCacheKeyNotFound).Once()

	driver := session.NewCacheDriver(cacheMock)
	_, err := driver.Get(ctx, "missing")

	require.ErrorIs(t, err, contract.ErrCacheKeyNotFound)
}
