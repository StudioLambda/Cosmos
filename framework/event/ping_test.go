package event_test

import (
	"context"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/framework/event"

	"github.com/stretchr/testify/require"
)

var (
	_ contract.EventDriver = (*event.AMQPBroker)(nil)
	_ contract.EventDriver = (*event.MemoryBroker)(nil)
	_ contract.EventDriver = (*event.MQTTBroker)(nil)
	_ contract.EventDriver = (*event.NATSBroker)(nil)
	_ contract.EventDriver = (*event.RedisBroker)(nil)
)

func TestMemoryBrokerPingSucceeds(t *testing.T) {
	t.Parallel()

	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		require.NoError(t, broker.Close())
	})

	err := broker.Ping(context.Background())

	require.NoError(t, err)
}

func TestMemoryBrokerPingAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	broker := event.NewMemoryBroker()

	err := broker.Close()
	require.NoError(t, err)

	err = broker.Ping(context.Background())

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		require.NoError(t, broker.Close())
	})

	err := broker.Ping(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestRedisBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := event.NewRedisBroker(&event.RedisBrokerConfig{Addr: "localhost:6379"})

	t.Cleanup(func() {
		require.NoError(t, broker.Close())
	})

	err := broker.Ping(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestNATSBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var broker event.NATSBroker

	err := broker.Ping(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestMQTTBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := event.NewMQTTBrokerFrom(nil, 0)

	err := broker.Ping(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestAMQPBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var broker event.AMQPBroker

	err := broker.Ping(ctx)

	require.ErrorIs(t, err, context.Canceled)
}
