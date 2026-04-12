package response_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/response"
)

// flushRecorder wraps httptest.ResponseRecorder with http.Flusher support.
type flushRecorder struct {
	*httptest.ResponseRecorder
	isFlushed bool
}

func (f *flushRecorder) Flush() {
	f.isFlushed = true
	f.ResponseRecorder.Flush()
}

// nonFlushWriter is an http.ResponseWriter that does NOT
// implement http.Flusher, used to test the non-flush path.
type nonFlushWriter struct {
	header     http.Header
	statusCode int
	body       []byte
}

func newNonFlushWriter() *nonFlushWriter {
	return &nonFlushWriter{header: make(http.Header)}
}

func (w *nonFlushWriter) Header() http.Header {
	return w.header
}

func (w *nonFlushWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)

	return len(b), nil
}

func (w *nonFlushWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// errFlushWriter is an http.ResponseWriter that implements
// http.Flusher but always returns an error on Write.
type errFlushWriter struct {
	header http.Header
}

func newErrFlushWriter() *errFlushWriter {
	return &errFlushWriter{header: make(http.Header)}
}

func (w *errFlushWriter) Header() http.Header {
	return w.header
}

func (w *errFlushWriter) Write([]byte) (int, error) {
	return 0, http.ErrAbortHandler
}

func (w *errFlushWriter) WriteHeader(int) {}

func (w *errFlushWriter) Flush() {}

func TestStreamSendsDataFromChannel(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte, 2)
	ch <- []byte("chunk1")
	ch <- []byte("chunk2")
	close(ch)

	err := response.Stream(w, r, ch)

	require.NoError(t, err)
	require.Equal(t, "chunk1chunk2", w.Body.String())
	require.True(t, w.isFlushed)
}

func TestStreamSetsDefaultContentType(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)
	close(ch)

	err := response.Stream(w, r, ch)

	require.NoError(t, err)
	require.Equal(
		t,
		"application/octet-stream",
		w.Header().Get("Content-Type"),
	)
}

func TestStreamPreservesExistingContentType(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	w.Header().Set("Content-Type", "text/plain")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)
	close(ch)

	err := response.Stream(w, r, ch)

	require.NoError(t, err)
	require.Equal(t, "text/plain", w.Header().Get("Content-Type"))
}

func TestStreamSetsCacheControlHeader(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)
	close(ch)

	err := response.Stream(w, r, ch)

	require.NoError(t, err)
	require.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

func TestStreamSetsConnectionHeader(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)
	close(ch)

	err := response.Stream(w, r, ch)

	require.NoError(t, err)
	require.Equal(
		t,
		"keep-alive",
		w.Header().Get("Connection"),
	)
}

func TestStreamReturnsNilOnChannelClose(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)
	close(ch)

	err := response.Stream(w, r, ch)

	require.NoError(t, err)
}

func TestStreamReturnsErrorOnContextCancellation(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := httptest.NewRequest(
		http.MethodGet, "/", nil,
	).WithContext(ctx)
	ch := make(chan []byte)

	cancel()

	err := response.Stream(w, r, ch)

	require.ErrorIs(t, err, context.Canceled)
}

func TestStreamReturnsErrorOnWriteFailure(t *testing.T) {
	t.Parallel()

	w := newErrFlushWriter()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte, 1)
	ch <- []byte("data")

	err := response.Stream(w, r, ch)

	require.ErrorIs(t, err, http.ErrAbortHandler)
}

func TestStreamReturnsErrNonFlushableWriter(t *testing.T) {
	t.Parallel()

	w := newNonFlushWriter()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)

	err := response.Stream(w, r, ch)

	require.ErrorIs(t, err, response.ErrNonFlushableWriter)
}

func TestErrNonFlushableWriterMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"non-flushable response writer",
		response.ErrNonFlushableWriter.Error(),
	)
}

func TestSSESetsEventStreamContentType(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)
	close(ch)

	err := response.SSE(w, r, ch)

	require.NoError(t, err)
	require.Equal(
		t,
		"text/event-stream",
		w.Header().Get("Content-Type"),
	)
}

func TestSSESendsDataFromChannel(t *testing.T) {
	t.Parallel()

	w := &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte, 1)
	ch <- []byte("data: hello\n\n")
	close(ch)

	err := response.SSE(w, r, ch)

	require.NoError(t, err)
	require.Equal(t, "data: hello\n\n", w.Body.String())
}

func TestSSEReturnsErrNonFlushableWriter(t *testing.T) {
	t.Parallel()

	w := newNonFlushWriter()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	ch := make(chan []byte)

	err := response.SSE(w, r, ch)

	require.ErrorIs(t, err, response.ErrNonFlushableWriter)
}
