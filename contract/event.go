package contract

import "context"

type EventPayload = func(dest any) error
type EventHandler = func(payload EventPayload)
type EventUnsubscribeFunc = func() error

type Events interface {
	Publish(ctx context.Context, event string, payload any) error
	Subscribe(ctx context.Context, event string, handler EventHandler) (EventUnsubscribeFunc, error)
	Close() error
}
