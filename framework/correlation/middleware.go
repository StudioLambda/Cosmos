package correlation

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"
	"github.com/studiolambda/cosmos/framework"
)

// DefaultHeader is the default HTTP header used to
// propagate correlation IDs between services.
const DefaultHeader = "X-Correlation-ID"

// Generator is a function that generates a new
// correlation ID. It should return a unique string suitable
// for distributed tracing.
type Generator = func() (string, error)

// Options configures the correlation ID middleware.
type Options struct {
	// Header is the HTTP header name used to read and write
	// the correlation ID. Defaults to "X-Correlation-ID".
	Header string

	// Generate is the function used to create a new correlation
	// ID when one is not present in the request. Defaults to a
	// 16-byte random hex string (matching OpenTelemetry trace ID
	// format).
	Generate Generator
}

// Middleware returns middleware that ensures every request has
// a correlation ID for distributed tracing. It checks for an
// existing ID in the following order:
//
//  1. The W3C traceparent header (extracts the trace ID component)
//  2. The X-Correlation-ID header
//
// If a client-provided header value is present, it is accepted only
// when it matches a constrained safe format (ASCII alphanumeric plus
// '-', '_', '.' and max length 64). Otherwise, a new 16-byte random
// hex ID is generated. The correlation ID is stored in the request
// context and set on the response header.
//
// Retrieve the correlation ID downstream with [From].
//
// Example usage:
//
//	app.Use(correlation.Middleware())
func Middleware() framework.Middleware {
	return MiddlewareWith(Options{})
}

// MiddlewareWith returns correlation ID middleware with custom
// options. See [Options] for available configuration.
//
// Example usage:
//
//	app.Use(correlation.MiddlewareWith(correlation.Options{
//	    Header: "X-Request-ID",
//	}))
func MiddlewareWith(options Options) framework.Middleware {
	if options.Header == "" {
		options.Header = DefaultHeader
	}

	if options.Generate == nil {
		options.Generate = generate
	}

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			id := extractTraceID(r)

			if id == "" {
				candidate := strings.TrimSpace(r.Header.Get(options.Header))

				if isSafeCorrelationID(candidate) {
					id = candidate
				}
			}

			if id == "" {
				id = generateSafeID(options.Generate)
			}

			w.Header().Set(options.Header, id)

			ctx := context.WithValue(r.Context(), contract.CorrelationIDKey, id)

			return next(w, r.WithContext(ctx))
		}
	}
}

// From retrieves the correlation ID from the request
// context. Returns an empty string if the correlation ID middleware
// was not applied to the request.
//
// This is a convenience alias for [request.CorrelationID].
func From(r *http.Request) string {
	return request.CorrelationID(r)
}

// extractTraceID attempts to parse a W3C traceparent header and
// extract the trace ID component. The traceparent format is:
// {version}-{trace-id}-{parent-id}-{trace-flags}
//
// Returns an empty string if the header is missing or malformed.
func extractTraceID(r *http.Request) string {
	traceparent := r.Header.Get("Traceparent")

	if traceparent == "" {
		return ""
	}

	parts := strings.SplitN(traceparent, "-", 4)

	if len(parts) < 4 {
		return ""
	}

	traceID := parts[1]

	// W3C trace IDs are exactly 32 hex characters (16 bytes).
	if len(traceID) != 32 {
		return ""
	}

	// Validate that it's valid hex and not all zeros (invalid per spec).
	if traceID == "00000000000000000000000000000000" {
		return ""
	}

	_, err := hex.DecodeString(traceID)

	if err != nil {
		return ""
	}

	return traceID
}

// generate creates a new 16-byte random hex string
// (32 characters), matching the OpenTelemetry trace ID format.
func generate() (string, error) {
	buf := make([]byte, 16)

	_, err := rand.Read(buf)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

var fallbackSequence atomic.Uint64

// generateSafeID attempts to generate a correlation ID using the configured
// generator. If generation fails or produces an unsafe ID, it falls back to a
// deterministic-safe ID derived from current time and an atomic sequence.
func generateSafeID(generator Generator) string {
	if generated, err := generator(); err == nil {
		generated = strings.TrimSpace(generated)

		if isSafeCorrelationID(generated) {
			return generated
		}
	}

	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[:8], uint64(time.Now().UnixNano()))
	binary.BigEndian.PutUint64(buf[8:], fallbackSequence.Add(1))

	return hex.EncodeToString(buf)
}

// isSafeCorrelationID reports whether a client-provided correlation
// ID is acceptable for propagation and logging. It enforces a bounded
// length and a conservative ASCII character set.
func isSafeCorrelationID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}

	for i := 0; i < len(id); i++ {
		c := id[i]

		if c >= 'a' && c <= 'z' {
			continue
		}

		if c >= 'A' && c <= 'Z' {
			continue
		}

		if c >= '0' && c <= '9' {
			continue
		}

		if c == '-' || c == '_' || c == '.' {
			continue
		}

		return false
	}

	return true
}
