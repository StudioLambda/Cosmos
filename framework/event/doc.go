// Package event provides event bus implementations and topic matching utilities.
//
// All implementations expose publish/subscribe semantics through contract
// abstractions and encode payloads as JSON for transport.
//
// # Concurrency
//
// Subscribers are invoked asynchronously where supported by the backend. Handler
// code should be concurrency-safe and resilient to duplicate delivery semantics
// depending on broker guarantees.
//
// Example
//
//	bus := event.NewMemoryBroker()
//	defer bus.Close()
//
//	ev := contract.NewEvents(bus)
//	_, _ = ev.Subscribe[UserCreated](ctx, "user.created", func(decode contract.EventDecoder[UserCreated]) {
//		e, _ := decode()
//		_ = e
//	})
//
//	_ = ev.Publish(ctx, "user.created", UserCreated{ID: 1})
package event
