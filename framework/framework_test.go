package framework_test

import (
	"testing"

	"github.com/studiolambda/cosmos/framework"

	"github.com/stretchr/testify/require"
)

func TestNewCreatesRouter(t *testing.T) {
	t.Parallel()

	router := framework.New()

	require.NotNil(t, router)
}
