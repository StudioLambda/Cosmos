package framework

import (
	"slices"
	"sync"

	"github.com/studiolambda/cosmos/contract"
)

// Hooks provides lifecycle hook registration for the HTTP
// request/response cycle. Middleware and handlers can attach
// callbacks that fire before headers are written, before the
// body is written, and after the response completes.
//
// All methods are safe for concurrent use.
type Hooks struct {
	mutex                  sync.Mutex
	afterResponseHooks     []contract.AfterResponseHook
	beforeWriteHeaderHooks []contract.BeforeWriteHeaderHook
	beforeWriteHooks       []contract.BeforeWriteHook
}

// NewHooks creates a Hooks instance with empty callback slices
// ready to accept registrations via the Before* and After* methods.
func NewHooks() *Hooks {
	return &Hooks{
		beforeWriteHeaderHooks: []contract.BeforeWriteHeaderHook{},
		beforeWriteHooks:       []contract.BeforeWriteHook{},
	}
}

// BeforeWriteHeader registers one or more callbacks that will be
// invoked just before the response status code is written. This
// is the last opportunity to inspect or modify headers.
func (hooks *Hooks) BeforeWriteHeader(callbacks ...contract.BeforeWriteHeaderHook) {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	hooks.beforeWriteHeaderHooks = append(hooks.beforeWriteHeaderHooks, callbacks...)
}

// BeforeWriteHeaderFuncs returns a reversed clone of the registered
// BeforeWriteHeader callbacks. The reversal ensures that the most
// recently registered callback executes first (LIFO order).
func (hooks *Hooks) BeforeWriteHeaderFuncs() []contract.BeforeWriteHeaderHook {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	clone := slices.Clone(hooks.beforeWriteHeaderHooks)
	slices.Reverse(clone)

	return clone
}

// BeforeWrite registers one or more callbacks that will be
// invoked just before the response body bytes are written.
// This is useful for logging, metrics, or content transformation.
func (hooks *Hooks) BeforeWrite(callbacks ...contract.BeforeWriteHook) {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	hooks.beforeWriteHooks = append(hooks.beforeWriteHooks, callbacks...)
}

// BeforeWriteFuncs returns a reversed clone of the registered
// BeforeWrite callbacks. The reversal ensures that the most
// recently registered callback executes first (LIFO order).
func (hooks *Hooks) BeforeWriteFuncs() []contract.BeforeWriteHook {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	clone := slices.Clone(hooks.beforeWriteHooks)
	slices.Reverse(clone)

	return clone
}

// AfterResponse registers one or more callbacks that will be
// invoked after the handler has completed and all response data
// has been written. The callback receives the handler's error
// (or nil if the handler succeeded).
func (hooks *Hooks) AfterResponse(callbacks ...contract.AfterResponseHook) {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	hooks.afterResponseHooks = append(hooks.afterResponseHooks, callbacks...)
}

// AfterResponseFuncs returns a reversed clone of the registered
// AfterResponse callbacks. The reversal ensures that the most
// recently registered callback executes first (LIFO order).
func (hooks *Hooks) AfterResponseFuncs() []contract.AfterResponseHook {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	clone := slices.Clone(hooks.afterResponseHooks)
	slices.Reverse(clone)

	return clone
}
