package framework_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func handler(w http.ResponseWriter, r *http.Request) error {
	return response.Status(w, http.StatusOK)
}

func TestHandlerRecordReturnsResponse(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	res := framework.Handler(handler).Record(req)

	require.Equal(t, http.StatusOK, res.StatusCode)
}
