package framework_test

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

type plainWriter struct {
	http.ResponseWriter
}

type flusherWriter struct {
	http.ResponseWriter
	flushed atomic.Bool
}

type optionalWriter struct {
	http.ResponseWriter
	flushed  atomic.Bool
	pushed   atomic.Bool
	hijacked atomic.Bool
}

func (writer *flusherWriter) Flush() {
	writer.flushed.Store(true)
}

func (writer *optionalWriter) Flush() {
	writer.flushed.Store(true)
}

func (writer *optionalWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	writer.hijacked.Store(true)
	client, server := net.Pipe()

	go client.Close()

	return server, bufio.NewReadWriter(bufio.NewReader(server), bufio.NewWriter(server)), nil
}

func (writer *optionalWriter) Push(string, *http.PushOptions) error {
	writer.pushed.Store(true)

	return nil
}

func (writer *optionalWriter) ReadFrom(reader io.Reader) (int64, error) {
	return io.Copy(writer.ResponseWriter, reader)
}

func TestNewResponseWriterNonFlusher(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(&plainWriter{rec}, hooks)

	_, isFlusher := wrapped.(http.Flusher)

	require.False(t, isFlusher)
}

func TestNewResponseWriterFlusher(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(
		&flusherWriter{ResponseWriter: rec},
		hooks,
	)

	flusher, isFlusher := wrapped.(http.Flusher)

	require.True(t, isFlusher)

	flusher.Flush()
}

func TestNewResponseWriterPreservesOptionalInterfaces(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	recorder := httptest.NewRecorder()
	writer := &optionalWriter{ResponseWriter: recorder}
	wrapped := framework.NewResponseWriter(writer, hooks)

	flusher, isFlusher := wrapped.(http.Flusher)
	hijacker, isHijacker := wrapped.(http.Hijacker)
	pusher, isPusher := wrapped.(http.Pusher)
	readerFrom, isReaderFrom := wrapped.(io.ReaderFrom)

	require.True(t, isFlusher)
	require.True(t, isHijacker)
	require.True(t, isPusher)
	require.True(t, isReaderFrom)

	flusher.Flush()
	require.NoError(t, pusher.Push("/asset", nil))
	connection, _, err := hijacker.Hijack()
	require.NoError(t, err)
	require.NoError(t, connection.Close())

	_, err = readerFrom.ReadFrom(strings.NewReader("body"))
	require.NoError(t, err)
	require.True(t, writer.flushed.Load())
	require.True(t, writer.hijacked.Load())
	require.True(t, writer.pushed.Load())
}

func TestResponseWriterReaderFromFiresWriteHooks(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	recorder := httptest.NewRecorder()
	writer := &optionalWriter{ResponseWriter: recorder}
	wrapped := framework.NewResponseWriter(writer, hooks)
	readerFrom, ok := wrapped.(io.ReaderFrom)
	require.True(t, ok)
	var content []byte

	hooks.BeforeWrite(func(_ http.ResponseWriter, value []byte) {
		content = append(content, value...)
	})

	count, err := readerFrom.ReadFrom(strings.NewReader("body"))

	require.NoError(t, err)
	require.Equal(t, int64(4), count)
	require.Equal(t, []byte("body"), content)
	require.Equal(t, "body", recorder.Body.String())
}

func TestWriteHeaderCalledInitiallyFalse(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	require.False(t, wrapped.WriteHeaderCalled())
}

func TestWriteHeaderSetsCalledFlag(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	wrapped.WriteHeader(http.StatusOK)

	require.True(t, wrapped.WriteHeaderCalled())
}

func TestWriteHeaderFiresBeforeWriteHeaderHooks(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	var capturedStatus int

	hooks.BeforeWriteHeader(
		func(w http.ResponseWriter, status int) {
			capturedStatus = status
		},
	)

	wrapped.WriteHeader(http.StatusCreated)

	require.Equal(t, http.StatusCreated, capturedStatus)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestWriteHeaderSecondCallIsNoop(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	var callCount int

	hooks.BeforeWriteHeader(
		func(w http.ResponseWriter, status int) {
			callCount++
		},
	)

	wrapped.WriteHeader(http.StatusOK)
	wrapped.WriteHeader(http.StatusNotFound)

	require.Equal(t, 1, callCount)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWriteFiresBeforeWriteHooks(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	var capturedContent []byte

	hooks.BeforeWrite(
		func(w http.ResponseWriter, content []byte) {
			capturedContent = content
		},
	)

	wrapped.WriteHeader(http.StatusOK)

	n, err := wrapped.Write([]byte("hello"))

	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, []byte("hello"), capturedContent)
}

func TestWriteAutoCallsWriteHeaderWith200(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	_, err := wrapped.Write([]byte("body"))

	require.NoError(t, err)
	require.True(t, wrapped.WriteHeaderCalled())
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWriteAfterWriteHeaderDoesNotCallAgain(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	var headerCallCount int

	hooks.BeforeWriteHeader(
		func(w http.ResponseWriter, status int) {
			headerCallCount++
		},
	)

	wrapped.WriteHeader(http.StatusCreated)

	_, err := wrapped.Write([]byte("data"))

	require.NoError(t, err)
	require.Equal(t, 1, headerCallCount)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestBeforeWriteHeaderHookPanicIsRecovered(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	hooks.BeforeWriteHeader(func(w http.ResponseWriter, status int) {
		panic("header hook panic")
	})

	require.NotPanics(t, func() {
		wrapped.WriteHeader(http.StatusOK)
	})

	require.True(t, wrapped.WriteHeaderCalled())
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestBeforeWriteHookPanicIsRecovered(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	hooks.BeforeWrite(func(w http.ResponseWriter, content []byte) {
		panic("write hook panic")
	})

	wrapped.WriteHeader(http.StatusOK)

	var n int
	var err error

	require.NotPanics(t, func() {
		n, err = wrapped.Write([]byte("hello"))
	})

	require.NoError(t, err)
	require.Equal(t, 5, n)
}

func TestResponseWriterUnwrapReturnsUnderlying(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	plain := &plainWriter{rec}
	wrapped := framework.NewResponseWriter(plain, hooks)

	type unwrapper interface {
		Unwrap() http.ResponseWriter
	}

	u, ok := wrapped.(unwrapper)

	require.True(t, ok)
	require.Equal(t, plain, u.Unwrap())
}

func TestResponseControllerFlushThroughWrappedWriter(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	fw := &flusherWriter{ResponseWriter: rec}
	wrapped := framework.NewResponseWriter(fw, hooks)

	controller := http.NewResponseController(wrapped)
	err := controller.Flush()

	require.NoError(t, err)
	require.True(t, fw.flushed.Load())
}

func TestResponseWriterFlusherUnwrapReturnsUnderlying(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	fw := &flusherWriter{ResponseWriter: rec}
	wrapped := framework.NewResponseWriter(fw, hooks)

	type unwrapper interface {
		Unwrap() http.ResponseWriter
	}

	u, ok := wrapped.(unwrapper)

	require.True(t, ok)
	require.Equal(t, fw, u.Unwrap())
}

func TestWriteHeaderHookReceivesUnderlyingWriter(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	hooks.BeforeWriteHeader(
		func(w http.ResponseWriter, status int) {
			w.Header().Set("X-Custom", "value")
		},
	)

	wrapped.WriteHeader(http.StatusOK)

	require.Equal(t, "value", rec.Header().Get("X-Custom"))
}
