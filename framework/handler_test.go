package framework_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/contract/response"
	"github.com/studiolambda/cosmos/framework"
)

func handler(w http.ResponseWriter, r *http.Request) error {
	return response.Status(w, http.StatusOK)
}

func TestItCanRecordHandler(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), "GET", "/", nil)
	res := framework.Handler(handler).Record(req)

	if expected := http.StatusOK; res.StatusCode != expected {
		t.Errorf("expected satatus code %d but got %d", expected, res.StatusCode)
	}
}
