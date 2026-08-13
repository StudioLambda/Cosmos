package framework

import (
	"bufio"
	"io"
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

// ResponseWriterFlusher extends ResponseWriter with the
// http.Flusher interface. It is returned by NewResponseWriter
// when the underlying writer supports flushing, preserving
// streaming capabilities through the hook layer.
type ResponseWriterFlusher struct {
	*ResponseWriter
	http.Flusher
}

func (writer *ResponseWriterFlusher) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

type responseWriterHijacker struct{ *ResponseWriter }

func (writer *responseWriterHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

type responseWriterPusher struct{ *ResponseWriter }

func (writer *responseWriterPusher) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

type responseWriterReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterFlusherHijacker struct{ *ResponseWriter }

func (writer *responseWriterFlusherHijacker) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

type responseWriterFlusherPusher struct{ *ResponseWriter }

func (writer *responseWriterFlusherPusher) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherPusher) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

type responseWriterFlusherReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterFlusherReaderFrom) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterHijackerPusher struct{ *ResponseWriter }

func (writer *responseWriterHijackerPusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

func (writer *responseWriterHijackerPusher) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

type responseWriterHijackerReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterHijackerReaderFrom) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

func (writer *responseWriterHijackerReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterPusherReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterPusherReaderFrom) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

func (writer *responseWriterPusherReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterFlusherHijackerPusher struct{ *ResponseWriter }

func (writer *responseWriterFlusherHijackerPusher) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherHijackerPusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

func (writer *responseWriterFlusherHijackerPusher) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

type responseWriterFlusherHijackerReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterFlusherHijackerReaderFrom) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherHijackerReaderFrom) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

func (writer *responseWriterFlusherHijackerReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterFlusherPusherReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterFlusherPusherReaderFrom) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherPusherReaderFrom) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

func (writer *responseWriterFlusherPusherReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterHijackerPusherReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterHijackerPusherReaderFrom) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

func (writer *responseWriterHijackerPusherReaderFrom) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

func (writer *responseWriterHijackerPusherReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
}

type responseWriterFlusherHijackerPusherReaderFrom struct{ *ResponseWriter }

func (writer *responseWriterFlusherHijackerPusherReaderFrom) Flush() {
	flushResponseWriter(writer.ResponseWriter)
}

func (writer *responseWriterFlusherHijackerPusherReaderFrom) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.hijack()
}

func (writer *responseWriterFlusherHijackerPusherReaderFrom) Push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.push(target, options)
}

func (writer *responseWriterFlusherHijackerPusherReaderFrom) ReadFrom(reader io.Reader) (int64, error) {
	return writer.ResponseWriter.readFrom(reader)
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
// the given hooks on write operations. It preserves http.Flusher,
// http.Hijacker, http.Pusher, and io.ReaderFrom when the underlying writer
// implements them.
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

	_, flusher := writer.(http.Flusher)
	_, hijacker := writer.(http.Hijacker)
	_, pusher := writer.(http.Pusher)
	_, readerFrom := writer.(io.ReaderFrom)

	switch {
	case flusher && hijacker && pusher && readerFrom:
		return &responseWriterFlusherHijackerPusherReaderFrom{wrapped}
	case flusher && hijacker && pusher:
		return &responseWriterFlusherHijackerPusher{wrapped}
	case flusher && hijacker && readerFrom:
		return &responseWriterFlusherHijackerReaderFrom{wrapped}
	case flusher && pusher && readerFrom:
		return &responseWriterFlusherPusherReaderFrom{wrapped}
	case hijacker && pusher && readerFrom:
		return &responseWriterHijackerPusherReaderFrom{wrapped}
	case flusher && hijacker:
		return &responseWriterFlusherHijacker{wrapped}
	case flusher && pusher:
		return &responseWriterFlusherPusher{wrapped}
	case flusher && readerFrom:
		return &responseWriterFlusherReaderFrom{wrapped}
	case hijacker && pusher:
		return &responseWriterHijackerPusher{wrapped}
	case hijacker && readerFrom:
		return &responseWriterHijackerReaderFrom{wrapped}
	case pusher && readerFrom:
		return &responseWriterPusherReaderFrom{wrapped}
	case flusher:
		return &ResponseWriterFlusher{ResponseWriter: wrapped, Flusher: writer.(http.Flusher)}
	case hijacker:
		return &responseWriterHijacker{wrapped}
	case pusher:
		return &responseWriterPusher{wrapped}
	case readerFrom:
		return &responseWriterReaderFrom{wrapped}
	default:
		return wrapped
	}
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

func flushResponseWriter(writer *ResponseWriter) {
	if !writer.WriteHeaderCalled() {
		writer.WriteHeader(http.StatusOK)
	}

	writer.ResponseWriter.(http.Flusher).Flush()
}

func (writer *ResponseWriter) hijack() (net.Conn, *bufio.ReadWriter, error) {
	return writer.ResponseWriter.(http.Hijacker).Hijack()
}

func (writer *ResponseWriter) push(target string, options *http.PushOptions) error {
	return writer.ResponseWriter.(http.Pusher).Push(target, options)
}

func (writer *ResponseWriter) readFrom(reader io.Reader) (int64, error) {
	return io.Copy(writerOnly{writer}, reader)
}

type writerOnly struct {
	io.Writer
}

func (writer *ResponseWriter) logHookPanic(message string, recovered any) {
	if writer.logger == nil {
		return
	}

	writer.logger.Load().Error(message, "error", recovered)
}
