package contract

import (
	"net/http"
	"slices"
	"sync"
)

// hooksKey is the unexported type used as the context key
// for storing and retrieving [Hooks] from a request context.
type hooksKey struct{}

// HooksKey is the context key used to store and retrieve
// the [Hooks] instance from a request's context. Middleware
// should set this value so handlers can access lifecycle hooks.
var HooksKey = hooksKey{}

// AfterResponseHook is a callback invoked after the full HTTP response
// has been written. It receives the error returned by the handler,
// which may be nil if the handler succeeded.
type AfterResponseHook = func(err error)

// BeforeWriteHeaderHook is a callback invoked just before the
// response status code is written. It receives the response writer
// and the status code that is about to be sent.
type BeforeWriteHeaderHook = func(w http.ResponseWriter, status int)

// BeforeWriteHook is a callback invoked just before response body
// bytes are written. It receives the response writer and the byte
// slice that is about to be sent.
type BeforeWriteHook = func(w http.ResponseWriter, content []byte)

// Hooks provides lifecycle hook registration for the HTTP
// request/response cycle. Middleware and handlers can attach
// callbacks that fire before headers are written, before the
// body is written, and after the response completes.
//
// All methods are safe for concurrent use.
type Hooks struct {
	mutex                  sync.Mutex
	afterResponseHooks     []AfterResponseHook
	beforeWriteHeaderHooks []BeforeWriteHeaderHook
	beforeWriteHooks       []BeforeWriteHook
}

// NewHooks creates a [Hooks] instance with empty callback slices
// ready to accept registrations via the Before* and After* methods.
func NewHooks() *Hooks {
	return &Hooks{
		beforeWriteHeaderHooks: []BeforeWriteHeaderHook{},
		beforeWriteHooks:       []BeforeWriteHook{},
		afterResponseHooks:     []AfterResponseHook{},
	}
}

// AfterResponse registers one or more callbacks to be invoked
// after the HTTP response has been fully written.
func (hooks *Hooks) AfterResponse(callbacks ...AfterResponseHook) {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	hooks.afterResponseHooks = append(hooks.afterResponseHooks, callbacks...)
}

// AfterResponseFuncs returns a reversed clone of the registered
// AfterResponse callbacks. The reversal ensures that the most
// recently registered callback executes first (LIFO order).
func (hooks *Hooks) AfterResponseFuncs() []AfterResponseHook {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	clone := slices.Clone(hooks.afterResponseHooks)
	slices.Reverse(clone)

	return clone
}

// BeforeWrite registers one or more callbacks to be invoked
// just before response body bytes are written.
func (hooks *Hooks) BeforeWrite(callbacks ...BeforeWriteHook) {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	hooks.beforeWriteHooks = append(hooks.beforeWriteHooks, callbacks...)
}

// BeforeWriteFuncs returns a reversed clone of the registered
// BeforeWrite callbacks. The reversal ensures that the most
// recently registered callback executes first (LIFO order).
func (hooks *Hooks) BeforeWriteFuncs() []BeforeWriteHook {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	clone := slices.Clone(hooks.beforeWriteHooks)
	slices.Reverse(clone)

	return clone
}

// BeforeWriteHeader registers one or more callbacks to be invoked
// just before the response status code is written.
func (hooks *Hooks) BeforeWriteHeader(callbacks ...BeforeWriteHeaderHook) {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	hooks.beforeWriteHeaderHooks = append(hooks.beforeWriteHeaderHooks, callbacks...)
}

// BeforeWriteHeaderFuncs returns a reversed clone of the registered
// BeforeWriteHeader callbacks. The reversal ensures that the most
// recently registered callback executes first (LIFO order).
func (hooks *Hooks) BeforeWriteHeaderFuncs() []BeforeWriteHeaderHook {
	hooks.mutex.Lock()
	defer hooks.mutex.Unlock()

	clone := slices.Clone(hooks.beforeWriteHeaderHooks)
	slices.Reverse(clone)

	return clone
}
