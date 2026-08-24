// Package environment parses environment-variable configuration overlays.
package environment

import (
	"os"
	"strings"
)

// Values returns configuration values from variables named
// PREFIX__SECTION__KEY. Empty prefix disables environment loading.
func Values(prefix string) map[string]any {
	if prefix == "" {
		return nil
	}

	values := make(map[string]any)
	marker := strings.ToUpper(prefix) + "__"

	for _, entry := range os.Environ() {
		name, value, found := strings.Cut(entry, "=")
		if !found || !strings.HasPrefix(strings.ToUpper(name), marker) {
			continue
		}

		key := strings.TrimPrefix(name, name[:len(marker)])
		key = strings.ToLower(strings.ReplaceAll(key, "__", "."))

		if key != "" {
			values[key] = value
		}
	}

	return values
}
