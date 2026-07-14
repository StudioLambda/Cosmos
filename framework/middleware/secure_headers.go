package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// SecureHeadersConfig configures which security headers are
// set by the [SecureHeaders] middleware. Each field maps to a
// standard HTTP security header. Empty strings disable the
// corresponding header.
type SecureHeadersConfig struct {
	// ContentTypeOptions controls the X-Content-Type-Options
	// header, which prevents MIME-type sniffing.
	ContentTypeOptions string

	// FrameOptions controls the X-Frame-Options header, which
	// protects against clickjacking by restricting iframe embedding.
	FrameOptions string

	// ReferrerPolicy controls the Referrer-Policy header, which
	// limits referrer information sent with requests.
	ReferrerPolicy string

	// XSSProtection controls the X-XSS-Protection header.
	// The default value "0" disables the legacy XSS auditor,
	// which is the current OWASP recommendation. The auditor
	// has been removed from modern browsers and can introduce
	// XSS vulnerabilities when enabled.
	XSSProtection string

	// StrictTransportSecurity controls the Strict-Transport-Security
	// header (HSTS), which forces HTTPS for subsequent requests.
	// Leave empty to omit (e.g. during local development).
	StrictTransportSecurity string

	// ContentSecurityPolicy controls the Content-Security-Policy
	// header, which mitigates XSS and data-injection attacks.
	// Leave empty to omit since CSP is highly application-specific.
	ContentSecurityPolicy string

	// PermissionsPolicy controls the Permissions-Policy header,
	// which restricts browser feature access (camera, mic, etc.).
	// Leave empty to omit.
	PermissionsPolicy string
}

// DefaultSecureHeadersConfig holds safe default values for
// all commonly recommended security headers.
var DefaultSecureHeadersConfig = SecureHeadersConfig{
	ContentTypeOptions:      "nosniff",
	FrameOptions:            "DENY",
	ReferrerPolicy:          "strict-origin-when-cross-origin",
	XSSProtection:           "0",
	StrictTransportSecurity: "max-age=63072000; includeSubDomains",
}

// SecureHeaders returns middleware that sets standard HTTP
// security response headers using [DefaultSecureHeadersConfig].
// This protects against MIME sniffing, clickjacking, referrer
// leakage, and protocol downgrade attacks.
func SecureHeaders() framework.Middleware {
	return SecureHeadersWith(DefaultSecureHeadersConfig)
}

// SecureHeadersWith returns middleware that sets HTTP security
// response headers using the provided configuration. Headers with
// empty values are skipped.
func SecureHeadersWith(config SecureHeadersConfig) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			if config.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			}

			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			if config.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", config.XSSProtection)
			}

			if config.StrictTransportSecurity != "" {
				w.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
			}

			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			}

			return next(w, r)
		}
	}
}
