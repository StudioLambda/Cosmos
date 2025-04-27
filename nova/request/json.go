package request

import (
	"encoding/json"
	"net/http"
)

// JSON decodes the given value from the request body.
func JSON[T any](r *http.Request) (value T, err error) {
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		return value, err
	}

	return value, nil
}
