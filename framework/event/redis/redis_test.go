package redis_test

import (
	"context"
	"github.com/studiolambda/cosmos/framework/event/redis"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisBrokerPingCancelledContextReturnsError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := redis.NewRedisBroker(redis.RedisBrokerConfig{Addr: "localhost:6379"})

	t.Cleanup(func() {
		require.NoError(t, broker.Close())
	})

	require.ErrorIs(t, broker.Ping(ctx), context.Canceled)
}
