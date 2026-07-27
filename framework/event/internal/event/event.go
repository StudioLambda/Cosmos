// Package event provides shared event validation.
package event

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidEvent is returned when an event name is empty or contains
// invalid characters.
var ErrInvalidEvent = errors.New("invalid event name")

// Validate checks an event name or subscription pattern. Wildcards must occupy
// complete tokens; # matches one or more trailing tokens and must be final.
func Validate(event string) error {
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

	tokens := strings.Split(event, ".")
	for index, token := range tokens {
		if token == "" {
			return fmt.Errorf("%w: contains empty token", ErrInvalidEvent)
		}

		if strings.ContainsAny(token, "*#") && token != "*" && token != "#" {
			return fmt.Errorf("%w: wildcard must occupy a complete token", ErrInvalidEvent)
		}

		if token == "#" && index != len(tokens)-1 {
			return fmt.Errorf("%w: # wildcard must be final", ErrInvalidEvent)
		}
	}

	return nil
}

// Match reports whether event matches pattern using the portable event grammar.
func Match(pattern string, event string) bool {
	patternTokens := strings.Split(pattern, ".")
	eventTokens := strings.Split(event, ".")
	hasTailWildcard := patternTokens[len(patternTokens)-1] == "#"

	if hasTailWildcard {
		patternTokens = patternTokens[:len(patternTokens)-1]
		if len(eventTokens) <= len(patternTokens) {
			return false
		}
	} else if len(patternTokens) != len(eventTokens) {
		return false
	}

	for index, token := range patternTokens {
		if token != "*" && token != eventTokens[index] {
			return false
		}
	}

	return true
}
