package contract_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
)

func TestHooksKeyIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.HooksKey)
}

func TestHooksKeyIsDistinctType(t *testing.T) {
	t.Parallel()

	var other any = struct{}{}

	require.NotEqual(t, other, contract.HooksKey)
}

func TestNewHooksReturnsNonNil(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	require.NotNil(t, hooks)
}

func TestHooksAfterResponseRegisters(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var called bool
	hooks.AfterResponse(func(err error) { called = true })

	fns := hooks.AfterResponseFuncs()

	require.Len(t, fns, 1)

	fns[0](nil)

	require.True(t, called)
}

func TestHooksAfterResponseFuncsReturnsLIFO(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var order []int
	hooks.AfterResponse(func(err error) { order = append(order, 1) })
	hooks.AfterResponse(func(err error) { order = append(order, 2) })

	for _, fn := range hooks.AfterResponseFuncs() {
		fn(nil)
	}

	require.Equal(t, []int{2, 1}, order)
}

func TestHooksBeforeWriteRegisters(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var called bool
	hooks.BeforeWrite(func(w http.ResponseWriter, content []byte) { called = true })

	fns := hooks.BeforeWriteFuncs()

	require.Len(t, fns, 1)

	fns[0](httptest.NewRecorder(), nil)

	require.True(t, called)
}

func TestHooksBeforeWriteFuncsReturnsLIFO(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var order []int
	hooks.BeforeWrite(func(w http.ResponseWriter, content []byte) { order = append(order, 1) })
	hooks.BeforeWrite(func(w http.ResponseWriter, content []byte) { order = append(order, 2) })

	for _, fn := range hooks.BeforeWriteFuncs() {
		fn(httptest.NewRecorder(), nil)
	}

	require.Equal(t, []int{2, 1}, order)
}

func TestHooksBeforeWriteHeaderRegisters(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var called bool
	hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) { called = true })

	fns := hooks.BeforeWriteHeaderFuncs()

	require.Len(t, fns, 1)

	fns[0](httptest.NewRecorder(), 200)

	require.True(t, called)
}

func TestHooksBeforeWriteHeaderFuncsReturnsLIFO(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var order []int
	hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) { order = append(order, 1) })
	hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) { order = append(order, 2) })

	for _, fn := range hooks.BeforeWriteHeaderFuncs() {
		fn(httptest.NewRecorder(), 200)
	}

	require.Equal(t, []int{2, 1}, order)
}

func TestHooksEmptyFuncsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	require.Empty(t, hooks.AfterResponseFuncs())
	require.Empty(t, hooks.BeforeWriteFuncs())
	require.Empty(t, hooks.BeforeWriteHeaderFuncs())
}

func TestHooksConcurrentAccess(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			hooks.AfterResponse(func(err error) {})
			hooks.AfterResponseFuncs()
		}()
	}

	wg.Wait()

	require.Len(t, hooks.AfterResponseFuncs(), 100)
}
