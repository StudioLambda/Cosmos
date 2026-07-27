package nats_test

import (
	"context"
	"github.com/studiolambda/cosmos/framework/event/nats"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNATSBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var broker nats.NATSBroker

	require.ErrorIs(t, broker.Ping(ctx), context.Canceled)
}
