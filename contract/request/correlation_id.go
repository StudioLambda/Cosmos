package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
)

// CorrelationID retrieves the correlation ID from the request
// context. Returns an empty string if no correlation ID middleware
// was applied to the request.
func CorrelationID(r *http.Request) string {
	id, _ := r.Context().Value(contract.CorrelationIDKey).(string)

	return id
}
