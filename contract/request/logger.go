package request

import (
	"net/http"
	"sync/atomic"

	"github.com/studiolambda/cosmos/contract"
)

// Logger returns the request logger. If the request does not have a
// framework-installed logger, it returns a discard logger.
func Logger(r *http.Request) *contract.Logger {
	logger, ok := r.Context().Value(contract.LoggerKey).(*atomic.Pointer[contract.Logger])
	if !ok {
		return contract.NewLogger(nil)
	}

	return logger.Load()
}

// SetLogger atomically replaces the framework-installed request logger.
// It panics when the request does not have a framework-installed logger.
func SetLogger(r *http.Request, logger *contract.Logger) {
	r.Context().
		Value(contract.LoggerKey).(*atomic.Pointer[contract.Logger]).
		Store(logger)
}
