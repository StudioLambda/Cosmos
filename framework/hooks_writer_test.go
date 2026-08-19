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
	"time"

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

type statusWriter struct {
	http.ResponseWriter
	statuses []int
}

type optionalWriter struct {
	http.ResponseWriter
	flushed       atomic.Bool
	hijacked      atomic.Bool
	readDeadline  atomic.Bool
	writeDeadline atomic.Bool
	fullDuplex    atomic.Bool
}

func (writer *flusherWriter) Flush() {
	writer.flushed.Store(true)
}

func (writer *statusWriter) WriteHeader(status int) {
	writer.statuses = append(writer.statuses, status)
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

func (writer *optionalWriter) SetReadDeadline(time.Time) error {
	writer.readDeadline.Store(true)

	return nil
}

func (writer *optionalWriter) SetWriteDeadline(time.Time) error {
	writer.writeDeadline.Store(true)

	return nil
}

func (writer *optionalWriter) EnableFullDuplex() error {
	writer.fullDuplex.Store(true)

	return nil
}

func TestNewResponseWriterNonFlusher(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(&plainWriter{rec}, hooks)

	_, isFlusher := wrapped.(http.Flusher)

	require.False(t, isFlusher)
}

func TestNewResponseWriterDoesNotExposeFlusher(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(
		&flusherWriter{ResponseWriter: rec},
		hooks,
	)

	_, isFlusher := wrapped.(http.Flusher)

	require.False(t, isFlusher)
}

func TestNewResponseWriterDoesNotExposeOptionalInterfaces(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	recorder := httptest.NewRecorder()
	writer := &optionalWriter{ResponseWriter: recorder}
	wrapped := framework.NewResponseWriter(writer, hooks)

	_, isFlusher := wrapped.(http.Flusher)
	_, isPusher := wrapped.(http.Pusher)
	_, isReaderFrom := wrapped.(io.ReaderFrom)

	require.False(t, isFlusher)
	require.False(t, isPusher)
	require.False(t, isReaderFrom)
}

func TestResponseWriterCopyFiresWriteHooks(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	recorder := httptest.NewRecorder()
	writer := &optionalWriter{ResponseWriter: recorder}
	wrapped := framework.NewResponseWriter(writer, hooks)
	var content []byte

	hooks.BeforeWrite(func(_ http.ResponseWriter, value []byte) {
		content = append(content, value...)
	})

	count, err := io.Copy(wrapped, strings.NewReader("body"))

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
	require.True(t, wrapped.WriteHeaderCalled())
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestResponseControllerDiscoversUnderlyingCapabilities(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	recorder := httptest.NewRecorder()
	writer := &optionalWriter{ResponseWriter: recorder}
	wrapped := framework.NewResponseWriter(writer, hooks)
	controller := http.NewResponseController(wrapped)

	require.NoError(t, controller.SetReadDeadline(time.Now()))
	require.NoError(t, controller.SetWriteDeadline(time.Now()))
	require.NoError(t, controller.EnableFullDuplex())
	require.True(t, writer.readDeadline.Load())
	require.True(t, writer.writeDeadline.Load())
	require.True(t, writer.fullDuplex.Load())
}

func TestResponseWriterAllowsInformationalResponsesBeforeFinalResponse(t *testing.T) {
	t.Parallel()

	hooks := contract.NewHooks()
	writer := &statusWriter{ResponseWriter: httptest.NewRecorder()}
	wrapped := framework.NewResponseWriter(writer, hooks)

	wrapped.WriteHeader(http.StatusEarlyHints)

	require.False(t, wrapped.WriteHeaderCalled())

	wrapped.WriteHeader(http.StatusOK)

	require.True(t, wrapped.WriteHeaderCalled())
	require.Equal(t, []int{http.StatusEarlyHints, http.StatusOK}, writer.statuses)
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
