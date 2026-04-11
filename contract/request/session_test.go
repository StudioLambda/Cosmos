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

// stubSession is a minimal implementation of contract.Session for testing.
type stubSession struct {
	id string
}

func (s stubSession) SessionID() string              { return s.id }
func (s stubSession) OriginalSessionID() string      { return s.id }
func (s stubSession) Get(string) (any, bool)         { return nil, false }
func (s stubSession) Put(string, any)                {}
func (s stubSession) Delete(string)                  {}
func (s stubSession) Extend(time.Time)               {}
func (s stubSession) Regenerate() error              { return nil }
func (s stubSession) Clear()                         {}
func (s stubSession) CreatedAt() time.Time           { return time.Time{} }
func (s stubSession) ExpiresAt() time.Time           { return time.Time{} }
func (s stubSession) HasExpired() bool               { return false }
func (s stubSession) ExpiresSoon(time.Duration) bool { return false }
func (s stubSession) HasChanged() bool               { return false }
func (s stubSession) HasRegenerated() bool           { return false }
func (s stubSession) MarkAsUnchanged()               {}

func TestSessionReturnsTrueWhenPresent(t *testing.T) {
	sess := stubSession{id: "sess-1"}
	ctx := context.WithValue(
		context.Background(),
		contract.SessionKey,
		contract.Session(sess),
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.Session(r)

	require.True(t, ok)
	require.Equal(t, "sess-1", result.SessionID())
}

func TestSessionReturnsFalseWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result, ok := request.Session(r)

	require.False(t, ok)
	require.Nil(t, result)
}

func TestSessionKeyedReturnsTrueWhenPresent(t *testing.T) {
	type customKey struct{}
	sess := stubSession{id: "sess-2"}
	ctx := context.WithValue(
		context.Background(),
		customKey{},
		contract.Session(sess),
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.SessionKeyed(r, customKey{})

	require.True(t, ok)
	require.Equal(t, "sess-2", result.SessionID())
}

func TestSessionKeyedReturnsFalseWhenMissing(t *testing.T) {
	type customKey struct{}
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result, ok := request.SessionKeyed(r, customKey{})

	require.False(t, ok)
	require.Nil(t, result)
}

func TestSessionKeyedReturnsFalseWhenWrongType(t *testing.T) {
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
	sess := stubSession{id: "sess-3"}
	ctx := context.WithValue(
		context.Background(),
		contract.SessionKey,
		contract.Session(sess),
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result := request.MustSession(r)

	require.Equal(t, "sess-3", result.SessionID())
}

func TestMustSessionPanicsWhenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	require.Panics(t, func() {
		request.MustSession(r)
	})
}

func TestMustSessionPanicsWithErrSessionNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	defer func() {
		recovered := recover()
		require.Equal(t, request.ErrSessionNotFound, recovered)
	}()

	request.MustSession(r)
}

func TestMustSessionKeyedReturnsSessionWhenPresent(t *testing.T) {
	type customKey struct{}
	sess := stubSession{id: "sess-4"}
	ctx := context.WithValue(
		context.Background(),
		customKey{},
		contract.Session(sess),
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result := request.MustSessionKeyed(r, customKey{})

	require.Equal(t, "sess-4", result.SessionID())
}

func TestMustSessionKeyedPanicsWhenMissing(t *testing.T) {
	type customKey struct{}
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	require.Panics(t, func() {
		request.MustSessionKeyed(r, customKey{})
	})
}

func TestMustSessionKeyedPanicsWithErrSessionNotFound(t *testing.T) {
	type customKey struct{}
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	defer func() {
		recovered := recover()
		require.Equal(t, request.ErrSessionNotFound, recovered)
	}()

	request.MustSessionKeyed(r, customKey{})
}

func TestErrSessionNotFoundHasCorrectTitle(t *testing.T) {
	require.Equal(
		t,
		"Session not found",
		request.ErrSessionNotFound.Title,
	)
}

func TestErrSessionNotFoundHasCorrectStatus(t *testing.T) {
	require.Equal(
		t,
		http.StatusInternalServerError,
		request.ErrSessionNotFound.Status,
	)
}
