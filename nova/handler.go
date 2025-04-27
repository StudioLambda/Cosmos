package nova

import "net/http"

// Handler determines the function type that's used
// by the router to handle incoming HTTP requests.
type Handler func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the [http.Handler] interface by providing
// a way to handle the additional return error provided by the
// [Handler] type.
func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if there was an error handling the request / response.
	// We'll make our best effort to return a http response
	// but keep in mind that the writer could already be
	// polluted somehow due to streaming or partially sent
	// responses that failed while writting.
	if err := handler(w, r); err != nil {
		// This is a fallback error handler that will be used as
		// a last resort for errors that have not been handled already
		// with middlewares such as [middleware.ErrorHandler].
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
