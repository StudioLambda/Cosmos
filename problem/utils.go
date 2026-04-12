package problem

import "errors"

// stackTrace creates a stack trace of all the errors found
// that have been either Joined or Wrapped using [errors.Join]
// or [fmt.Errorf] with `%w` directive.
func stackTrace(err error) []error {
	if err == nil {
		return nil
	}

	var result []error

	type joined interface {
		Unwrap() []error
	}

	var target joined

	if errors.As(err, &target) {
		for _, err := range target.Unwrap() {
			result = append(result, stackTrace(err)...)
		}

		return result
	}

	return append(result, err)
}
