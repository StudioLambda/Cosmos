package memory

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/studiolambda/cosmos/contract"
	core "github.com/studiolambda/cosmos/framework/event/internal/event"
)

// ErrBrokerClosed is returned when attempting operations on a closed broker.
var ErrBrokerClosed = errors.New("broker is closed")

// DefaultMaxConcurrentDeliveries is the maximum number of
// concurrent handler goroutines allowed per MemoryBroker.
const DefaultMaxConcurrentDeliveries = 1024

// MemoryBrokerConfig configures the in-memory event broker.
type MemoryBrokerConfig struct {
	// MaxConcurrentDeliveries is the maximum number of concurrent
	// handler goroutines allowed per MemoryBroker.
	MaxConcurrentDeliveries int

	// Logger records recovered handler panics. A nil logger discards records.
	Logger *contract.Logger
}

// DefaultMemoryBrokerConfig returns the default in-memory event broker configuration.
func DefaultMemoryBrokerConfig() MemoryBrokerConfig {
	return MemoryBrokerConfig{MaxConcurrentDeliveries: DefaultMaxConcurrentDeliveries}
}

// FromConfiguration populates the memory-broker configuration from configuration.
func (config *MemoryBrokerConfig) FromConfiguration(configuration *contract.Configuration) {
	config.MaxConcurrentDeliveries = configuration.GetOr("max_concurrent_deliveries", 0)
}

// MemoryBroker implements [contract.EventPublisherDriver] and
// [contract.EventSubscriberDriver] using only in-memory
// data structures with no external dependencies.
//
// Wildcard patterns: '*' matches a single dot-separated token,
// '#' matches zero or more tokens (must be the last token in the pattern).
type MemoryBroker struct {
	// lifecycle serializes delivery admission with Shutdown so no WaitGroup work
	// is added after shutdown begins waiting.
	lifecycle sync.Mutex
	mu        sync.RWMutex
	handlers  map[string]map[string]contract.EventHandler
	nextID    atomic.Uint64
	closed    atomic.Bool
	sem       chan struct{}
	wg        sync.WaitGroup
	logger    *contract.Logger
}

// NewMemoryBroker creates a new in-memory event broker.
func NewMemoryBroker(config MemoryBrokerConfig) *MemoryBroker {
	if config.MaxConcurrentDeliveries == 0 {
		config.MaxConcurrentDeliveries = DefaultMemoryBrokerConfig().MaxConcurrentDeliveries
	}

	if config.Logger == nil {
		config.Logger = contract.NewLogger(nil)
	}

	return &MemoryBroker{
		handlers: make(map[string]map[string]contract.EventHandler),
		sem:      make(chan struct{}, config.MaxConcurrentDeliveries),
		logger:   config.Logger,
	}
}

// Publish sends raw payload bytes to all matching subscribers.
// Handlers are invoked asynchronously with panic recovery.
func (broker *MemoryBroker) Publish(
	ctx context.Context,
	event string,
	payload []byte,
) error {
	if err := core.ValidateName(event); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	broker.mu.RLock()

	var matched []contract.EventHandler

	for pattern, patternHandlers := range broker.handlers {
		if matchEvent(pattern, event) {
			for _, handler := range patternHandlers {
				matched = append(matched, handler)
			}
		}
	}

	broker.mu.RUnlock()

	broker.lifecycle.Lock()

	if broker.closed.Load() {
		broker.lifecycle.Unlock()

		return ErrBrokerClosed
	}

	broker.wg.Add(len(matched))
	broker.lifecycle.Unlock()

	for _, handler := range matched {
		broker.sem <- struct{}{}

		go func() {
			defer func() {
				<-broker.sem
				broker.wg.Done()
			}()

			broker.deliverToHandler(handler, payload)
		}()
	}

	return nil
}

// Subscribe registers a handler for events matching the given pattern.
// Returns an unsubscribe function.
func (broker *MemoryBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if err := core.ValidatePattern(event); err != nil {
		return nil, err
	}

	handlerID := strconv.FormatUint(broker.nextID.Add(1), 10)

	broker.lifecycle.Lock()
	defer broker.lifecycle.Unlock()

	if broker.closed.Load() {
		return nil, ErrBrokerClosed
	}

	broker.mu.Lock()
	defer broker.mu.Unlock()

	if broker.handlers[event] == nil {
		broker.handlers[event] = make(map[string]contract.EventHandler)
	}

	broker.handlers[event][handlerID] = handler

	return func() error {
		broker.mu.Lock()
		defer broker.mu.Unlock()

		if patternHandlers, ok := broker.handlers[event]; ok {
			delete(patternHandlers, handlerID)

			if len(patternHandlers) == 0 {
				delete(broker.handlers, event)
			}
		}

		return nil
	}, nil
}

// Shutdown stops the broker and waits for in-flight deliveries until ctx expires.
func (broker *MemoryBroker) Shutdown(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	broker.lifecycle.Lock()
	broker.closed.Store(true)
	broker.lifecycle.Unlock()

	if err := waitGroupContext(ctx, &broker.wg); err != nil {
		return err
	}

	broker.mu.Lock()
	defer broker.mu.Unlock()

	broker.handlers = make(map[string]map[string]contract.EventHandler)

	return nil
}

// waitGroupContext waits for group until ctx expires.
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

// deliverToHandler invokes a handler with the raw payload,
// recovering from any panic.
func (broker *MemoryBroker) deliverToHandler(
	handler contract.EventHandler,
	payload []byte,
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			broker.logger.Error(
				"event handler panicked",
				"error", recovered,
			)
		}
	}()

	handler(payload)
}

// matchEvent checks if a subscription pattern matches an event name.
func matchEvent(pattern, event string) bool {
	return core.Match(pattern, event)
}
