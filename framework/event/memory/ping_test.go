package memory_test

import (
	"context"
	"testing"

	event "github.com/studiolambda/cosmos/framework/event/memory"

	"github.com/stretchr/testify/require"
)

func TestMemoryBrokerPingSucceeds(t *testing.T) {
	t.Parallel()

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig)

	t.Cleanup(func() {
		require.NoError(t, broker.Close())
	})

	err := broker.Ping(context.Background())

	require.NoError(t, err)
}

func TestMemoryBrokerPingAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig)

	err := broker.Close()
	require.NoError(t, err)

	err = broker.Ping(context.Background())

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig)

	t.Cleanup(func() {
		require.NoError(t, broker.Close())
	})

	err := broker.Ping(ctx)

	require.ErrorIs(t, err, context.Canceled)
}
