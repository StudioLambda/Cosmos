package contract

import (
	"context"
	"encoding/json"
)

// EventHandler is a callback function invoked when a subscribed
// event is received. It receives the raw JSON payload bytes.
type EventHandler = func(payload []byte)

// EventUnsubscribeFunc is a function returned by subscription
// that cancels the subscription when called.
type EventUnsubscribeFunc = func() error

// EventDriver defines the contract for a publish/subscribe event
// system backend. Drivers handle raw byte delivery; the [Events]
// wrapper adds JSON serialization on top.
type EventDriver interface {
	// Publish sends raw bytes to all subscribers of the named event.
	Publish(ctx context.Context, event string, payload []byte) error

	// Subscribe registers a handler for the named event. The handler
	// receives raw payload bytes. Returns a function to cancel the
	// subscription.
	Subscribe(ctx context.Context, event string, handler EventHandler) (EventUnsubscribeFunc, error)

	// Close shuts down the event system and releases resources.
	Close() error
}

// Events provides a type-safe event bus over an [EventDriver].
// It handles JSON serialization of payloads and deserialization
// in subscriber callbacks. When generic methods become available
// in Go, Publish and Subscribe will accept typed values directly.
type Events struct {
	driver EventDriver
}

// NewEvents creates a new [Events] that delegates to the given driver.
func NewEvents(driver EventDriver) *Events {
	return &Events{driver: driver}
}

// Driver returns the underlying [EventDriver].
func (events *Events) Driver() EventDriver {
	return events.driver
}

// Publish JSON-encodes the payload and sends it to all subscribers
// of the named event.
func (events *Events) Publish(ctx context.Context, event string, payload any) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	return events.driver.Publish(ctx, event, encoded)
}

// Subscribe registers a handler for the named event. The handler
// receives a decode function that unmarshals the event payload into
// a destination pointer. Returns a function to cancel the subscription.
func (events *Events) Subscribe(ctx context.Context, event string, handler func(func(dest any) error)) (EventUnsubscribeFunc, error) {
	return events.driver.Subscribe(ctx, event, func(payload []byte) {
		handler(func(dest any) error {
			return json.Unmarshal(payload, dest)
		})
	})
}

// Close shuts down the event system and releases resources.
func (events *Events) Close() error {
	return events.driver.Close()
}
