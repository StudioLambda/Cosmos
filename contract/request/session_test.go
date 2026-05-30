package request_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
)

func newTestSession(id string) *contract.Session {
	return contract.NewSessionFrom(id, time.Now(), time.Now().Add(time.Hour), map[string]any{})
}

func TestSessionReturnsTrueWhenPresent(t *testing.T) {
	t.Parallel()

	sess := newTestSession("sess-1")
	ctx := context.WithValue(
		context.Background(),
		contract.SessionKey,
		sess,
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.Session(r)

	require.True(t, ok)
	require.Equal(t, "sess-1", result.SessionID())
}

func TestSessionReturnsFalseWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result, ok := request.Session(r)

	require.False(t, ok)
	require.Nil(t, result)
}

func TestSessionKeyedReturnsTrueWhenPresent(t *testing.T) {
	t.Parallel()

	type customKey struct{}
	sess := newTestSession("sess-2")
	ctx := context.WithValue(
		context.Background(),
		customKey{},
		sess,
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.SessionKeyed(r, customKey{})

	require.True(t, ok)
	require.Equal(t, "sess-2", result.SessionID())
}

func TestSessionKeyedReturnsFalseWhenMissing(t *testing.T) {
	t.Parallel()

	type customKey struct{}
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result, ok := request.SessionKeyed(r, customKey{})

	require.False(t, ok)
	require.Nil(t, result)
}

func TestSessionKeyedReturnsFalseWhenWrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(
		context.Background(),
		contract.SessionKey,
		"not a session",
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.SessionKeyed(r, contract.SessionKey)

	require.False(t, ok)
	require.Nil(t, result)
}

func TestMustSessionReturnsSessionWhenPresent(t *testing.T) {
	t.Parallel()

	sess := newTestSession("sess-3")
	ctx := context.WithValue(
		context.Background(),
		contract.SessionKey,
		sess,
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result := request.MustSession(r)

	require.Equal(t, "sess-3", result.SessionID())
}

func TestMustSessionPanicsWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	require.Panics(t, func() {
		request.MustSession(r)
	})
}

func TestMustSessionPanicsWithErrSessionNotFound(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	defer func() {
		recovered := recover()
		require.Equal(t, request.ErrSessionNotFound, recovered)
	}()

	request.MustSession(r)
}

func TestMustSessionKeyedReturnsSessionWhenPresent(t *testing.T) {
	t.Parallel()

	type customKey struct{}
	sess := newTestSession("sess-4")
	ctx := context.WithValue(
		context.Background(),
		customKey{},
		sess,
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result := request.MustSessionKeyed(r, customKey{})

	require.Equal(t, "sess-4", result.SessionID())
}

func TestMustSessionKeyedPanicsWhenMissing(t *testing.T) {
	t.Parallel()

	type customKey struct{}
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	require.Panics(t, func() {
		request.MustSessionKeyed(r, customKey{})
	})
}

func TestMustSessionKeyedPanicsWithErrSessionNotFound(t *testing.T) {
	t.Parallel()

	type customKey struct{}
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	defer func() {
		recovered := recover()
		require.Equal(t, request.ErrSessionNotFound, recovered)
	}()

	request.MustSessionKeyed(r, customKey{})
}

func TestErrSessionNotFoundHasCorrectTitle(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"Session not found",
		request.ErrSessionNotFound.Title,
	)
}

func TestErrSessionNotFoundHasCorrectStatus(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		http.StatusInternalServerError,
		request.ErrSessionNotFound.Status,
	)
}
