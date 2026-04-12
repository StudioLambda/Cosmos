package event

import (
	"errors"
	"fmt"
)

// ErrInvalidEvent is returned when an event name is empty or contains
// invalid characters.
var ErrInvalidEvent = errors.New("invalid event name")

// validateEvent checks that an event name is non-empty, does not exceed
// 255 bytes, and contains no control characters.
func validateEvent(event string) error {
	if event == "" {
		return fmt.Errorf("%w: must not be empty", ErrInvalidEvent)
	}

	if len(event) > 255 {
		return fmt.Errorf("%w: exceeds 255 bytes", ErrInvalidEvent)
	}

	for _, r := range event {
		if r < 0x20 || r == 0x7F {
			return fmt.Errorf("%w: contains control character", ErrInvalidEvent)
		}
	}

	return nil
}
