package middleware

import (
	"net/http"

	"github.com/studiolambda/cosmos/framework"
	"github.com/studiolambda/cosmos/problem"
)

var ErrCSRFBlocked = problem.Problem{
	Title:  "Cross-Origin Request Blocked",
	Detail: "The request was rejected because its origin or fetch context did not meet security requirements.",
	Status: http.StatusForbidden,
}

// CSRF returns a middleware that uses [http.CrossOriginProtection] to protect
// routes against CSRF attacks and forwards the error upwards.
func CSRF(origins ...string) framework.Middleware {
	csrf := http.NewCrossOriginProtection()

	for _, origin := range origins {
		csrf.AddTrustedOrigin(origin)
	}

	return CSRFWith(csrf, ErrCSRFBlocked)
}

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
