package framework

import (
	"log/slog"
	"net/http"
	"sync/atomic"
)

// ResponseWriter wraps an http.ResponseWriter to intercept
// WriteHeader and Write calls, firing registered lifecycle
// hooks before delegating to the underlying writer. It also
// tracks whether WriteHeader has been called to prevent
// duplicate status line writes. The tracking flag uses
// sync/atomic for safe concurrent access.
type ResponseWriter struct {
	http.ResponseWriter
	*Hooks
	writeHeaderCalled atomic.Bool
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
func NewResponseWriter(writer http.ResponseWriter, hooks *Hooks) WrappedResponseWriter {
	wrapped := &ResponseWriter{
		ResponseWriter: writer,
		Hooks:          hooks,
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
					slog.Error("before write header hook panicked", "error", r)
				}
			}()

			hook(writer.ResponseWriter, status)
		}()
	}

	writer.ResponseWriter.WriteHeader(status)
	writer.writeHeaderCalled.Store(true)
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
					slog.Error("before write hook panicked", "error", r)
				}
			}()

			hook(writer.ResponseWriter, content)
		}()
	}

	return writer.ResponseWriter.Write(content)
}
