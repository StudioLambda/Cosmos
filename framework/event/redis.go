package event

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/studiolambda/cosmos/contract"
)

type RedisBroker struct {
	client *redis.Client
}

type RedisBrokerOptions = redis.Options

func NewRedisBroker(options *redis.Options) *RedisBroker {
	client := redis.NewClient(options)

	return NewRedisBrokerFrom(client)
}

func NewRedisBrokerFrom(client *redis.Client) *RedisBroker {
	return &RedisBroker{
		client: client,
	}
}

func (b *RedisBroker) Publish(ctx context.Context, event string, payload any) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	return b.client.Publish(ctx, event, encoded).Err()
}

func (b *RedisBroker) Subscribe(ctx context.Context, event string, handler contract.EventHandler) (contract.EventUnsubscribeFunc, error) {
	event = strings.ReplaceAll(event, "#", "*")
	sub := b.client.PSubscribe(ctx, event)
	wg := sync.WaitGroup{}

	wg.Go(func() {
		for message := range sub.Channel() {
			handler(func(dest any) error {
				return json.Unmarshal([]byte(message.Payload), dest)
			})
		}
	})

	return func() error {
		defer wg.Wait()
		return sub.Close()
	}, nil
}

func (b *RedisBroker) Close() error {
	return b.client.Close()
}
