package amqp_test

import (
	"context"
	"github.com/studiolambda/cosmos/framework/event/amqp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAMQPBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var broker amqp.AMQPBroker

	require.ErrorIs(t, broker.Ping(ctx), context.Canceled)
}
