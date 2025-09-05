package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

// ErrCSRFBlocked is the default error returned when a CSRF attack is detected.
// It contains a structured problem response with appropriate HTTP status and details
// that can be safely returned to clients without exposing security implementation details.
var ErrCSRFBlocked = problem.Problem{
	Title:  "Cross-Origin Request Blocked",
	Detail: "The request was rejected because its origin or fetch context did not meet security requirements.",
	Status: http.StatusForbidden,
}

// CSRF returns a middleware that protects against Cross-Site Request Forgery attacks
// using Go's built-in http.CrossOriginProtection. It creates a new CSRF protection
// instance with the specified trusted origins and uses the default ErrCSRFBlocked
// error for rejected requests.
//
// Parameters:
//   - origins: A list of trusted origin URLs that are allowed to make cross-origin requests
//
// Returns a middleware function that can be applied to routes or route groups.
func CSRF(origins ...string) framework.Middleware {
	csrf := http.NewCrossOriginProtection()

	for _, origin := range origins {
		csrf.AddTrustedOrigin(origin)
	}

	return CSRFWith(csrf, ErrCSRFBlocked)
}

// CSRFWith creates a CSRF protection middleware using a custom CrossOriginProtection
// instance and a custom error response. This provides full control over the CSRF
// configuration and error handling behavior.
//
// Parameters:
//   - csrf: A configured CrossOriginProtection instance with desired settings
//   - p: The problem.Problem to return when CSRF validation fails
//
// Returns a middleware function that validates requests and returns the custom error on failure.
func CSRFWith(csrf *http.CrossOriginProtection, p problem.Problem) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			if err := csrf.Check(r); err != nil {
				return p.WithError(err)
			}

			return next(w, r)
		}
	}
}
