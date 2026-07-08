package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract/request"
)

func TestPaginationReturnsDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	page, perPage := request.Pagination(r)

	require.Equal(t, 1, page)
	require.Equal(t, 25, perPage)
}

func TestPaginationParsesQueryParams(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=3&per_page=50", nil)

	page, perPage := request.Pagination(r)

	require.Equal(t, 3, page)
	require.Equal(t, 50, perPage)
}

func TestPaginationClampsPerPageToMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=999", nil)

	_, perPage := request.Pagination(r)

	require.Equal(t, 100, perPage)
}

func TestPaginationClampsPageBelowOne(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=0", nil)

	page, _ := request.Pagination(r)

	require.Equal(t, 1, page)
}

func TestPaginationClampsNegativePage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=-5", nil)

	page, _ := request.Pagination(r)

	require.Equal(t, 1, page)
}

func TestPaginationClampsPerPageBelowOne(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=0", nil)

	_, perPage := request.Pagination(r)

	require.Equal(t, 1, perPage)
}

func TestPaginationWithCustomDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	page, perPage := request.PaginationWith(r, 2, 10, 50)

	require.Equal(t, 2, page)
	require.Equal(t, 10, perPage)
}

func TestPaginationWithCustomMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=100", nil)

	_, perPage := request.PaginationWith(r, 1, 10, 50)

	require.Equal(t, 50, perPage)
}

func TestPaginationIgnoresInvalidPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)

	page, _ := request.Pagination(r)

	require.Equal(t, 1, page)
}

func TestPaginationIgnoresInvalidPerPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=abc", nil)

	_, perPage := request.Pagination(r)

	require.Equal(t, 25, perPage)
}

func TestCursorPaginationReturnsDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	cursor, perPage := request.CursorPagination(r)

	require.Empty(t, cursor)
	require.Equal(t, 25, perPage)
}

func TestCursorPaginationParsesCursor(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?cursor=abc123&per_page=50", nil)

	cursor, perPage := request.CursorPagination(r)

	require.Equal(t, "abc123", cursor)
	require.Equal(t, 50, perPage)
}

func TestCursorPaginationClampsPerPageToMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=999", nil)

	_, perPage := request.CursorPagination(r)

	require.Equal(t, 100, perPage)
}

func TestCursorPaginationClampsPerPageBelowOne(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=0", nil)

	_, perPage := request.CursorPagination(r)

	require.Equal(t, 1, perPage)
}

func TestCursorPaginationWithCustomDefaults(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	_, perPage := request.CursorPaginationWith(r, 10, 50)

	require.Equal(t, 10, perPage)
}

func TestCursorPaginationWithCustomMax(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=100", nil)

	_, perPage := request.CursorPaginationWith(r, 10, 50)

	require.Equal(t, 50, perPage)
}

func TestCursorPaginationIgnoresInvalidPerPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=abc", nil)

	_, perPage := request.CursorPagination(r)

	require.Equal(t, 25, perPage)
}

func TestCursorPaginationNegativePerPage(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodGet, "/?per_page=-5", nil)

	_, perPage := request.CursorPagination(r)

	require.Equal(t, 1, perPage)
}
