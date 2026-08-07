package request

import (
	"net/http"
	"sync/atomic"

	"github.com/studiolambda/cosmos/contract"
)

// Logger returns the request logger.
func Logger(r *http.Request) *contract.Logger {
	return r.Context().
		Value(contract.LoggerKey).(*atomic.Pointer[contract.Logger]).
		Load()
}

// SetLogger atomically replaces the request logger.
func SetLogger(r *http.Request, logger *contract.Logger) {
	r.Context().
		Value(contract.LoggerKey).(*atomic.Pointer[contract.Logger]).
		Store(logger)
}
