package sqs

import (
	"context"
	"encoding/json/v2"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/studiolambda/cosmos/contract"

	"github.com/stretchr/testify/require"
)

func TestDriverDispatchSendsJSONMessageToMappedQueue(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	driver := newDriver(t, client)
	message := contract.JobMessage{Name: "email", Queue: "work", Payload: []byte(`{"id":1}`)}

	err := driver.Dispatch(context.Background(), message)

	require.NoError(t, err)
	require.Len(t, client.sendInputs(), 1)
	input := client.sendInputs()[0]
	require.Equal(t, "https://sqs.example/work", aws.ToString(input.QueueUrl))
	var sent contract.JobMessage
	require.NoError(t, json.Unmarshal([]byte(aws.ToString(input.MessageBody)), &sent))
	require.Equal(t, message, sent)
}

func TestDriverConsumeReceivesAndAcknowledgesDelivery(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(contract.JobMessage{Name: "email", Queue: "untrusted", Payload: []byte(`{"id":1}`)})
	require.NoError(t, err)
	client := &fakeClient{receiveOutputs: []*awssqs.ReceiveMessageOutput{{Messages: []types.Message{{
		Body:          aws.String(string(body)),
		MessageId:     aws.String("message-1"),
		ReceiptHandle: aws.String("receipt-1"),
		Attributes:    map[string]string{"ApproximateReceiveCount": "2"},
	}}}}}
	driver := newDriver(t, client)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	handled := make(chan contract.JobMessage, 1)
	consumed := make(chan error, 1)
	go func() {
		consumed <- driver.Consume(ctx, "work", func(ctx context.Context, delivery contract.JobDelivery) error {
			handled <- delivery.Message()

			return delivery.Acknowledge(ctx)
		})
	}()

	message := <-handled
	cancel()
	require.ErrorIs(t, <-consumed, context.Canceled)
	require.Equal(t, "work", message.Queue)
	require.Equal(t, "message-1", message.ID)
	require.Equal(t, 2, message.Attempts)
	require.Len(t, client.deleteInputs(), 1)
	require.Equal(t, "receipt-1", aws.ToString(client.deleteInputs()[0].ReceiptHandle))
	require.NotEmpty(t, client.receiveInputs())
	receive := client.receiveInputs()[0]
	require.Equal(t, "https://sqs.example/work", aws.ToString(receive.QueueUrl))
	require.Equal(t, int32(20), receive.WaitTimeSeconds)
	require.Equal(t, int32(1), receive.MaxNumberOfMessages)
	require.Equal(t, []types.MessageSystemAttributeName{types.MessageSystemAttributeNameApproximateReceiveCount}, receive.MessageSystemAttributeNames)
}

func TestDeliveryRetryChangesVisibility(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Retry(ctx, 1500*time.Millisecond)
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, client.visibilityInputs(), 1)
	require.Equal(t, int32(2), client.visibilityInputs()[0].VisibilityTimeout)
}

func TestDeliveryRetryWithZeroDelayMakesMessageImmediatelyVisible(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Retry(ctx, 0)
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, client.visibilityInputs(), 1)
	require.Equal(t, int32(0), client.visibilityInputs()[0].VisibilityTimeout)
}

func TestDeliveryRetryRejectsDelayAboveSQSMaximum(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Retry(ctx, 15*time.Minute+time.Second)
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, client.visibilityInputs())
}

func TestDeliveryRejectDeletesMessage(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Reject(ctx)
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, client.deleteInputs(), 1)
}

func TestDriverConsumeLeavesDeliveryWhenHandlerReturnsError(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(context.Context, contract.JobDelivery) error {
		return errors.New("handler failed")
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, client.deleteInputs())
	require.Empty(t, client.visibilityInputs())
}

func TestDeliveryFailPublishesFailureThenDeletesMessage(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newFailureDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Fail(ctx, errors.New("failed"))
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, client.sendInputs(), 1)
	require.Equal(t, "https://sqs.example/failures", aws.ToString(client.sendInputs()[0].QueueUrl))
	require.Len(t, client.deleteInputs(), 1)
}

func TestDeliveryFailWithoutFailureQueueDeletesMessage(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	driver := newDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Fail(ctx, errors.New("failed"))
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, client.sendInputs())
	require.Len(t, client.deleteInputs(), 1)
}

func TestDeliveryFailLeavesMessageWhenFailurePublishingFails(t *testing.T) {
	t.Parallel()

	client := receivedClient(t)
	client.sendErr = errors.New("unavailable")
	driver := newFailureDriver(t, client)
	err := driver.Consume(context.Background(), "work", func(ctx context.Context, delivery contract.JobDelivery) error {
		return delivery.Fail(ctx, errors.New("failed"))
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, client.deleteInputs())
}

func TestDriverConsumeFailsMalformedMessageAndDeletesIt(t *testing.T) {
	t.Parallel()

	client := &fakeClient{receiveOutputs: []*awssqs.ReceiveMessageOutput{{Messages: []types.Message{{
		Body:          aws.String("not json"),
		ReceiptHandle: aws.String("receipt-1"),
	}}}}}
	driver := newFailureDriver(t, client)

	err := driver.Consume(context.Background(), "work", func(context.Context, contract.JobDelivery) error {
		return errors.New("handler should not be called")
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, client.sendInputs(), 1)
	require.Len(t, client.deleteInputs(), 1)
}

func TestDriverRejectsUnknownQueueAndInvalidConfiguration(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	_, err := New(SQSDriverConfig{})
	require.ErrorIs(t, err, ErrInvalidConfig)
	driver := newDriver(t, client)
	require.ErrorIs(t, driver.Dispatch(context.Background(), contract.JobMessage{Queue: "unknown"}), ErrUnknownQueue)
	require.ErrorIs(t, driver.Consume(context.Background(), "unknown", func(context.Context, contract.JobDelivery) error { return nil }), ErrUnknownQueue)
}

func TestDriverConsumeReturnsCanceledContext(t *testing.T) {
	t.Parallel()

	driver := newDriver(t, &fakeClient{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := driver.Consume(ctx, "work", func(context.Context, contract.JobDelivery) error { return nil })

	require.ErrorIs(t, err, context.Canceled)
}

func newDriver(t *testing.T, client *fakeClient) *Driver {
	t.Helper()

	config := DefaultSQSDriverConfig()
	config.Queues = map[string]string{"work": "https://sqs.example/work"}
	driver, err := newFrom(client, config)
	require.NoError(t, err)

	return driver
}

func newFailureDriver(t *testing.T, client *fakeClient) *Driver {
	t.Helper()

	config := DefaultSQSDriverConfig()
	config.Queues = map[string]string{"work": "https://sqs.example/work"}
	config.FailureQueueURL = "https://sqs.example/failures"
	driver, err := newFrom(client, config)
	require.NoError(t, err)

	return driver
}

func receivedClient(t *testing.T) *fakeClient {
	t.Helper()

	body, err := json.Marshal(contract.JobMessage{Name: "email", Payload: []byte(`{"id":1}`)})
	require.NoError(t, err)

	return &fakeClient{receiveOutputs: []*awssqs.ReceiveMessageOutput{{Messages: []types.Message{{
		Body:          aws.String(string(body)),
		ReceiptHandle: aws.String("receipt-1"),
	}}}}}
}

type fakeClient struct {
	mu             sync.Mutex
	receiveOutputs []*awssqs.ReceiveMessageOutput
	receive        []*awssqs.ReceiveMessageInput
	send           []*awssqs.SendMessageInput
	deleted        []*awssqs.DeleteMessageInput
	visibility     []*awssqs.ChangeMessageVisibilityInput
	sendErr        error
}

func (client *fakeClient) SendMessage(_ context.Context, input *awssqs.SendMessageInput, _ ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.send = append(client.send, input)

	return &awssqs.SendMessageOutput{}, client.sendErr
}

func (client *fakeClient) ReceiveMessage(_ context.Context, input *awssqs.ReceiveMessageInput, _ ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error) {
	client.mu.Lock()
	client.receive = append(client.receive, input)
	if len(client.receiveOutputs) > 0 {
		output := client.receiveOutputs[0]
		client.receiveOutputs = client.receiveOutputs[1:]
		client.mu.Unlock()

		return output, nil
	}
	client.mu.Unlock()

	return nil, context.Canceled
}

func (client *fakeClient) DeleteMessage(_ context.Context, input *awssqs.DeleteMessageInput, _ ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.deleted = append(client.deleted, input)

	return &awssqs.DeleteMessageOutput{}, nil
}

func (client *fakeClient) ChangeMessageVisibility(_ context.Context, input *awssqs.ChangeMessageVisibilityInput, _ ...func(*awssqs.Options)) (*awssqs.ChangeMessageVisibilityOutput, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.visibility = append(client.visibility, input)

	return &awssqs.ChangeMessageVisibilityOutput{}, nil
}

func (client *fakeClient) receiveInputs() []*awssqs.ReceiveMessageInput {
	client.mu.Lock()
	defer client.mu.Unlock()

	return append([]*awssqs.ReceiveMessageInput(nil), client.receive...)
}

func (client *fakeClient) sendInputs() []*awssqs.SendMessageInput {
	client.mu.Lock()
	defer client.mu.Unlock()

	return append([]*awssqs.SendMessageInput(nil), client.send...)
}

func (client *fakeClient) deleteInputs() []*awssqs.DeleteMessageInput {
	client.mu.Lock()
	defer client.mu.Unlock()

	return append([]*awssqs.DeleteMessageInput(nil), client.deleted...)
}

func (client *fakeClient) visibilityInputs() []*awssqs.ChangeMessageVisibilityInput {
	client.mu.Lock()
	defer client.mu.Unlock()

	return append([]*awssqs.ChangeMessageVisibilityInput(nil), client.visibility...)
}
