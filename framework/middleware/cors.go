package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/studiolambda/cosmos/framework"
)

// CORSOptions configures the Cross-Origin Resource Sharing
// (CORS) middleware behaviour. Each field maps directly to a
// CORS response header.
type CORSOptions struct {
	// AllowedOrigins is the list of origins permitted to make
	// cross-origin requests. Use "*" to allow any origin (not
	// recommended when AllowCredentials is true).
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods allowed in
	// cross-origin requests. These are returned in the
	// Access-Control-Allow-Methods header for preflight requests.
	AllowedMethods []string

	// AllowedHeaders is the list of request headers allowed in
	// cross-origin requests. These are returned in the
	// Access-Control-Allow-Headers header for preflight requests.
	AllowedHeaders []string

	// ExposedHeaders is the list of response headers that
	// browsers are allowed to access. These are returned in
	// the Access-Control-Expose-Headers header.
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include
	// credentials such as cookies, HTTP authentication, or client
	// TLS certificates. When true, the wildcard "*" must not be
	// used in AllowedOrigins.
	AllowCredentials bool

	// MaxAge is the duration in seconds that preflight results
	// may be cached by the browser. A zero value omits the
	// Access-Control-Max-Age header entirely.
	MaxAge int
}

// DefaultCORSOptions provides sensible CORS defaults that allow
// common JSON API usage from any origin without credentials. The
// defaults permit GET, POST, and HEAD with standard headers and
// a 5-minute preflight cache.
var DefaultCORSOptions = CORSOptions{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodHead,
	},
	AllowedHeaders:   []string{"Accept", "Content-Type"},
	ExposedHeaders:   nil,
	AllowCredentials: false,
	MaxAge:           300,
}

// CORS creates a Cross-Origin Resource Sharing middleware with
// the given options. It sets appropriate CORS headers on every
// response and handles preflight OPTIONS requests by responding
// with a 204 No Content after setting the required headers.
//
// For the default configuration, pass [DefaultCORSOptions].
func CORS(options CORSOptions) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(
			w http.ResponseWriter,
			r *http.Request,
		) error {
			origin := r.Header.Get("Origin")

			if origin == "" {
				return next(w, r)
			}

			if !originAllowed(options.AllowedOrigins, origin) {
				return next(w, r)
			}

			setCORSHeaders(w, options, origin)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)

				return nil
			}

			return next(w, r)
		}
	}
}

// originAllowed reports whether the given origin is permitted by
// the configured allow list. A wildcard "*" entry matches all
// origins.
func originAllowed(allowed []string, origin string) bool {
	for _, candidate := range allowed {
		if candidate == "*" || candidate == origin {
			return true
		}
	}

	return false
}

// setCORSHeaders writes the appropriate Access-Control-* headers
// to the response based on the configured options and the request
// origin.
func setCORSHeaders(
	w http.ResponseWriter,
	options CORSOptions,
	origin string,
) {
	header := w.Header()

	if len(options.AllowedOrigins) == 1 &&
		options.AllowedOrigins[0] == "*" &&
		!options.AllowCredentials {
		header.Set("Access-Control-Allow-Origin", "*")
	} else {
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Vary", "Origin")
	}

	if len(options.AllowedMethods) > 0 {
		header.Set(
			"Access-Control-Allow-Methods",
			strings.Join(options.AllowedMethods, ", "),
		)
	}

	if len(options.AllowedHeaders) > 0 {
		header.Set(
			"Access-Control-Allow-Headers",
			strings.Join(options.AllowedHeaders, ", "),
		)
	}

	if len(options.ExposedHeaders) > 0 {
		header.Set(
			"Access-Control-Expose-Headers",
			strings.Join(options.ExposedHeaders, ", "),
		)
	}

	if options.AllowCredentials {
		header.Set(
			"Access-Control-Allow-Credentials",
			"true",
		)
	}

	if options.MaxAge > 0 {
		header.Set(
			"Access-Control-Max-Age",
			strconv.Itoa(options.MaxAge),
		)
	}
}
