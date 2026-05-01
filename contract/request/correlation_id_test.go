package request_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/request"

	"github.com/stretchr/testify/require"
)

func TestCorrelationIDReturnsValueFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), contract.CorrelationIDKey, "abc123")
	r := httptest.NewRequest("GET", "/", nil).WithContext(ctx)

	require.Equal(t, "abc123", request.CorrelationID(r))
}

func TestCorrelationIDReturnsEmptyWhenMissing(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "/", nil)

	require.Empty(t, request.CorrelationID(r))
}
