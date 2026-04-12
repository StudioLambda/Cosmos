package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/studiolambda/cosmos/contract"
)

var (
	// ErrBrokerClosed is returned when attempting operations on a closed
	// broker.
	// Once a broker is closed, it cannot be reused and a new instance
	// must be created.
	ErrBrokerClosed = errors.New("broker is closed")
)

// DefaultMaxConcurrentDeliveries is the maximum number of
// concurrent handler goroutines allowed per MemoryBroker.
// This prevents goroutine exhaustion under high event throughput
// with slow handlers.
const DefaultMaxConcurrentDeliveries = 1024

// MemoryBroker implements the EventBroker interface using only in-memory
// data structures with no external dependencies.
// It provides a lightweight, zero-configuration broker ideal for testing,
// local development, and single-instance applications.
// All message delivery happens asynchronously in separate goroutines with
// panic recovery to ensure one handler's failure doesn't affect others.
// Concurrent deliveries are bounded by a semaphore to prevent goroutine
// exhaustion.
type MemoryBroker struct {
	// mu protects concurrent access to the handlers map during
	// subscribe and unsubscribe operations.
	mu sync.RWMutex

	// handlers stores event handlers organized by subscription pattern
	// and unique handler ID.
	// The outer map key is the pattern (which may contain wildcards),
	// and the inner map associates handler IDs with their handlers.
	handlers map[string]map[string]contract.EventHandler

	// nextID generates unique identifiers for each subscribed handler
	// to enable precise unsubscribe operations.
	nextID atomic.Uint64

	// closed indicates whether the broker has been closed.
	// Once closed, all operations return ErrBrokerClosed.
	closed atomic.Bool

	// sem limits the number of concurrent delivery goroutines to
	// prevent resource exhaustion under high throughput.
	sem chan struct{}

	// wg tracks in-flight deliveries so [Close] can wait for
	// them to complete.
	wg sync.WaitGroup
}

// NewMemoryBroker creates a new in-memory event broker with no external
// dependencies or configuration required.
// The broker is ready to use immediately and supports concurrent access
// from multiple goroutines.
// It must be closed with Close when no longer needed to release resources.
func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{
		handlers: make(map[string]map[string]contract.EventHandler),
		sem:      make(chan struct{}, DefaultMaxConcurrentDeliveries),
	}
}

// Publish sends an event with the given payload to all matching subscribers.
// The payload is JSON-encoded and delivered asynchronously to handlers
// whose subscription patterns match the event name.
// Handlers are invoked in separate goroutines with panic recovery,
// ensuring one handler's failure doesn't affect others.
//
// Wildcard matching supports:
//   - "*" matches a single token (e.g., "user.*.created" matches
//     "user.123.created")
//   - "#" matches zero or more tokens (e.g., "logs.#" matches "logs",
//     "logs.error", "logs.error.database")
//
// Returns an error if JSON encoding fails or if the broker is closed.
// The context is checked once at the start of the publish operation.
func (broker *MemoryBroker) Publish(
	ctx context.Context,
	event string,
	payload any,
) error {
	if broker.closed.Load() {
		return ErrBrokerClosed
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	encoded, err := json.Marshal(payload)

	if err != nil {
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

			broker.deliverToHandler(handler, encoded)
		}()
	}

	return nil
}

// Subscribe registers a handler for events matching the given pattern.
// The pattern supports wildcards:
//   - "*" matches a single token (e.g., "user.*.created")
//   - "#" matches zero or more tokens (e.g., "logs.#")
//
// Multiple handlers can subscribe to the same pattern and all will receive
// messages (fan-out).
// Handlers are invoked asynchronously in separate goroutines.
//
// Returns an unsubscribe function that removes only this specific handler
// subscription.
// Returns an error if the broker is closed.
// The context is used only for the subscription setup, not for the handler
// lifecycle.
func (broker *MemoryBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if broker.closed.Load() {
		return nil, ErrBrokerClosed
	}

	handlerID := fmt.Sprintf("%d", broker.nextID.Add(1))

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

// Close shuts down the broker and removes all subscribed handlers.
// After Close is called, all operations return ErrBrokerClosed and the
// broker cannot be reused.
// Close waits for all in-flight deliveries to complete before returning.
func (broker *MemoryBroker) Close() error {
	broker.closed.Store(true)

	broker.wg.Wait()

	broker.mu.Lock()
	defer broker.mu.Unlock()

	broker.handlers = make(map[string]map[string]contract.EventHandler)

	return nil
}

// deliverToHandler invokes a handler with the encoded payload
// in a goroutine with panic recovery. Recovered panics are
// logged via slog so they remain visible for debugging.
// This ensures that a panic in one handler doesn't affect
// other handlers or the broker itself.
func (broker *MemoryBroker) deliverToHandler(
	handler contract.EventHandler,
	encoded []byte,
) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error(
				"event handler panicked",
				"error", recovered,
			)
		}
	}()

	handler(func(dest any) error {
		return json.Unmarshal(encoded, dest)
	})
}

// matchEvent checks if a subscription pattern matches an
// event name. It supports dot-separated tokens with wildcards:
//   - "*" matches exactly one token
//   - "#" matches zero or more tokens
func matchEvent(pattern, event string) bool {
	if pattern == event {
		return true
	}

	patternParts := strings.Split(pattern, ".")
	eventParts := strings.Split(event, ".")

	return matchEventParts(patternParts, eventParts)
}

// matchEventParts recursively matches event parts against
// pattern parts with wildcard support. It handles "*" for
// single-token and "#" for multi-token matching.
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
