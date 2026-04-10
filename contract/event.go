package contract

import "context"

// EventPayload is a function that decodes an event's raw payload
// into the provided destination. The destination must be a pointer
// to the expected type.
type EventPayload = func(dest any) error

// EventHandler is a callback function invoked when a subscribed
// event is received. It receives an [EventPayload] that can be
// used to decode the event data.
type EventHandler = func(payload EventPayload)

// EventUnsubscribeFunc is a function returned by [Events.Subscribe]
// that cancels the subscription when called. It returns an error
// if the unsubscription fails.
type EventUnsubscribeFunc = func() error

// Events defines the contract for a publish/subscribe event system.
// Implementations handle event routing between publishers and subscribers
// across different transport backends.
type Events interface {
	// Publish sends the given payload to all subscribers of the named event.
	// The payload is serialized by the implementation before delivery.
	Publish(ctx context.Context, event string, payload any) error

	// Subscribe registers a handler for the named event and returns
	// a function to cancel the subscription. The handler is called
	// each time a matching event is received.
	Subscribe(ctx context.Context, event string, handler EventHandler) (EventUnsubscribeFunc, error)

	// Close shuts down the event system and releases any underlying
	// connections or resources.
	Close() error
}
