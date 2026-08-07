package memory_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	job "github.com/studiolambda/cosmos/framework/job/memory"

	"github.com/stretchr/testify/require"
)

func TestMemoryDriverDispatchDeliversClonedMessageWithMetadata(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	payload := []byte(`{"id":"first"}`)
	received := make(chan contract.JobMessage, 1)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(ctx, "emails", func(ctx context.Context, delivery contract.JobDelivery) error {
			received <- delivery.Message()
			return delivery.Acknowledge(ctx)
		})
	}()

	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Name: "email", Queue: "emails", Payload: payload}))
	payload[7] = 'X'
	message := <-received
	require.NotEmpty(t, message.ID)
	require.Equal(t, 1, message.Attempts)
	require.Equal(t, []byte(`{"id":"first"}`), message.Payload)
	cancel()
	require.ErrorIs(t, <-consumer, context.Canceled)
}

func TestMemoryDriverConsumersCompeteForEachMessage(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	var delivered atomic.Int64
	var consumers sync.WaitGroup
	for range 2 {
		consumers.Add(1)
		go func() {
			defer consumers.Done()
			err := driver.Consume(ctx, "work", func(ctx context.Context, delivery contract.JobDelivery) error {
				delivered.Add(1)
				return delivery.Acknowledge(ctx)
			})
			require.ErrorIs(t, err, context.Canceled)
		}()
	}

	for range 40 {
		require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	}
	require.Eventually(t, func() bool { return delivered.Load() == 40 }, time.Second, time.Millisecond)
	cancel()
	consumers.Wait()
}

func TestMemoryDriverIsolatesQueues(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	received := make(chan string, 1)
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(ctx, "second", func(ctx context.Context, delivery contract.JobDelivery) error {
			received <- delivery.Message().Queue
			return delivery.Acknowledge(ctx)
		})
	}()
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "first"}))
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "second"}))
	require.Equal(t, "second", <-received)
	cancel()
	require.ErrorIs(t, <-consumer, context.Canceled)
}

func TestMemoryDriverRetryDelaysRedelivery(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	started := time.Now()
	received := make(chan contract.JobMessage, 2)
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(ctx, "work", func(ctx context.Context, delivery contract.JobDelivery) error {
			message := delivery.Message()
			received <- message
			if message.Attempts == 1 {
				return delivery.Retry(ctx, 40*time.Millisecond)
			}

			return delivery.Acknowledge(ctx)
		})
	}()
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	require.Equal(t, 1, (<-received).Attempts)
	require.Equal(t, 2, (<-received).Attempts)
	require.GreaterOrEqual(t, time.Since(started), 35*time.Millisecond)
	cancel()
	require.ErrorIs(t, <-consumer, context.Canceled)
}

func TestMemoryDriverIntegratesWithWorkerMaxAttempts(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	var handled atomic.Int64
	worker := contract.NewJobWorker(driver, resolver(func(string) (contract.Job, error) {
		return &failingJob{handled: &handled}, nil
	}))
	worked := make(chan error, 1)
	go func() { worked <- worker.Work(ctx, "work") }()
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Name: "failing", Queue: "work", Payload: []byte(`{"value":"x"}`)}))
	require.Eventually(t, func() bool { return handled.Load() == 2 }, time.Second, time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int64(2), handled.Load())
	cancel()
	require.ErrorIs(t, <-worked, context.Canceled)
}

func TestMemoryDriverSettlementIsExactlyOnce(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	settlement := make(chan error, 2)
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(ctx, "work", func(ctx context.Context, delivery contract.JobDelivery) error {
			var calls sync.WaitGroup
			for range 2 {
				calls.Add(1)
				go func() {
					defer calls.Done()
					settlement <- delivery.Acknowledge(ctx)
				}()
			}
			calls.Wait()
			return nil
		})
	}()
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	first := <-settlement
	second := <-settlement
	require.True(t, (first == nil && errors.Is(second, job.ErrDeliverySettled)) || (second == nil && errors.Is(first, job.ErrDeliverySettled)))
	cancel()
	require.ErrorIs(t, <-consumer, context.Canceled)
}

func TestMemoryDriverRecoversHandlerPanic(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(context.Background(), "work", func(context.Context, contract.JobDelivery) error {
			panic("handler panic")
		})
	}()

	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	require.Error(t, <-consumer)
}

func TestMemoryDriverRespectsCanceledContexts(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, driver.Dispatch(ctx, contract.JobMessage{Queue: "work"}), context.Canceled)
	require.ErrorIs(t, driver.Consume(ctx, "work", func(context.Context, contract.JobDelivery) error { return nil }), context.Canceled)
}

func TestMemoryDriverShutdownRejectsAdmissionsAndDiscardsDelayedJobs(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	started := make(chan struct{})
	release := make(chan struct{})
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(ctx, "work", func(ctx context.Context, delivery contract.JobDelivery) error {
			close(started)
			<-release
			return delivery.Retry(ctx, time.Hour)
		})
	}()
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	<-started
	shutdown := make(chan error, 1)
	go func() { shutdown <- driver.Shutdown(context.Background()) }()
	require.Eventually(t, func() bool {
		return errors.Is(driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}), job.ErrDriverClosed)
	}, time.Second, time.Millisecond)
	close(release)
	require.ErrorIs(t, <-consumer, job.ErrDriverClosed)
	require.NoError(t, <-shutdown)
}

func TestMemoryDriverShutdownHonorsDeadline(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	started := make(chan struct{})
	release := make(chan struct{})
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(context.Background(), "work", func(context.Context, contract.JobDelivery) error {
			close(started)
			<-release
			return nil
		})
	}()
	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	t.Cleanup(cancel)
	require.ErrorIs(t, driver.Shutdown(ctx), context.DeadlineExceeded)
	close(release)
	require.Error(t, <-consumer)
	require.NoError(t, driver.Close())
}

func TestMemoryDriverShutdownDiscardsScheduledRetry(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.DefaultMemoryDriverConfig)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	scheduled := make(chan struct{})
	consumer := make(chan error, 1)
	go func() {
		consumer <- driver.Consume(ctx, "work", func(ctx context.Context, delivery contract.JobDelivery) error {
			defer close(scheduled)

			return delivery.Retry(ctx, time.Hour)
		})
	}()

	require.NoError(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "work"}))
	<-scheduled
	require.NoError(t, driver.Shutdown(context.Background()))
	require.ErrorIs(t, <-consumer, job.ErrDriverClosed)
}

func TestMemoryDriverRejectsInvalidConfigurationInputs(t *testing.T) {
	t.Parallel()

	driver := job.NewMemoryDriver(job.MemoryDriverConfig{RetryDelay: time.Millisecond})
	t.Cleanup(func() { require.NoError(t, driver.Close()) })
	require.ErrorIs(t, driver.Dispatch(context.Background(), contract.JobMessage{}), job.ErrInvalidQueue)
	require.ErrorIs(t, driver.Consume(context.Background(), "", func(context.Context, contract.JobDelivery) error { return nil }), job.ErrInvalidQueue)
	require.ErrorIs(t, driver.Consume(context.Background(), "work", nil), job.ErrInvalidHandler)
}

type resolver func(string) (contract.Job, error)

func (resolve resolver) ResolveJob(name string) (contract.Job, error) {
	return resolve(name)
}

type failingJob struct {
	Value   string `json:"value"`
	handled *atomic.Int64
}

func (*failingJob) Name() string                      { return "failing" }
func (*failingJob) Queue() string                     { return "work" }
func (*failingJob) RetryPolicy() contract.RetryPolicy { return contract.RetryPolicy{MaxAttempts: 2} }
func (job *failingJob) Handle(context.Context) error {
	job.handled.Add(1)

	return errors.New("failed")
}
