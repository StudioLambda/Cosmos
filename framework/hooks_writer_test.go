package framework_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

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

func (writer *flusherWriter) Flush() {
	writer.flushed.Store(true)
}

func TestNewResponseWriterNonFlusher(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(&plainWriter{rec}, hooks)

	_, isFlusher := wrapped.(http.Flusher)

	require.False(t, isFlusher)
}

func TestNewResponseWriterFlusher(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(
		&flusherWriter{ResponseWriter: rec},
		hooks,
	)

	flusher, isFlusher := wrapped.(http.Flusher)

	require.True(t, isFlusher)

	flusher.Flush()
}

func TestWriteHeaderCalledInitiallyFalse(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	require.False(t, wrapped.WriteHeaderCalled())
}

func TestWriteHeaderSetsCalledFlag(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	wrapped.WriteHeader(http.StatusOK)

	require.True(t, wrapped.WriteHeaderCalled())
}

func TestWriteHeaderFiresBeforeWriteHeaderHooks(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
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

	hooks := framework.NewHooks()
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

	hooks := framework.NewHooks()
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

	hooks := framework.NewHooks()
	rec := httptest.NewRecorder()
	wrapped := framework.NewResponseWriter(rec, hooks)

	_, err := wrapped.Write([]byte("body"))

	require.NoError(t, err)
	require.True(t, wrapped.WriteHeaderCalled())
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestWriteAfterWriteHeaderDoesNotCallAgain(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
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

func TestWriteHeaderHookReceivesUnderlyingWriter(t *testing.T) {
	t.Parallel()

	hooks := framework.NewHooks()
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
