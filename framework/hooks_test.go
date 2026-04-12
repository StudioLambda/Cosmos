package framework_test

import (
	"net/http"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func TestNewHooksEmpty(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()

	require.Empty(t, hooks.BeforeWriteHeaderFuncs())
	require.Empty(t, hooks.BeforeWriteFuncs())
	require.Empty(t, hooks.AfterResponseFuncs())
}

func TestBeforeWriteHeaderRegistersAndReturnsReversed(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	var order []int

	first := contract.BeforeWriteHeaderHook(
		func(w http.ResponseWriter, status int) { order = append(order, 1) },
	)

	second := contract.BeforeWriteHeaderHook(
		func(w http.ResponseWriter, status int) { order = append(order, 2) },
	)

	hooks.BeforeWriteHeader(first, second)

	funcs := hooks.BeforeWriteHeaderFuncs()

	require.Len(t, funcs, 2)

	for _, fn := range funcs {
		fn(nil, 0)
	}

	// Reversed: second fires first.
	require.Equal(t, []int{2, 1}, order)
}

func TestBeforeWriteRegistersAndReturnsReversed(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	var order []int

	first := contract.BeforeWriteHook(
		func(w http.ResponseWriter, content []byte) { order = append(order, 1) },
	)

	second := contract.BeforeWriteHook(
		func(w http.ResponseWriter, content []byte) { order = append(order, 2) },
	)

	hooks.BeforeWrite(first, second)

	funcs := hooks.BeforeWriteFuncs()

	require.Len(t, funcs, 2)

	for _, fn := range funcs {
		fn(nil, nil)
	}

	require.Equal(t, []int{2, 1}, order)
}

func TestAfterResponseRegistersAndReturnsReversed(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	var order []int

	first := contract.AfterResponseHook(
		func(err error) { order = append(order, 1) },
	)

	second := contract.AfterResponseHook(
		func(err error) { order = append(order, 2) },
	)

	hooks.AfterResponse(first, second)

	funcs := hooks.AfterResponseFuncs()

	require.Len(t, funcs, 2)

	for _, fn := range funcs {
		fn(nil)
	}

	require.Equal(t, []int{2, 1}, order)
}

func TestHooksMultipleRegistrations(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	var order []int

	hooks.AfterResponse(func(err error) { order = append(order, 1) })
	hooks.AfterResponse(func(err error) { order = append(order, 2) })
	hooks.AfterResponse(func(err error) { order = append(order, 3) })

	funcs := hooks.AfterResponseFuncs()

	require.Len(t, funcs, 3)

	for _, fn := range funcs {
		fn(nil)
	}

	// LIFO: 3, 2, 1.
	require.Equal(t, []int{3, 2, 1}, order)
}

func TestHooksFuncsReturnClone(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()

	hooks.AfterResponse(func(err error) {})

	funcs := hooks.AfterResponseFuncs()
	funcs[0] = nil

	// Original should be unaffected by mutation of the returned slice.
	require.NotNil(t, hooks.AfterResponseFuncs()[0])
}
