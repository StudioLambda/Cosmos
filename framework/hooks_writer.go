package framework

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	*Hooks
	writeHeaderCalled bool
}

type ResponseWriterFlusher struct {
	*ResponseWriter
	http.Flusher
}

type WrappedResponseWriter interface {
	http.ResponseWriter
	WriteHeaderCalled() bool
}

func NewResponseWriter(w http.ResponseWriter, h *Hooks) WrappedResponseWriter {
	wrapped := &ResponseWriter{
		ResponseWriter: w,
		Hooks:          h,
	}

	if f, ok := w.(http.Flusher); ok {
		return &ResponseWriterFlusher{
			ResponseWriter: wrapped,
			Flusher:        f,
		}
	}

	return wrapped
}

func (w *ResponseWriter) WriteHeaderCalled() bool {
	return w.writeHeaderCalled
}

func (w *ResponseWriter) WriteHeader(status int) {
	if w.WriteHeaderCalled() {
		return
	}

	for _, hook := range w.Hooks.BeforeWriteHeaderFuncs() {
		hook(w.ResponseWriter, status)
	}

	w.ResponseWriter.WriteHeader(status)
	w.writeHeaderCalled = true
}

func (w *ResponseWriter) Write(content []byte) (int, error) {
	if !w.WriteHeaderCalled() {
		// Same behaviour as the [http.ResponseWriter]
		w.WriteHeader(http.StatusOK)
	}

	for _, hook := range w.Hooks.BeforeWriteFuncs() {
		hook(w.ResponseWriter, content)
	}

	return w.ResponseWriter.Write(content)
}
