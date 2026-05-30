package event_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/framework/event"

	"github.com/stretchr/testify/require"
)

func TestMemoryBrokerPublishAndSubscribe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received string
	var wg sync.WaitGroup

	wg.Add(1)

	unsub, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()

			var msg string

			_ = json.Unmarshal(payload, &msg)

			received = msg
		},
	)

	require.NoError(t, err)
	require.NotNil(t, unsub)

	data, _ := json.Marshal("hello")

	err = broker.Publish(ctx, "user.created", data)

	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, "hello", received)
}

func TestMemoryBrokerWildcardStar(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64
	var wg sync.WaitGroup

	wg.Add(1)

	_, err := broker.Subscribe(
		ctx, "user.*.created", func(payload []byte) {
			defer wg.Done()
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "user.123.created", []byte("data"))

	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(1), atomic.LoadInt64(&received))
}

func TestMemoryBrokerPublishDoesNotDeadlockWithSubscribeInHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	done := make(chan struct{})

	_, err := broker.Subscribe(
		ctx, "test.event", func(payload []byte) {
			// This Subscribe call requires broker.mu.Lock().
			// If Publish still holds broker.mu.RLock() during
			// dispatch, this will deadlock.
			_, _ = broker.Subscribe(
				ctx, "other.event", func(payload []byte) {},
			)

			close(done)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "test.event", []byte("data"))
	require.NoError(t, err)

	select {
	case <-done:
		// Handler completed without deadlock.
	case <-time.After(3 * time.Second):
		t.Fatal("deadlock detected: handler calling Subscribe blocked for 3 seconds")
	}
}

func TestMemoryBrokerWildcardHash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64
	var wg sync.WaitGroup

	wg.Add(3)

	_, err := broker.Subscribe(
		ctx, "logs.#", func(payload []byte) {
			defer wg.Done()
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "logs", []byte("data1"))
	require.NoError(t, err)

	err = broker.Publish(ctx, "logs.error", []byte("data2"))
	require.NoError(t, err)

	err = broker.Publish(ctx, "logs.error.database", []byte("data3"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(3), atomic.LoadInt64(&received))
}

func TestMemoryBrokerExactMatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64
	var wg sync.WaitGroup

	wg.Add(1)

	_, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(1), atomic.LoadInt64(&received))
}

func TestMemoryBrokerNoMatchDoesNotDeliver(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64

	_, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "order.created", []byte("data"))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	require.Equal(t, int64(0), atomic.LoadInt64(&received))
}

func TestMemoryBrokerUnsubscribeStopsDelivery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64

	unsub, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	err = unsub()
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	require.Equal(t, int64(0), atomic.LoadInt64(&received))
}

func TestMemoryBrokerMultipleSubscribers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64
	var wg sync.WaitGroup

	wg.Add(3)

	for range 3 {
		_, err := broker.Subscribe(
			ctx,
			"user.created",
			func(payload []byte) {
				defer wg.Done()
				atomic.AddInt64(&received, 1)
			},
		)
		require.NoError(t, err)
	}

	err := broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(3), atomic.LoadInt64(&received))
}

func TestMemoryBrokerPublishAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	err := broker.Close()
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerSubscribeAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	err := broker.Close()
	require.NoError(t, err)

	_, err = broker.Subscribe(
		ctx,
		"user.created",
		func(payload []byte) {},
	)

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerHandlerPanicDoesNotCrashBroker(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var wg sync.WaitGroup

	wg.Add(1)

	_, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()
			panic("handler panic")
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()
}

func TestMemoryBrokerContextCancellationStopsPublish(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	err := broker.Publish(ctx, "user.created", []byte("data"))

	require.Error(t, err)
}

func TestMemoryBrokerPayloadUnmarshal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var receivedUser User
	var wg sync.WaitGroup

	wg.Add(1)

	_, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()

			_ = json.Unmarshal(payload, &receivedUser)
		},
	)

	require.NoError(t, err)

	userData, _ := json.Marshal(User{Name: "Alice", Age: 30})

	err = broker.Publish(
		ctx, "user.created", userData,
	)

	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, "Alice", receivedUser.Name)
	require.Equal(t, 30, receivedUser.Age)
}

func TestMemoryBrokerPublishNilPayloadSucceeds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	err := broker.Publish(ctx, "user.created", nil)

	require.NoError(t, err)
}

func TestMemoryBrokerCloseWaitsForInFlightDeliveries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	var completed atomic.Bool

	_, err := broker.Subscribe(
		ctx, "slow.event", func(payload []byte) {
			time.Sleep(100 * time.Millisecond)
			completed.Store(true)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "slow.event", []byte("data"))
	require.NoError(t, err)

	err = broker.Close()
	require.NoError(t, err)

	require.True(t, completed.Load())
}

func TestMemoryBrokerCloseClearsHandlers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	_, err := broker.Subscribe(
		ctx,
		"user.created",
		func(payload []byte) {},
	)

	require.NoError(t, err)

	err = broker.Close()
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerUnsubscribeOneDoesNotAffectOther(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	var received int64
	var wg sync.WaitGroup

	unsub1, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	wg.Add(1)

	_, err = broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()
			atomic.AddInt64(&received, 1)
		},
	)

	require.NoError(t, err)

	err = unsub1()
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(1), atomic.LoadInt64(&received))
}

func TestMemoryBrokerPublishRejectsEmptyEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	err := broker.Publish(ctx, "", []byte("data"))

	require.Error(t, err)
	require.ErrorIs(t, err, event.ErrInvalidEvent)
}

func TestMemoryBrokerPublishRejectsControlCharacters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	err := broker.Publish(ctx, "user.\tcreated", []byte("data"))

	require.Error(t, err)
	require.ErrorIs(t, err, event.ErrInvalidEvent)
}

func TestMemoryBrokerPublishRejectsTooLongEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	longEvent := strings.Repeat("a", 256)

	err := broker.Publish(ctx, longEvent, []byte("data"))

	require.Error(t, err)
	require.ErrorIs(t, err, event.ErrInvalidEvent)
}

func TestMemoryBrokerSubscribeRejectsEmptyEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	unsub, err := broker.Subscribe(
		ctx, "", func(payload []byte) {},
	)

	require.Nil(t, unsub)
	require.Error(t, err)
	require.ErrorIs(t, err, event.ErrInvalidEvent)
}

func TestMemoryBrokerValidationErrorsWrapErrInvalidEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker()

	t.Cleanup(func() {
		_ = broker.Close()
	})

	err := broker.Publish(ctx, "", []byte("data"))

	require.Error(t, err)
	require.True(t, errors.Is(err, event.ErrInvalidEvent))
}
