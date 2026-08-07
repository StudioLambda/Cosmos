package memory

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/studiolambda/cosmos/contract"
)

var (
	// ErrDriverClosed is returned when an operation is attempted after shutdown begins.
	ErrDriverClosed = errors.New("job driver is closed")

	// ErrInvalidQueue is returned when a queue name is empty.
	ErrInvalidQueue = errors.New("invalid job queue")

	// ErrInvalidHandler is returned when Consume is given a nil handler.
	ErrInvalidHandler = errors.New("invalid job delivery handler")

	// ErrDeliverySettled is returned when a delivery is settled more than once.
	ErrDeliverySettled = errors.New("job delivery already settled")

	// ErrDeliveryUnsettled is returned when a handler returns without settling a delivery.
	ErrDeliveryUnsettled = errors.New("job delivery was not settled")
)

// MemoryDriverConfig configures a [MemoryDriver].
type MemoryDriverConfig struct {
	// RetryDelay is used when a delivery requests a retry with a zero delay.
	RetryDelay time.Duration
}

// DefaultMemoryDriverConfig holds the default in-memory job driver configuration.
var DefaultMemoryDriverConfig = MemoryDriverConfig{}

// MemoryDriver implements [contract.JobDispatcherDriver],
// [contract.JobConsumerDriver], and [contract.Shutdowner] with in-memory data
// structures. It is non-durable: process exit or crash loses queued jobs.
type MemoryDriver struct {
	// lifecycle serializes Consume and Dispatch admission with Shutdown.
	lifecycle     sync.Mutex
	mu            sync.Mutex
	queues        map[string]*queue
	delayed       delayedHeap
	changed       chan struct{}
	done          chan struct{}
	closed        bool
	retryDelay    time.Duration
	nextID        atomic.Uint64
	active        sync.WaitGroup
	schedulerDone chan struct{}
}

type queue struct {
	ready   []contract.JobMessage
	changed chan struct{}
}

type delayedMessage struct {
	message  contract.JobMessage
	due      time.Time
	sequence uint64
}

type delayedHeap []delayedMessage

func (messages delayedHeap) Len() int {
	return len(messages)
}

func (messages delayedHeap) Less(first, second int) bool {
	if messages[first].due.Equal(messages[second].due) {
		return messages[first].sequence < messages[second].sequence
	}

	return messages[first].due.Before(messages[second].due)
}

func (messages delayedHeap) Swap(first, second int) {
	messages[first], messages[second] = messages[second], messages[first]
}

func (messages *delayedHeap) Push(value any) {
	*messages = append(*messages, value.(delayedMessage))
}

func (messages *delayedHeap) Pop() any {
	old := *messages
	last := len(old) - 1
	message := old[last]
	*messages = old[:last]

	return message
}

// NewMemoryDriver creates a non-durable in-memory job driver.
func NewMemoryDriver(config MemoryDriverConfig) *MemoryDriver {
	driver := &MemoryDriver{
		queues:        make(map[string]*queue),
		changed:       make(chan struct{}),
		done:          make(chan struct{}),
		retryDelay:    config.RetryDelay,
		schedulerDone: make(chan struct{}),
	}

	go driver.schedule()

	return driver
}

// Dispatch adds message to its named queue. Payload is copied before acceptance.
func (driver *MemoryDriver) Dispatch(ctx context.Context, message contract.JobMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if message.Queue == "" {
		return ErrInvalidQueue
	}

	driver.lifecycle.Lock()
	defer driver.lifecycle.Unlock()

	if driver.closed {
		return ErrDriverClosed
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if message.ID == "" {
		message.ID = fmt.Sprintf("memory-%d", driver.nextID.Add(1))
	}

	// This driver supplies attempt counts at delivery time, not dispatch time.
	message.Attempts = 0
	message.Payload = slices.Clone(message.Payload)
	driver.enqueue(message)

	return nil
}

// Consume receives and handles deliveries from queue until ctx is canceled or
// the driver shuts down. Concurrent consumers for one queue compete for each
// message; each reserved message is delivered to only one handler.
func (driver *MemoryDriver) Consume(ctx context.Context, name string, handler contract.JobDeliveryHandler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if name == "" {
		return ErrInvalidQueue
	}

	if handler == nil {
		return ErrInvalidHandler
	}

	driver.lifecycle.Lock()

	if driver.closed {
		driver.lifecycle.Unlock()

		return ErrDriverClosed
	}

	driver.active.Add(1)
	driver.lifecycle.Unlock()
	defer driver.active.Done()

	for {
		message, changed := driver.reserve(name)
		if message != nil {
			delivery := &memoryDelivery{driver: driver, message: *message}
			if err := driver.handle(ctx, handler, delivery); err != nil {
				return err
			}

			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-driver.done:
			return ErrDriverClosed
		case <-changed:
		}
	}
}

// Shutdown stops new admissions, discards queued jobs, and waits for active
// consumers, handlers, and the delayed scheduler until ctx expires.
func (driver *MemoryDriver) Shutdown(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	driver.lifecycle.Lock()
	if !driver.closed {
		driver.closed = true
		close(driver.done)
	}
	driver.lifecycle.Unlock()

	driver.mu.Lock()
	driver.queues = make(map[string]*queue)
	driver.delayed = nil
	driver.signalSchedulerLocked()
	driver.mu.Unlock()

	if err := waitGroupContext(ctx, &driver.active); err != nil {
		return err
	}

	select {
	case <-driver.schedulerDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close closes the driver using a background context.
func (driver *MemoryDriver) Close() error {
	return driver.Shutdown(context.Background())
}

func (driver *MemoryDriver) reserve(name string) (*contract.JobMessage, <-chan struct{}) {
	driver.mu.Lock()
	defer driver.mu.Unlock()

	queue := driver.queueLocked(name)
	if len(queue.ready) == 0 {
		return nil, queue.changed
	}

	message := queue.ready[0]
	queue.ready = queue.ready[1:]
	message.Attempts++

	return &message, nil
}

func (driver *MemoryDriver) handle(ctx context.Context, handler contract.JobDeliveryHandler, delivery *memoryDelivery) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("job delivery handler panicked: %v", recovered)
			slog.Error("job delivery handler panicked", "error", recovered, "job_id", delivery.message.ID, "queue", delivery.message.Queue)
		}

		if err != nil {
			slog.Error("job delivery handler failed", "error", err, "job_id", delivery.message.ID, "queue", delivery.message.Queue)
			_ = delivery.Fail(context.Background(), err)
		}

		if !delivery.settled() {
			err = ErrDeliveryUnsettled
			slog.Error("job delivery handler did not settle delivery", "job_id", delivery.message.ID, "queue", delivery.message.Queue)
			_ = delivery.Fail(context.Background(), err)
		}
	}()

	return handler(ctx, delivery)
}

func (driver *MemoryDriver) enqueue(message contract.JobMessage) {
	driver.mu.Lock()
	defer driver.mu.Unlock()

	driver.enqueueLocked(message)
}

func (driver *MemoryDriver) enqueueLocked(message contract.JobMessage) {
	queue := driver.queueLocked(message.Queue)
	queue.ready = append(queue.ready, message)
	close(queue.changed)
	queue.changed = make(chan struct{})
}

func (driver *MemoryDriver) queueLocked(name string) *queue {
	if driver.queues[name] == nil {
		driver.queues[name] = &queue{changed: make(chan struct{})}
	}

	return driver.queues[name]
}

func (driver *MemoryDriver) retry(ctx context.Context, message contract.JobMessage, delay time.Duration) error {
	driver.lifecycle.Lock()
	defer driver.lifecycle.Unlock()

	if driver.closed {
		return ErrDriverClosed
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if delay == 0 {
		delay = driver.retryDelay
	}

	if delay <= 0 {
		driver.enqueue(message)

		return nil
	}

	driver.mu.Lock()
	heap.Push(&driver.delayed, delayedMessage{
		message:  message,
		due:      time.Now().Add(delay),
		sequence: driver.nextID.Add(1),
	})
	driver.signalSchedulerLocked()
	driver.mu.Unlock()

	return nil
}

func (driver *MemoryDriver) schedule() {
	defer close(driver.schedulerDone)

	for {
		due, changed, ok := driver.nextScheduled()
		if !ok {
			select {
			case <-driver.done:
				return
			case <-changed:
			}

			continue
		}

		timer := time.NewTimer(time.Until(due))
		select {
		case <-driver.done:
			stopTimer(timer)

			return
		case <-changed:
			stopTimer(timer)
		case <-timer.C:
			driver.promoteDue()
		}
	}
}

func (driver *MemoryDriver) nextScheduled() (time.Time, <-chan struct{}, bool) {
	driver.mu.Lock()
	defer driver.mu.Unlock()

	if len(driver.delayed) == 0 {
		return time.Time{}, driver.changed, false
	}

	return driver.delayed[0].due, driver.changed, true
}

func (driver *MemoryDriver) promoteDue() {
	driver.lifecycle.Lock()
	defer driver.lifecycle.Unlock()

	if driver.closed {
		return
	}

	driver.mu.Lock()
	defer driver.mu.Unlock()

	now := time.Now()
	for len(driver.delayed) > 0 && !driver.delayed[0].due.After(now) {
		message := heap.Pop(&driver.delayed).(delayedMessage)
		driver.enqueueLocked(message.message)
	}
}

func (driver *MemoryDriver) signalSchedulerLocked() {
	close(driver.changed)
	driver.changed = make(chan struct{})
}

func waitGroupContext(ctx context.Context, group *sync.WaitGroup) error {
	done := make(chan struct{})

	go func() {
		group.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

type memoryDelivery struct {
	driver   *MemoryDriver
	message  contract.JobMessage
	mu       sync.Mutex
	done     bool
	settling bool
}

func (delivery *memoryDelivery) Message() contract.JobMessage {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()

	message := delivery.message
	message.Payload = slices.Clone(message.Payload)

	return message
}

func (delivery *memoryDelivery) Acknowledge(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return nil })
}

func (delivery *memoryDelivery) Retry(ctx context.Context, delay time.Duration) error {
	return delivery.settle(ctx, func() error {
		return delivery.driver.retry(ctx, delivery.message, delay)
	})
}

func (delivery *memoryDelivery) Reject(ctx context.Context) error {
	return delivery.settle(ctx, func() error { return nil })
}

func (delivery *memoryDelivery) Fail(ctx context.Context, err error) error {
	return delivery.settle(ctx, func() error {
		if err != nil {
			slog.Error("job delivery failed", "error", err, "job_id", delivery.message.ID, "queue", delivery.message.Queue)
		}

		return nil
	})
}

func (delivery *memoryDelivery) settle(ctx context.Context, operation func() error) error {
	delivery.mu.Lock()

	if delivery.done || delivery.settling {
		delivery.mu.Unlock()

		return ErrDeliverySettled
	}

	if err := ctx.Err(); err != nil {
		delivery.mu.Unlock()

		return err
	}

	delivery.settling = true
	delivery.mu.Unlock()

	if err := operation(); err != nil {
		delivery.mu.Lock()
		delivery.settling = false
		delivery.mu.Unlock()

		return err
	}

	delivery.mu.Lock()
	delivery.settling = false
	delivery.done = true
	delivery.mu.Unlock()

	return nil
}

func (delivery *memoryDelivery) settled() bool {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()

	return delivery.done
}
