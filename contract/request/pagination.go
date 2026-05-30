package request

import "net/http"

// Pagination extracts the page number and per-page count from the
// request query parameters "page" and "per_page". It applies sensible
// defaults: page 1, 25 items per page, and a maximum of 100 items
// per page. Use [PaginationWith] for custom defaults and limits.
func Pagination(r *http.Request) (page, perPage int) {
	return PaginationWith(r, 1, 25, 100)
}

// PaginationWith extracts the page number and per-page count from the
// request query parameters "page" and "per_page" using the provided
// defaults and maximum per-page limit. The page is floored at 1 and
// the per-page is clamped between 1 and maxPerPage.
func PaginationWith(r *http.Request, defaultPage, defaultPerPage, maxPerPage int) (page, perPage int) {
	page = QueryIntOr(r, "page", defaultPage)
	perPage = QueryIntOr(r, "per_page", defaultPerPage)

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 1
	}

	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage
}

// CursorPagination extracts the cursor string and per-page count from
// the request query parameters "cursor" and "per_page". It applies
// sensible defaults: 25 items per page and a maximum of 100 items per
// page. Use [CursorPaginationWith] for custom defaults and limits.
func CursorPagination(r *http.Request) (cursor string, perPage int) {
	return CursorPaginationWith(r, 25, 100)
}

// CursorPaginationWith extracts the cursor string and per-page count
// from the request query parameters "cursor" and "per_page" using the
// provided defaults and maximum per-page limit. The per-page is clamped
// between 1 and maxPerPage.
func CursorPaginationWith(r *http.Request, defaultPerPage, maxPerPage int) (cursor string, perPage int) {
	cursor = Query(r, "cursor")
	perPage = QueryIntOr(r, "per_page", defaultPerPage)

	if perPage < 1 {
		perPage = 1
	}

	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return cursor, perPage
}
