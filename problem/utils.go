package problem

// stackTrace creates a stack trace of all the errors found
// that have been either Joined or Wrapped using [errors.Join]
// or [fmt.Errorf] with `%w` directive.
func stackTrace(err error) []error {
	result := make([]error, 0)

	if err == nil {
		return result
	}

	type joined interface {
		Unwrap() []error
	}

	if e, ok := err.(joined); ok {
		for _, err := range e.Unwrap() {
			result = append(result, stackTrace(err)...)
		}

		return result
	}

	return append(result, err)
}
