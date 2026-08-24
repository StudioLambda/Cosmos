package request

import (
	"net/http"

	"github.com/studiolambda/cosmos/contract"
)

// PaginationConfig configures offset-based pagination defaults.
type PaginationConfig struct {
	DefaultPage    int
	DefaultPerPage int
	MaxPerPage     int
}

// CursorPaginationConfig configures cursor-based pagination defaults.
type CursorPaginationConfig struct {
	DefaultPerPage int
	MaxPerPage     int
}

// DefaultPaginationConfig returns the default offset pagination settings.
func DefaultPaginationConfig() PaginationConfig {
	return PaginationConfig{
		DefaultPage:    1,
		DefaultPerPage: 25,
		MaxPerPage:     100,
	}
}

// DefaultCursorPaginationConfig returns the default cursor pagination settings.
func DefaultCursorPaginationConfig() CursorPaginationConfig {
	return CursorPaginationConfig{
		DefaultPerPage: 25,
		MaxPerPage:     100,
	}
}

// FromConfiguration populates the offset pagination configuration from configuration.
func (config *PaginationConfig) FromConfiguration(configuration *contract.Configuration) {
	*config = DefaultPaginationConfig()
	config.DefaultPage = configuration.GetOr("default_page", config.DefaultPage)
	config.DefaultPerPage = configuration.GetOr("default_per_page", config.DefaultPerPage)
	config.MaxPerPage = configuration.GetOr("max_per_page", config.MaxPerPage)
}

// FromConfiguration populates the cursor pagination configuration from configuration.
func (config *CursorPaginationConfig) FromConfiguration(configuration *contract.Configuration) {
	*config = DefaultCursorPaginationConfig()
	config.DefaultPerPage = configuration.GetOr("default_per_page", config.DefaultPerPage)
	config.MaxPerPage = configuration.GetOr("max_per_page", config.MaxPerPage)
}

// Pagination extracts the page number and per-page count from the
// request query parameters "page" and "per_page" using the provided
// configuration.
func Pagination(r *http.Request, config PaginationConfig) (page, perPage int) {
	return PaginationWith(r, config.DefaultPage, config.DefaultPerPage, config.MaxPerPage)
}

// PaginationWith extracts the page number and per-page count from the
// request query parameters "page" and "per_page" using the provided
// defaults and maximum per-page limit. The page is floored at 1 and
// the per-page is clamped between 1 and maxPerPage.
func PaginationWith(r *http.Request, defaultPage, defaultPerPage, maxPerPage int) (page, perPage int) {
	page = max(QueryIntOr(r, "page", defaultPage), 1)
	perPage = min(max(QueryIntOr(r, "per_page", defaultPerPage), 1), maxPerPage)

	return page, perPage
}

// CursorPagination extracts the cursor string and per-page count from
// the request query parameters "cursor" and "per_page" using the
// provided configuration.
func CursorPagination(r *http.Request, config CursorPaginationConfig) (cursor string, perPage int) {
	return CursorPaginationWith(r, config.DefaultPerPage, config.MaxPerPage)
}

// CursorPaginationWith extracts the cursor string and per-page count
// from the request query parameters "cursor" and "per_page" using the
// provided defaults and maximum per-page limit. The per-page is clamped
// between 1 and maxPerPage.
func CursorPaginationWith(r *http.Request, defaultPerPage, maxPerPage int) (cursor string, perPage int) {
	cursor = Query(r, "cursor")
	perPage = min(max(QueryIntOr(r, "per_page", defaultPerPage), 1), maxPerPage)

	return cursor, perPage
}
