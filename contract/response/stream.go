package response

import (
	"errors"
	"net/http"
)

// ErrNonFlushableWriter is returned when the http.ResponseWriter does not
// implement the http.Flusher interface, which is required for streaming responses.
var ErrNonFlushableWriter = errors.New("non-flushable response writer")

// Stream sends data from a channel to an HTTP client using streaming.
// It flushes data as it becomes available and returns when the channel
// is closed or the request context is canceled. Returns
// [ErrNonFlushableWriter] if the writer does not support flushing.
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

			if _, err := w.Write(v); err != nil {
				return err
			}

			f.Flush()
		}
	}
}

// SSE sends data from a channel to an HTTP client using the Server-Sent
// Events protocol. It sets the appropriate Content-Type header and
// delegates to [Stream] for the actual streaming logic.
func SSE(w http.ResponseWriter, r *http.Request, c <-chan []byte) error {
	w.Header().Set("Content-Type", "text/event-stream")

	return Stream(w, r, c)
}
