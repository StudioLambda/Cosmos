package middleware

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
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

// DefaultHeader is the default HTTP header used to
// propagate correlation IDs between services.
const DefaultHeader = "X-Correlation-ID"

// Generator produces a correlation ID when request headers provide no safe ID.
// Output is trimmed and must be nonempty, at most 64 characters, and use the
// safe correlation ID character set; invalid output is replaced by a fallback.
type Generator = func() string

// CorrelationConfig configures the correlation ID middleware.
type CorrelationConfig struct {
	// Header is the HTTP header name used to read and write
	// the correlation ID. Defaults to "X-Correlation-ID".
	Header string

	// ProblemKey is the RFC 9457 problem extension member name used to include
	// the correlation ID in error responses. An empty value disables this.
	ProblemKey string
}

// DefaultCorrelationConfig returns the default correlation middleware configuration.
func DefaultCorrelationConfig() CorrelationConfig {
	return CorrelationConfig{
		Header:     DefaultHeader,
		ProblemKey: "correlation_id",
	}
}

// Correlation returns middleware that ensures every request has
// a correlation ID for distributed tracing. It checks for an
// existing ID in the following order:
//
//  1. The W3C traceparent header (extracts the trace ID component)
//  2. The configured correlation header
//
// If a client-provided header value is present, it is accepted only
// when it matches a constrained safe format (ASCII alphanumeric plus
// '-', '_', '.' and max length 64). Otherwise, a new 16-byte random
// hex ID is generated. Invalid custom-generator output uses a time-and-sequence
// fallback, so correlation ID establishment never fails. The correlation ID is stored in the request
// context and set on the response header.
//
// Retrieve the correlation ID downstream with [request.CorrelationID].
//
// Example usage:
//
//	app.Use(middleware.Correlation(middleware.DefaultCorrelationConfig()))
func Correlation(config CorrelationConfig) framework.Middleware {
	return CorrelationWith(config, nil)
}

// CorrelationWith returns correlation ID middleware with custom
// configuration and generator. See [CorrelationConfig] for available
// configuration.
//
// Example usage:
//
//	app.Use(middleware.CorrelationWith(middleware.CorrelationConfig{
//	    Header: "X-Request-ID",
//	}, customGenerator))
func CorrelationWith(config CorrelationConfig, generate Generator) framework.Middleware {
	config = config.withDefaults()

	if generate == nil {
		generate = defaultGenerator
	}

	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			id := extractTraceID(r)

			if id == "" {
				candidate := strings.TrimSpace(r.Header.Get(config.Header))

				if isSafeCorrelationID(candidate) {
					id = candidate
				}
			}

			if id == "" {
				id = generateSafeID(generate)
			}

			w.Header().Set(config.Header, id)

			ctx := context.WithValue(r.Context(), contract.CorrelationIDKey, id)
			ctx = contract.WithLogValues(ctx, map[string]any{
				"correlation_id": id,
			})

			if config.ProblemKey != "" {
				ctx = problem.WithContextValues(ctx, map[string]any{
					config.ProblemKey: id,
				})
			}

			return next(w, r.WithContext(ctx))
		}
	}
}

func (config *CorrelationConfig) FromConfiguration(configuration *contract.Configuration) {
	*config = DefaultCorrelationConfig()
	config.Header = configuration.GetOr("header", config.Header)
	config.ProblemKey = configuration.GetOr("problem_key", config.ProblemKey)
}

func (config CorrelationConfig) withDefaults() CorrelationConfig {
	if config.Header == "" {
		config.Header = DefaultCorrelationConfig().Header
	}

	return config
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

// defaultGenerator creates a new 16-byte random hex string
// (32 characters), matching the OpenTelemetry trace ID format.
func defaultGenerator() string {
	buf := make([]byte, 16)

	_, _ = rand.Read(buf)

	return hex.EncodeToString(buf)
}

var fallbackSequence atomic.Uint64

// generateSafeID attempts to generate a correlation ID using the configured
// generator. Unsafe output falls back to a deterministic-safe ID derived from
// current time and an atomic sequence. The fallback is safe for propagation but
// is not cryptographically random.
func generateSafeID(generator Generator) string {
	if generated := strings.TrimSpace(generator()); isSafeCorrelationID(generated) {
		return generated
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
