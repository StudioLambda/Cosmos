package contract

import "net/http"

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

// Hooks defines the contract for registering and retrieving
// lifecycle callbacks during HTTP request processing. Middleware
// and handlers use these hooks to observe response events.
type Hooks interface {
	// AfterResponse registers one or more callbacks to be invoked
	// after the HTTP response has been fully written.
	AfterResponse(callbacks ...AfterResponseHook)

	// AfterResponseFuncs returns all registered after-response callbacks.
	AfterResponseFuncs() []AfterResponseHook

	// BeforeWrite registers one or more callbacks to be invoked
	// just before response body bytes are written.
	BeforeWrite(callbacks ...BeforeWriteHook)

	// BeforeWriteFuncs returns all registered before-write callbacks.
	BeforeWriteFuncs() []BeforeWriteHook

	// BeforeWriteHeader registers one or more callbacks to be invoked
	// just before the response status code is written.
	BeforeWriteHeader(callbacks ...BeforeWriteHeaderHook)

	// BeforeWriteHeaderFuncs returns all registered before-write-header callbacks.
	BeforeWriteHeaderFuncs() []BeforeWriteHeaderHook
}
