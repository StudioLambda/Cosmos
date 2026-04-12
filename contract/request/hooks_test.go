package request_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
)

// stubHooks is a minimal implementation of contract.Hooks for testing.
type stubHooks struct{}

func (stubHooks) AfterResponse(...contract.AfterResponseHook)         {}
func (stubHooks) AfterResponseFuncs() []contract.AfterResponseHook    { return nil }
func (stubHooks) BeforeWrite(...contract.BeforeWriteHook)             {}
func (stubHooks) BeforeWriteFuncs() []contract.BeforeWriteHook        { return nil }
func (stubHooks) BeforeWriteHeader(...contract.BeforeWriteHeaderHook) {}
func (stubHooks) BeforeWriteHeaderFuncs() []contract.BeforeWriteHeaderHook {
	return nil
}

func TestHooksReturnsHooksFromContext(t *testing.T) {
	t.Parallel()

	hooks := stubHooks{}
	ctx := context.WithValue(
		context.Background(), contract.HooksKey, contract.Hooks(hooks),
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result := request.Hooks(r)

	require.Equal(t, hooks, result)
}

func TestHooksPanicsWithoutContext(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	require.Panics(t, func() {
		request.Hooks(r)
	})
}

func TestHooksPanicsWithErrNoHooksMiddleware(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	defer func() {
		recovered := recover()
		require.Equal(t, request.ErrNoHooksMiddleware, recovered)
	}()

	request.Hooks(r)
}

func TestTryHooksReturnsTrueWhenPresent(t *testing.T) {
	t.Parallel()

	hooks := stubHooks{}
	ctx := context.WithValue(
		context.Background(), contract.HooksKey, contract.Hooks(hooks),
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.TryHooks(r)

	require.True(t, ok)
	require.Equal(t, hooks, result)
}

func TestTryHooksReturnsFalseWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	result, ok := request.TryHooks(r)

	require.False(t, ok)
	require.Nil(t, result)
}

func TestTryHooksReturnsFalseWhenWrongType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(
		context.Background(), contract.HooksKey, "not hooks",
	)
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)

	result, ok := request.TryHooks(r)

	require.False(t, ok)
	require.Nil(t, result)
}

func TestErrNoHooksMiddlewareHasCorrectTitle(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"No hooks context",
		request.ErrNoHooksMiddleware.Title,
	)
}

func TestErrNoHooksMiddlewareHasCorrectStatus(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		http.StatusInternalServerError,
		request.ErrNoHooksMiddleware.Status,
	)
}
