package response

import (
	"errors"
	"net/http"
)

var (
	ErrNonFlushableWriter = errors.New("non-flushable response writer")
)

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
			// that's the case as nova considers this a
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

func SSE(w http.ResponseWriter, r *http.Request, c <-chan []byte) error {
	w.Header().Set("Content-Type", "text/event-stream")

	return Stream(w, r, c)
}
