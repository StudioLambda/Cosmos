package amqp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"

	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/stretchr/testify/require"
)

func TestNormalizeConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	config, err := normalizeConfig(DriverConfig{Queues: map[string]QueueConfig{"work": {Name: "jobs.work", RoutingKey: "work"}}}, false)

	require.NoError(t, err)
	require.Equal(t, DefaultExchange, config.Exchange)
	require.Equal(t, defaultPrefetch, config.Prefetch)
	require.Equal(t, defaultRetryDelay, config.RetryDelay)
}

func TestNormalizeConfigRejectsMissingQueueRoutingKey(t *testing.T) {
	t.Parallel()

	_, err := normalizeConfig(DriverConfig{Queues: map[string]QueueConfig{"work": {Name: "jobs.work"}}}, false)

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestNormalizeConfigRequiresURLForNew(t *testing.T) {
	t.Parallel()

	_, err := normalizeConfig(DriverConfig{Queues: map[string]QueueConfig{"work": {Name: "jobs.work", RoutingKey: "work"}}}, true)

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestDeliveryMessageUsesAMQPMessageIDAndAttemptHeader(t *testing.T) {
	t.Parallel()

	message, err := deliveryMessage("work", amqp091.Delivery{
		Body:      []byte(`{"name":"email","queue":"untrusted","payload":"e30="}`),
		MessageId: "amqp-id",
		Headers:   amqp091.Table{attemptHeader: int32(3)},
	})

	require.NoError(t, err)
	require.Equal(t, "work", message.Queue)
	require.Equal(t, "amqp-id", message.ID)
	require.Equal(t, 3, message.Attempts)
}

func TestDeliveryMessageAddsAttemptForBrokerRedelivery(t *testing.T) {
	t.Parallel()

	message, err := deliveryMessage("work", amqp091.Delivery{Body: []byte(`{"name":"email","payload":"e30="}`), Redelivered: true})

	require.NoError(t, err)
	require.Equal(t, 2, message.Attempts)
}

func TestDeliveryMessageRejectsMalformedBody(t *testing.T) {
	t.Parallel()

	_, err := deliveryMessage("work", amqp091.Delivery{Body: []byte("not JSON")})

	require.ErrorIs(t, err, ErrInvalidDelivery)
}

func TestPublishingCreatesPersistentInitialAttempt(t *testing.T) {
	t.Parallel()

	published := publishing("message-1", []byte("body"), 1)

	require.Equal(t, amqp091.Persistent, published.DeliveryMode)
	require.Equal(t, "application/json", published.ContentType)
	require.Equal(t, "message-1", published.MessageId)
	require.Equal(t, int32(1), published.Headers[attemptHeader])
}

func TestRetryQueueNameUsesRoundedMilliseconds(t *testing.T) {
	t.Parallel()

	queue := retryQueueName("jobs.work", 1500*time.Millisecond+time.Nanosecond)

	require.Equal(t, "jobs.work.retry.1501", queue)
}

func TestDelayedPublishingIncrementsAttemptAndSetsTTL(t *testing.T) {
	t.Parallel()

	published := delayedPublishing("message-1", []byte("body"), 2, 1500*time.Millisecond+time.Nanosecond, amqp091.Table{"trace": "trace-1"})

	require.Equal(t, "1501", published.Expiration)
	require.Equal(t, int32(2), published.Headers[attemptHeader])
	require.Equal(t, "trace-1", published.Headers["trace"])
}

func TestRetryQueueArgumentsDeadLetterToWorkExchange(t *testing.T) {
	t.Parallel()

	arguments := retryQueueArguments("cosmos:jobs", "work")

	require.Equal(t, "cosmos:jobs", arguments["x-dead-letter-exchange"])
	require.Equal(t, "work", arguments["x-dead-letter-routing-key"])
}

func TestWaitConfirmIgnoresEarlierConfirmation(t *testing.T) {
	t.Parallel()

	confirms := make(chan amqp091.Confirmation, 2)
	confirms <- amqp091.Confirmation{DeliveryTag: 1, Ack: true}
	confirms <- amqp091.Confirmation{DeliveryTag: 2, Ack: true}

	err := waitConfirm(context.Background(), confirms, 2)

	require.NoError(t, err)
}

func TestWaitConfirmRejectsNegativeConfirmation(t *testing.T) {
	t.Parallel()

	confirms := make(chan amqp091.Confirmation, 1)
	confirms <- amqp091.Confirmation{DeliveryTag: 2, Ack: false}

	err := waitConfirm(context.Background(), confirms, 2)

	require.ErrorIs(t, err, ErrPublishNotConfirmed)
}

func TestAttemptsFallsBackToInitialDelivery(t *testing.T) {
	t.Parallel()

	require.Equal(t, 1, attempts(nil, false))
}

func TestDeliverySettlementRejectsSecondOperation(t *testing.T) {
	t.Parallel()

	delivery := &delivery{}
	called := 0
	err := delivery.settle(context.Background(), func() error {
		called++

		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 1, called)
	require.ErrorIs(t, delivery.settle(context.Background(), func() error { return nil }), ErrDeliverySettled)
}

func TestDeliverySettlementAllowsRetryAfterOperationFailure(t *testing.T) {
	t.Parallel()

	delivery := &delivery{}
	err := delivery.settle(context.Background(), func() error { return errors.New("broker unavailable") })

	require.Error(t, err)
	require.NoError(t, delivery.settle(context.Background(), func() error { return nil }))
}

func TestDeliverySettlementHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	delivery := &delivery{}

	err := delivery.settle(ctx, func() error { return nil })

	require.ErrorIs(t, err, context.Canceled)
}

func TestDeliveryMessagePayloadCanBeHandledByWorker(t *testing.T) {
	t.Parallel()

	message, err := deliveryMessage("work", amqp091.Delivery{Body: []byte(`{"Name":"email","Payload":"eyJpZCI6MX0="}`)})

	require.NoError(t, err)
	require.Equal(t, contract.JobMessage{Name: "email", Queue: "work", Payload: []byte(`{"id":1}`), Attempts: 1}, message)
}
