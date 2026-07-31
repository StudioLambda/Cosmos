package framework_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/framework"
)

func TestHTTPAdaptsStandardHandler(t *testing.T) {
	t.Parallel()

	handler := framework.HTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/health", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))

	response := handler.Record(httptest.NewRequest(http.MethodGet, "/health", nil))

	require.Equal(t, http.StatusNoContent, response.StatusCode)
}
