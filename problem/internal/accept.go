package internal

import (
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// acceptPair is a simple structure that
// holds the media and quality of an accept value.
type acceptPair struct {
	media   string
	quality float64
}

// Accept is a type designed to help working
// with header values found in the "Accept" header.
type Accept struct {
	values []acceptPair
}

// ParseAccept creates a new [Accept] based on a
// [http.Request] by parsing its headers.
func ParseAccept(request *http.Request) Accept {
	accept := Accept{
		values: make([]acceptPair, 0),
	}

	for _, header := range request.Header.Values("Accept") {
		for line := range strings.SplitSeq(header, ",") {
			media, parameters, err := mime.ParseMediaType(line)

			if err != nil {
				continue
			}

			quality := 1.0

			if param, ok := parameters["q"]; ok {
				if q, err := strconv.ParseFloat(param, 64); err == nil {
					quality = q
				}
			}

			accept.values = append(accept.values, acceptPair{
				media:   media,
				quality: quality,
			})
		}
	}

	return accept
}

// find looks for a given media in the accept header and
// returns its [acceptPair] if found.
//
// The second return value is true when is found, and false otherwise.
func (accept Accept) find(media string) (acceptPair, bool) {
	for _, pair := range accept.values {
		if media == pair.media {
			return pair, true
		}

		// Test for wildcard in media type
		if strings.Contains(media, "/*") {
			// Compare only the first part.
			if trimmed := strings.TrimSuffix(media, "/*"); strings.HasPrefix(pair.media, trimmed) {
				return pair, true
			}
		}

		// Test for wildcard in accept media type
		if strings.Contains(pair.media, "/*") {
			// Compare only the first part.
			if trimmed := strings.TrimSuffix(pair.media, "/*"); strings.HasPrefix(media, trimmed) {
				return pair, true
			}
		}
	}

	return acceptPair{}, false
}

// Accepts reports whether the given media
// is found in the accept headers.
func (accept Accept) Accepts(media string) bool {
	_, found := accept.find(media)

	return found
}

// Quality returns the quality of the given media
// found in the accept headers.
//
// Returns 0 if not found.
func (accept Accept) Quality(media string) float64 {
	if pair, found := accept.find(media); found {
		return pair.quality
	}

	return 0
}

// Order creates an ordered slice that contains the
// actual acceptance order based on the accept quality.
func (accept Accept) Order() []string {
	values := slices.Clone(accept.values)

	slices.SortFunc(values, func(a, b acceptPair) int {
		if a.quality > b.quality {
			return -1
		}

		if a.quality < b.quality {
			return 1
		}

		return 0
	})

	keys := make([]string, len(accept.values))

	for i, pair := range values {
		keys[i] = pair.media
	}

	return keys
}
