package memory_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	core "github.com/studiolambda/cosmos/framework/event/internal/event"
	event "github.com/studiolambda/cosmos/framework/event/memory"

	"github.com/stretchr/testify/require"
)

type loggerDriver struct {
	errors chan string
}

func (driver loggerDriver) DebugContext(context.Context, string, ...any) {}

func (driver loggerDriver) InfoContext(context.Context, string, ...any) {}

func (driver loggerDriver) WarnContext(context.Context, string, ...any) {}

func (driver loggerDriver) ErrorContext(_ context.Context, message string, _ ...any) {
	driver.errors <- message
}

func (driver loggerDriver) With(...any) contract.LoggerDriver {
	return driver
}

func TestMemoryBrokerPublishAndSubscribe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
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

func TestMemoryBrokerLogsRecoveredHandlerPanic(t *testing.T) {
	t.Parallel()

	logs := make(chan string, 1)
	broker := event.NewMemoryBroker(event.MemoryBrokerConfig{
		Logger: contract.NewLogger(loggerDriver{errors: logs}),
	})

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	_, err := broker.Subscribe(context.Background(), "user.created", func([]byte) {
		panic("unexpected")
	})
	require.NoError(t, err)
	require.NoError(t, broker.Publish(context.Background(), "user.created", nil))

	select {
	case message := <-logs:
		require.Equal(t, "event handler panicked", message)
	case <-time.After(time.Second):
		t.Fatal("expected recovered handler panic to be logged")
	}
}

func TestMemoryBrokerWildcardStar(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64
	var wg sync.WaitGroup

	wg.Add(1)

	_, err := broker.Subscribe(
		ctx, "user.*.created", func(payload []byte) {
			defer wg.Done()
			received.Add(1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "user.123.created", []byte("data"))

	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(1), received.Load())
}

func TestMemoryBrokerPublishDoesNotDeadlockWithSubscribeInHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
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
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64
	var wg sync.WaitGroup

	wg.Add(3)

	_, err := broker.Subscribe(
		ctx, "logs.#", func(payload []byte) {
			defer wg.Done()
			received.Add(1)
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

	require.Equal(t, int64(3), received.Load())
}

func TestMemoryBrokerExactMatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64
	var wg sync.WaitGroup

	wg.Add(1)

	_, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()
			received.Add(1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(1), received.Load())
}

func TestMemoryBrokerNoMatchDoesNotDeliver(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64

	_, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			received.Add(1)
		},
	)

	require.NoError(t, err)

	err = broker.Publish(ctx, "order.created", []byte("data"))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	require.Equal(t, int64(0), received.Load())
}

func TestMemoryBrokerUnsubscribeStopsDelivery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64

	unsub, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			received.Add(1)
		},
	)

	require.NoError(t, err)

	err = unsub()
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	require.Equal(t, int64(0), received.Load())
}

func TestMemoryBrokerMultipleSubscribers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64
	var wg sync.WaitGroup

	wg.Add(3)

	for range 3 {
		_, err := broker.Subscribe(
			ctx,
			"user.created",
			func(payload []byte) {
				defer wg.Done()
				received.Add(1)
			},
		)
		require.NoError(t, err)
	}

	err := broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(3), received.Load())
}

func TestMemoryBrokerPublishAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	err := broker.Shutdown(ctx)
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerSubscribeAfterCloseReturnsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	err := broker.Shutdown(ctx)
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
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
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

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	err := broker.Publish(ctx, "user.created", []byte("data"))

	require.Error(t, err)
}

func TestMemoryBrokerPayloadUnmarshal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
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
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	err := broker.Publish(ctx, "user.created", nil)

	require.NoError(t, err)
}

func TestMemoryBrokerShutdownWaitsForInFlightDeliveries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

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

	err = broker.Shutdown(ctx)
	require.NoError(t, err)

	require.True(t, completed.Load())
}

func TestMemoryBrokerShutdownReturnsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	err := broker.Shutdown(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestMemoryBrokerShutdownReturnsDeadlineExceeded(t *testing.T) {
	t.Parallel()

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())
	release := make(chan struct{})
	started := make(chan struct{})

	_, err := broker.Subscribe(context.Background(), "slow.event", func([]byte) {
		close(started)
		<-release
	})
	require.NoError(t, err)
	require.NoError(t, broker.Publish(context.Background(), "slow.event", nil))

	<-started

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	require.ErrorIs(t, broker.Shutdown(ctx), context.DeadlineExceeded)

	close(release)
	require.NoError(t, broker.Shutdown(context.Background()))
}

func TestMemoryBrokerShutdownRejectsDeliveryAdmittedAfterShutdown(t *testing.T) {
	t.Parallel()

	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())
	started := make(chan struct{})
	release := make(chan struct{})
	completed := make(chan error, 1)

	_, err := broker.Subscribe(context.Background(), "test.event", func([]byte) {
		close(started)
		<-release
	})
	require.NoError(t, err)
	require.NoError(t, broker.Publish(context.Background(), "test.event", nil))

	<-started

	go func() {
		completed <- broker.Shutdown(context.Background())
	}()

	require.Eventually(t, func() bool {
		return broker.Publish(context.Background(), "test.event", nil) == event.ErrBrokerClosed
	}, time.Second, time.Millisecond)

	close(release)
	require.NoError(t, <-completed)
}

func TestMemoryBrokerShutdownClearsHandlers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	_, err := broker.Subscribe(
		ctx,
		"user.created",
		func(payload []byte) {},
	)

	require.NoError(t, err)

	err = broker.Shutdown(ctx)
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))

	require.ErrorIs(t, err, event.ErrBrokerClosed)
}

func TestMemoryBrokerUnsubscribeOneDoesNotAffectOther(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	var received atomic.Int64
	var wg sync.WaitGroup

	unsub1, err := broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			received.Add(1)
		},
	)

	require.NoError(t, err)

	wg.Add(1)

	_, err = broker.Subscribe(
		ctx, "user.created", func(payload []byte) {
			defer wg.Done()
			received.Add(1)
		},
	)

	require.NoError(t, err)

	err = unsub1()
	require.NoError(t, err)

	err = broker.Publish(ctx, "user.created", []byte("data"))
	require.NoError(t, err)

	wg.Wait()

	require.Equal(t, int64(1), received.Load())
}

func TestMemoryBrokerPublishRejectsEmptyEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	err := broker.Publish(ctx, "", []byte("data"))

	require.Error(t, err)
	require.ErrorIs(t, err, core.ErrInvalidEvent)
}

func TestMemoryBrokerPublishRejectsControlCharacters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	err := broker.Publish(ctx, "user.\tcreated", []byte("data"))

	require.Error(t, err)
	require.ErrorIs(t, err, core.ErrInvalidEvent)
}

func TestMemoryBrokerPublishRejectsTooLongEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	longEvent := strings.Repeat("a", 256)

	err := broker.Publish(ctx, longEvent, []byte("data"))

	require.Error(t, err)
	require.ErrorIs(t, err, core.ErrInvalidEvent)
}

func TestMemoryBrokerSubscribeRejectsEmptyEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	unsub, err := broker.Subscribe(
		ctx, "", func(payload []byte) {},
	)

	require.Nil(t, unsub)
	require.Error(t, err)
	require.ErrorIs(t, err, core.ErrInvalidEvent)
}

func TestMemoryBrokerValidationErrorsWrapErrInvalidEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	broker := event.NewMemoryBroker(event.DefaultMemoryBrokerConfig())

	t.Cleanup(func() {
		_ = broker.Shutdown(context.Background())
	})

	err := broker.Publish(ctx, "", []byte("data"))

	require.Error(t, err)
	require.True(t, errors.Is(err, core.ErrInvalidEvent))
}
