package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/studiolambda/cosmos/nova"
)

var (
	// ErrRecoverUnexpectedError is the default error that's passed to
	// the recover response when the error cannot be determined from the
	// given recover()'s value.
	ErrRecoverUnexpectedError = errors.New("an unexpected error occurred")
)

func defaultRecoverHandler(value any) error {
	switch r := value.(type) {
	case error:
		return r
	case string:
		return errors.New(r)
	case fmt.Stringer:
		return errors.New(r.String())
	default:
		return errors.Join(ErrRecoverUnexpectedError, fmt.Errorf("%+v", r))
	}
}

func Recover() nova.Middleware {
	return RecoverWith(defaultRecoverHandler)
}

func RecoverWith(handler func(value any) error) nova.Middleware {
	return func(next nova.Handler) nova.Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if e := recover(); e != nil {
					err = handler(recover())
				}
			}()

			return next(w, r)
		}
	}
}
