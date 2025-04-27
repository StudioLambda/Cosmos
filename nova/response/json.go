package response

import (
	"encoding/json"
	"net/http"
)

// JSON serializes and writes the given json value to the
// response writer.
//
// It automatically sets the content-type to `application/json`
func JSON[T any](w http.ResponseWriter, status int, data T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
