package contract

import (
	"context"
	"encoding/json/v2"
)

// EventHandler is a callback function invoked when a subscribed event is
// received. It receives the raw payload bytes.
type EventHandler = func(payload []byte)

// EventUnsubscribeFunc is a function returned by subscription
// that cancels the subscription when called.
type EventUnsubscribeFunc = func() error

// EventDecoder decodes an event payload into T.
type EventDecoder[T any] = func() (T, error)

// EventPublisherDriver publishes raw event payloads. A successful publish does
// not confirm that any subscriber processed the event.
type EventPublisherDriver interface {
	// Publish sends raw bytes to all active subscribers of the named event.
	// Event names are dot-separated tokens and cannot contain wildcards.
	Publish(ctx context.Context, event string, payload []byte) error
}

// EventSubscriberDriver subscribes to raw event payloads. Drivers do not
// provide persistence or retries.
type EventSubscriberDriver interface {
	// Subscribe registers a handler for an event name or pattern. Patterns use
	// '*' for exactly one token and a final '#' for zero or more trailing tokens.
	// The handler receives raw payload bytes. Returns a function to cancel the
	// subscription.
	Subscribe(ctx context.Context, event string, handler EventHandler) (EventUnsubscribeFunc, error)
}

// EventPublisher JSON-encodes event payloads before publishing them through an
// [EventPublisherDriver].
type EventPublisher struct {
	driver EventPublisherDriver
}

// NewEventPublisher creates an [EventPublisher] that delegates to driver.
//
// Example:
//
//	publisher := contract.NewEventPublisher(driver)
func NewEventPublisher(driver EventPublisherDriver) *EventPublisher {
	return &EventPublisher{driver: driver}
}

// Driver returns the underlying [EventPublisherDriver].
//
// Example:
//
//	publisher := contract.NewEventPublisher(driver)
//	raw := publisher.Driver()
//	_ = raw
func (publisher *EventPublisher) Driver() EventPublisherDriver {
	return publisher.driver
}

// Publish JSON-encodes the payload and sends it to all subscribers
// of the named event.
//
// Example:
//
//	if err := publisher.Publish(ctx, "users.created", UserCreated{ID: 1}); err != nil {
//		return err
//	}
func (publisher *EventPublisher) Publish[T any](ctx context.Context, event string, payload T) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	return publisher.driver.Publish(ctx, event, encoded)
}

// EventSubscriber JSON-decodes event payloads received through an
// [EventSubscriberDriver].
type EventSubscriber struct {
	driver EventSubscriberDriver
}

// NewEventSubscriber creates an [EventSubscriber] that delegates to driver.
//
// Example:
//
//	subscriber := contract.NewEventSubscriber(driver)
func NewEventSubscriber(driver EventSubscriberDriver) *EventSubscriber {
	return &EventSubscriber{driver: driver}
}

// Driver returns the underlying [EventSubscriberDriver].
//
// Example:
//
//	subscriber := contract.NewEventSubscriber(driver)
//	raw := subscriber.Driver()
//	_ = raw
func (subscriber *EventSubscriber) Driver() EventSubscriberDriver {
	return subscriber.driver
}

// Subscribe registers a handler for an event name or pattern. The handler
// receives a decode function that unmarshals the event payload into T. Returns
// a function to cancel the subscription.
//
// Example:
//
//	unsubscribe, err := subscriber.Subscribe[UserCreated](ctx, "users.created", func(decode EventDecoder[UserCreated]) {
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
func (subscriber *EventSubscriber) Subscribe[T any](ctx context.Context, event string, handler func(decode EventDecoder[T])) (EventUnsubscribeFunc, error) {
	return subscriber.driver.Subscribe(ctx, event, func(payload []byte) {
		handler(func() (res T, err error) {
			if err := json.Unmarshal(payload, &res); err != nil {
				return res, err
			}

			return res, nil
		})
	})
}
