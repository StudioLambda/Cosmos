package jetstream

import (
	"context"
	"encoding/base64"
	"encoding/json/v2"
	"errors"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestNormalizeConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	config, err := normalizeConfig(Config{Stream: "JOBS", Queues: map[string]QueueConfig{"work": {Subject: "jobs.work", Durable: "work"}}})

	require.NoError(t, err)
	require.Equal(t, 1, config.Batch)
	require.Equal(t, 1, config.MaxAckPending)
	require.Equal(t, 30*time.Second, config.AckWait)
}

func TestNormalizeConfigRejectsMissingDurable(t *testing.T) {
	t.Parallel()

	_, err := normalizeConfig(Config{Stream: "JOBS", Queues: map[string]QueueConfig{"work": {Subject: "jobs.work"}}})

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestNormalizeConfigRejectsDuplicateSubject(t *testing.T) {
	t.Parallel()

	_, err := normalizeConfig(Config{Stream: "JOBS", Queues: map[string]QueueConfig{
		"first":  {Subject: "jobs.work", Durable: "first"},
		"second": {Subject: "jobs.work", Durable: "second"},
	}})

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestValidateConsumerRejectsPushConsumer(t *testing.T) {
	t.Parallel()

	err := validateConsumer(Config{Batch: 1, MaxAckPending: 1, AckWait: time.Second}, QueueConfig{Subject: "jobs.work", Durable: "work"}, nats.ConsumerConfig{
		Durable:        "work",
		DeliverSubject: "deliver.work",
		AckPolicy:      nats.AckExplicitPolicy,
		FilterSubject:  "jobs.work",
	})

	require.ErrorIs(t, err, ErrResourcesNotProvisioned)
}

func TestValidateStreamRejectsMissingQueueSubject(t *testing.T) {
	t.Parallel()

	err := validateStream(Config{Stream: "JOBS", Queues: map[string]QueueConfig{"work": {Subject: "jobs.work", Durable: "work"}}}, nats.StreamConfig{Subjects: []string{"jobs.other"}})

	require.ErrorIs(t, err, ErrResourcesNotProvisioned)
}

func TestMarshalMessagePreservesFullEnvelope(t *testing.T) {
	t.Parallel()

	want := contract.JobMessage{ID: "job-1", Name: "email", Queue: "work", Payload: []byte(`{"id":1}`), Attempts: 9}
	body, err := marshalMessage(want)

	require.NoError(t, err)
	var got contract.JobMessage
	require.NoError(t, json.Unmarshal(body, &got))
	require.Equal(t, want, got)
}

func TestNewMessageIDIsOpaqueAndNonEmpty(t *testing.T) {
	t.Parallel()

	id, err := newMessageID()

	require.NoError(t, err)
	require.NotEmpty(t, id)
	_, err = base64.RawURLEncoding.DecodeString(id)
	require.NoError(t, err)
}

func TestDecodeMessageRejectsMalformedEnvelope(t *testing.T) {
	t.Parallel()

	_, err := decodeMessage("work", &nats.Msg{Data: []byte("not JSON")})

	require.ErrorIs(t, err, ErrInvalidDelivery)
}

func TestDeliveryRetryRejectsNegativeDelay(t *testing.T) {
	t.Parallel()

	delivery := newDelivery(&Driver{}, &nats.Msg{}, contract.JobMessage{})

	err := delivery.Retry(context.Background(), -time.Second)

	require.ErrorIs(t, err, ErrUnsupportedRetryDelay)
}

func TestDeliverySettlementRejectsSecondOperation(t *testing.T) {
	t.Parallel()

	delivery := newDelivery(&Driver{}, &nats.Msg{}, contract.JobMessage{})
	require.NotNil(t, delivery)
	require.NoError(t, delivery.settle(context.Background(), func() error { return nil }))

	err := delivery.settle(context.Background(), func() error { return nil })

	require.ErrorIs(t, err, ErrDeliverySettled)
}

func TestFailureMessageIncludesSafeEnvelope(t *testing.T) {
	t.Parallel()

	message := failureMessage{Message: contract.JobMessage{ID: "job-1", Queue: "work", Payload: []byte{}}, Error: errorText(errors.New("failed")), RawMessage: "not JSON"}
	body, err := json.Marshal(message)

	require.NoError(t, err)
	var got failureMessage
	require.NoError(t, json.Unmarshal(body, &got))
	require.Equal(t, message, got)
}
