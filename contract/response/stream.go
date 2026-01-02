package response

import (
	"errors"
	"net/http"
)

// ErrNonFlushableWriter is returned when the http.ResponseWriter does not
// implement the http.Flusher interface, which is required for streaming responses.
var ErrNonFlushableWriter = errors.New("non-flushable response writer")

// Stream sends data from a channel to an HTTP client in real-time using HTTP streaming.
// It sets appropriate headers for streaming and flushes data as it becomes available.
// The function returns when the channel is closed (graceful shutdown) or when the
// request context is canceled. If the ResponseWriter doesn't support flushing,
// it returns ErrNonFlushableWriter.
//
// Parameters:
//   - w: The HTTP response writer
//   - r: The HTTP request (used for context cancellation)
//   - c: A receive-only channel providing byte slices to stream
func Stream(w http.ResponseWriter, r *http.Request, c <-chan []byte) error {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	f, ok := w.(http.Flusher)

	if !ok {
		return ErrNonFlushableWriter
	}

	for {
		select {
		case <-r.Context().Done():
			return r.Context().Err()
		case v, ok := <-c:
			// The channel may be closed by the sender,
			// so we must ensure we end with no error if
			// that's the case as we consider this a
			// graceful shutdown. To cancel the work with
			// an error, use the request context.
			if !ok {
				return nil
			}

			w.Write(v)
			f.Flush()
		}
	}
}

// SSE (Server-Sent Events) sends data from a channel to an HTTP client using
// the Server-Sent Events protocol. It sets the appropriate Content-Type header
// for SSE and delegates to the Stream function for the actual streaming logic.
// This is useful for implementing real-time updates to web browsers that support
// the EventSource API.
//
// Parameters:
//   - w: The HTTP response writer
//   - r: The HTTP request (used for context cancellation)
//   - c: A receive-only channel providing byte slices to stream as SSE data
func SSE(w http.ResponseWriter, r *http.Request, c <-chan []byte) error {
	w.Header().Set("Content-Type", "text/event-stream")

	return Stream(w, r, c)
}
