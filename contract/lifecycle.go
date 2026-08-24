package contract

import (
	"context"
	"io"
)

// Closer is an alias for [io.Closer].
type Closer = io.Closer

// Pinger verifies that an implementation is available.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Shutdowner gracefully stops an implementation before ctx expires.
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}
