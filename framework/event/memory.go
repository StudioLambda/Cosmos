package event

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/studiolambda/cosmos/contract"
)

// ErrBrokerClosed is returned when attempting operations on a closed broker.
var ErrBrokerClosed = errors.New("broker is closed")

// DefaultMaxConcurrentDeliveries is the maximum number of
// concurrent handler goroutines allowed per MemoryBroker.
const DefaultMaxConcurrentDeliveries = 1024

// MemoryBroker implements [contract.EventDriver] using only in-memory
// data structures with no external dependencies.
//
// Wildcard patterns: '*' matches a single dot-separated token,
// '#' matches zero or more tokens (must be the last token in the pattern).
type MemoryBroker struct {
	mu       sync.RWMutex
	handlers map[string]map[string]contract.EventHandler
	nextID   atomic.Uint64
	closed   atomic.Bool
	sem      chan struct{}
	wg       sync.WaitGroup
}

// NewMemoryBroker creates a new in-memory event broker.
func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{
		handlers: make(map[string]map[string]contract.EventHandler),
		sem:      make(chan struct{}, DefaultMaxConcurrentDeliveries),
	}
}

// Publish sends raw payload bytes to all matching subscribers.
// Handlers are invoked asynchronously with panic recovery.
func (broker *MemoryBroker) Publish(
	ctx context.Context,
	event string,
	payload []byte,
) error {
	if broker.closed.Load() {
		return ErrBrokerClosed
	}

	if err := validateEvent(event); err != nil {
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

	for _, handler := range matched {
		broker.wg.Add(1)
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
	if broker.closed.Load() {
		return nil, ErrBrokerClosed
	}

	if err := validateEvent(event); err != nil {
		return nil, err
	}

	handlerID := strconv.FormatUint(broker.nextID.Add(1), 10)

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

// Close shuts down the broker and waits for in-flight deliveries.
func (broker *MemoryBroker) Close() error {
	broker.closed.Store(true)

	broker.wg.Wait()

	broker.mu.Lock()
	defer broker.mu.Unlock()

	broker.handlers = make(map[string]map[string]contract.EventHandler)

	return nil
}

// deliverToHandler invokes a handler with the raw payload,
// recovering from any panic.
func (broker *MemoryBroker) deliverToHandler(
	handler contract.EventHandler,
	payload []byte,
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error(
				"event handler panicked",
				"error", recovered,
			)
		}
	}()

	handler(payload)
}

// matchEvent checks if a subscription pattern matches an event name.
func matchEvent(pattern, event string) bool {
	if pattern == event {
		return true
	}

	patternParts := strings.Split(pattern, ".")
	eventParts := strings.Split(event, ".")

	return matchEventParts(patternParts, eventParts)
}

// matchEventParts recursively matches event parts against pattern parts.
func matchEventParts(pattern, event []string) bool {
	if len(pattern) == 0 {
		return len(event) == 0
	}

	if len(event) == 0 {
		return pattern[0] == "#"
	}

	if pattern[0] == "#" {
		return true
	}

	if pattern[0] == "*" || pattern[0] == event[0] {
		return matchEventParts(pattern[1:], event[1:])
	}

	return false
}
