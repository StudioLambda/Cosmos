package framework

import (
	"net/http"
	"sync/atomic"

	"github.com/studiolambda/cosmos/contract"
)

// ResponseWriter wraps an http.ResponseWriter to intercept
// WriteHeader and Write calls, firing registered lifecycle
// hooks before delegating to the underlying writer. It also
// tracks whether WriteHeader has been called to prevent
// duplicate status line writes. The tracking flag uses
// sync/atomic for safe concurrent access.
type ResponseWriter struct {
	http.ResponseWriter
	*contract.Hooks
	writeHeaderCalled atomic.Bool
	logger            *atomic.Pointer[contract.Logger]
}

// ResponseWriterFlusher extends ResponseWriter with the
// http.Flusher interface. It is returned by NewResponseWriter
// when the underlying writer supports flushing, preserving
// streaming capabilities through the hook layer.
type ResponseWriterFlusher struct {
	*ResponseWriter
	http.Flusher
}

// WrappedResponseWriter is the interface returned by
// NewResponseWriter. It combines the standard http.ResponseWriter
// with a WriteHeaderCalled check so callers can determine
// whether a status code has already been sent.
type WrappedResponseWriter interface {
	http.ResponseWriter
	WriteHeaderCalled() bool
}

// NewResponseWriter creates a WrappedResponseWriter that fires
// the given hooks on write operations. If the underlying writer
// implements http.Flusher, the returned value also satisfies
// http.Flusher via ResponseWriterFlusher.
func NewResponseWriter(
	writer http.ResponseWriter,
	hooks *contract.Hooks,
	logger ...*atomic.Pointer[contract.Logger],
) WrappedResponseWriter {
	wrapped := &ResponseWriter{
		ResponseWriter: writer,
		Hooks:          hooks,
	}

	if len(logger) > 0 {
		wrapped.logger = logger[0]
	}

	if flusher, ok := writer.(http.Flusher); ok {
		return &ResponseWriterFlusher{
			ResponseWriter: wrapped,
			Flusher:        flusher,
		}
	}

	return wrapped
}

// WriteHeaderCalled reports whether WriteHeader has already
// been invoked on this writer. Useful for middleware that
// needs to conditionally set a default status code.
//
// Example:
//
//	if !wrapped.WriteHeaderCalled() {
//		wrapped.WriteHeader(http.StatusNoContent)
//	}
func (writer *ResponseWriter) WriteHeaderCalled() bool {
	return writer.writeHeaderCalled.Load()
}

// WriteHeader sends the HTTP status code to the client after
// firing all registered BeforeWriteHeader hooks. Subsequent
// calls are no-ops to match http.ResponseWriter semantics.
func (writer *ResponseWriter) WriteHeader(status int) {
	if writer.WriteHeaderCalled() {
		return
	}

	for _, hook := range writer.Hooks.BeforeWriteHeaderFuncs() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					writer.logHookPanic("before write header hook panicked", r)
				}
			}()

			hook(writer.ResponseWriter, status)
		}()
	}

	writer.ResponseWriter.WriteHeader(status)
	writer.writeHeaderCalled.Store(true)
}

// Unwrap returns the underlying [http.ResponseWriter]. This enables
// [http.ResponseController] to discover interfaces on the original
// writer such as [http.Flusher] and [http.Hijacker].
func (writer *ResponseWriter) Unwrap() http.ResponseWriter {
	return writer.ResponseWriter
}

// Write sends the response body bytes to the client after
// firing all registered BeforeWrite hooks. If WriteHeader has
// not yet been called, it defaults to http.StatusOK, matching
// the standard http.ResponseWriter behaviour.
func (writer *ResponseWriter) Write(content []byte) (int, error) {
	if !writer.WriteHeaderCalled() {
		// Same behaviour as the [http.ResponseWriter]
		writer.WriteHeader(http.StatusOK)
	}

	for _, hook := range writer.Hooks.BeforeWriteFuncs() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					writer.logHookPanic("before write hook panicked", r)
				}
			}()

			hook(writer.ResponseWriter, content)
		}()
	}

	return writer.ResponseWriter.Write(content)
}

func (writer *ResponseWriter) logHookPanic(message string, recovered any) {
	if writer.logger == nil {
		return
	}

	writer.logger.Load().Error(message, "error", recovered)
}
