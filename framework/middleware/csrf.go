package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

// ErrCSRFBlocked is the default error returned when a CSRF attack is detected.
// It contains a structured problem response with appropriate HTTP status and details
// that can be safely returned to clients without exposing security implementation details.
var ErrCSRFBlocked = problem.Details{
	Title:  "Cross-Origin Request Blocked",
	Detail: "The request was rejected because its origin or fetch context did not meet security requirements.",
	Status: http.StatusForbidden,
}

// FromConfiguration populates the CSRF configuration from configuration.
func (config *CSRFConfig) FromConfiguration(configuration *contract.Configuration) {
	config.TrustedOrigins = configuration.GetOr("trusted_origins", []string(nil))
}

// CSRFConfig configures CSRF protection for trusted cross-origin requests.
type CSRFConfig struct {
	// TrustedOrigins is the list of trusted origins allowed to make
	// cross-origin requests.
	TrustedOrigins []string
}

// DefaultCSRFConfig returns the default CSRF middleware configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{}
}

// CSRF returns a middleware that protects against Cross-Site Request
// Forgery attacks using Go's built-in http.CrossOriginProtection.
func CSRF(config CSRFConfig) framework.Middleware {
	csrf := http.NewCrossOriginProtection()

	for _, origin := range config.TrustedOrigins {
		csrf.AddTrustedOrigin(origin)
	}

	return CSRFWith(config, csrf, ErrCSRFBlocked)
}

// CSRFWith creates a CSRF protection middleware using a custom
// CrossOriginProtection instance. This provides full control over the
// CSRF engine while preserving a serializable application config.
func CSRFWith(config CSRFConfig, csrf *http.CrossOriginProtection, errResponse problem.Details) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			if err := csrf.Check(r); err != nil {
				return errResponse.WithError(err)
			}

			return next(w, r)
		}
	}
}
