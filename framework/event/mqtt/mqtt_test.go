package mqtt_test

import (
	"context"
	"github.com/studiolambda/cosmos/framework/event/mqtt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMQTTBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := mqtt.NewMQTTBrokerFrom(nil, 0)

	require.ErrorIs(t, broker.Ping(ctx), context.Canceled)
}
