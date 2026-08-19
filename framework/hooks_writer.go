package framework

import (
	"bufio"
	"net"
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

// WrappedResponseWriter is the interface returned by
// NewResponseWriter. It combines the standard http.ResponseWriter
// with a WriteHeaderCalled check so callers can determine
// whether a status code has already been sent.
type WrappedResponseWriter interface {
	http.ResponseWriter
	WriteHeaderCalled() bool
}

// NewResponseWriter creates a WrappedResponseWriter that fires the given hooks
// on write operations. Use [http.NewResponseController] for optional response
// capabilities such as flushing, connection deadlines, and full duplex mode.
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

	return wrapped
}

// WriteHeaderCalled reports whether a final response has already
// been started on this writer. Useful for middleware that
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

// WriteHeader sends the HTTP status code to the client after firing all
// registered BeforeWriteHeader hooks. Informational responses do not commit
// the response, except for 101 Switching Protocols.
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

	if status < http.StatusContinue || status >= http.StatusOK || status == http.StatusSwitchingProtocols {
		writer.writeHeaderCalled.Store(true)
	}
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

// FlushError flushes buffered data to the client. It is discovered by
// [http.ResponseController] and commits a default 200 response when needed.
func (writer *ResponseWriter) FlushError() error {
	if !writer.WriteHeaderCalled() {
		writer.WriteHeader(http.StatusOK)
	}

	return http.NewResponseController(writer.ResponseWriter).Flush()
}

// Hijack implements [http.Hijacker] for compatibility with WebSocket
// libraries that assert the interface directly. New code should use
// [http.ResponseController.Hijack].
func (writer *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	connection, readWriter, err := http.NewResponseController(writer.ResponseWriter).Hijack()
	if err == nil {
		writer.writeHeaderCalled.Store(true)
	}

	return connection, readWriter, err
}

func (writer *ResponseWriter) logHookPanic(message string, recovered any) {
	if writer.logger == nil {
		return
	}

	writer.logger.Load().Error(message, "error", recovered)
}
