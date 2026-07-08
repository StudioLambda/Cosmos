package contract

import (
	"context"
	"encoding/json/v2"
)

// EventHandler is a callback function invoked when a subscribed
// event is received. It receives the raw JSON payload bytes.
type EventHandler = func(payload []byte)

// EventUnsubscribeFunc is a function returned by subscription
// that cancels the subscription when called.
type EventUnsubscribeFunc = func() error

type EventDecoder[T any] = func() (T, error)

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

	// Ping verifies that the connection is still alive.
	Ping(ctx context.Context) error

	// Close shuts down the event system and releases resources.
	Close() error
}

// Events provides a type-safe event bus over an [EventDriver].
// It handles JSON serialization of payloads and deserialization
// in subscriber callbacks.
type Events struct {
	driver EventDriver
}

// NewEvents creates a new [Events] that delegates to the given driver.
//
// Example:
//
//	events := contract.NewEvents(driver)
//	defer events.Close()
func NewEvents(driver EventDriver) *Events {
	return &Events{driver: driver}
}

// Driver returns the underlying [EventDriver].
//
// Example:
//
//	events := contract.NewEvents(driver)
//	raw := events.Driver()
//	_ = raw
func (events *Events) Driver() EventDriver {
	return events.driver
}

// Publish JSON-encodes the payload and sends it to all subscribers
// of the named event.
//
// Example:
//
//	if err := events.Publish(ctx, "users.created", UserCreated{ID: 1}); err != nil {
//		return err
//	}
func (events *Events) Publish[T any](ctx context.Context, event string, payload T) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	return events.driver.Publish(ctx, event, encoded)
}

// Subscribe registers a handler for the named event. The handler
// receives a decode function that unmarshals the event payload into
// a destination pointer. Returns a function to cancel the subscription.
//
// Example:
//
//	unsubscribe, err := events.Subscribe[UserCreated](ctx, "users.created", func(decode EventDecoder[UserCreated]) {
//		msg, err := decode()
//		if err != nil {
//			return
//		}
//		_ = msg
//	})
//	if err != nil {
//		return err
//	}
//	defer unsubscribe()
func (events *Events) Subscribe[T any](ctx context.Context, event string, handler func(decode EventDecoder[T])) (EventUnsubscribeFunc, error) {
	return events.driver.Subscribe(ctx, event, func(payload []byte) {
		handler(func() (res T, err error) {
			if err := json.Unmarshal(payload, &res); err != nil {
				return res, err
			}

			return res, nil
		})
	})
}

// Close shuts down the event system and releases resources.
//
// Example:
//
//	if err := events.Close(); err != nil {
//		return err
//	}
func (events *Events) Close() error {
	return events.driver.Close()
}
