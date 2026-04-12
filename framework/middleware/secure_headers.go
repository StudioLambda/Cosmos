package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/framework"
)

// SecureHeadersOptions configures which security headers are
// set by the [SecureHeaders] middleware. Each field maps to a
// standard HTTP security header. Empty strings disable the
// corresponding header.
type SecureHeadersOptions struct {
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

// DefaultSecureHeadersOptions holds safe default values for
// all commonly recommended security headers.
var DefaultSecureHeadersOptions = SecureHeadersOptions{
	ContentTypeOptions:      "nosniff",
	FrameOptions:            "DENY",
	ReferrerPolicy:          "strict-origin-when-cross-origin",
	XSSProtection:           "0",
	StrictTransportSecurity: "max-age=63072000; includeSubDomains",
}

// SecureHeaders returns middleware that sets standard HTTP
// security response headers using [DefaultSecureHeadersOptions].
// This protects against MIME sniffing, clickjacking, referrer
// leakage, and protocol downgrade attacks.
func SecureHeaders() framework.Middleware {
	return SecureHeadersWith(DefaultSecureHeadersOptions)
}

// SecureHeadersWith returns middleware that sets HTTP security
// response headers using the provided options. Headers with
// empty values are skipped.
func SecureHeadersWith(opts SecureHeadersOptions) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			if opts.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", opts.ContentTypeOptions)
			}

			if opts.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", opts.FrameOptions)
			}

			if opts.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", opts.ReferrerPolicy)
			}

			if opts.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", opts.XSSProtection)
			}

			if opts.StrictTransportSecurity != "" {
				w.Header().Set("Strict-Transport-Security", opts.StrictTransportSecurity)
			}

			if opts.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", opts.ContentSecurityPolicy)
			}

			if opts.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", opts.PermissionsPolicy)
			}

			return next(w, r)
		}
	}
}
