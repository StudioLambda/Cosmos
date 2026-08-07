package sns

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/stretchr/testify/require"
)

func TestNewFromRejectsNilClient(t *testing.T) {
	t.Parallel()

	_, err := NewFrom(nil, SNSPublisherConfig{Topics: map[string]string{"users.created": "arn:aws:sns:example"}})

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestNewRejectsMissingTopics(t *testing.T) {
	t.Parallel()

	_, err := New(SNSPublisherConfig{})

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestNewRejectsBlankTopicMapping(t *testing.T) {
	t.Parallel()

	_, err := New(SNSPublisherConfig{Topics: map[string]string{" ": " "}})

	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestDriverPublishRejectsUnknownEvent(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	driver := newDriver(t, client, map[string]string{"users.created": "arn:aws:sns:example:users"})

	err := driver.Publish(context.Background(), "users.deleted", nil)

	require.ErrorIs(t, err, ErrUnknownEvent)
	require.Empty(t, client.publishInputs())
}

func TestDriverPublishReturnsCanceledContextWithoutCallingClient(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	driver := newDriver(t, client, map[string]string{"users.created": "arn:aws:sns:example:users"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := driver.Publish(ctx, "users.created", nil)

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, client.publishInputs())
}

func TestDriverPublishSendsRawPayloadToMappedTopic(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	driver := newDriver(t, client, map[string]string{"users.created": "arn:aws:sns:example:users"})
	payload := []byte{0x00, 0xff, '{', '}', 0x00}

	err := driver.Publish(context.Background(), "users.created", payload)

	require.NoError(t, err)
	require.Len(t, client.publishInputs(), 1)
	input := client.publishInputs()[0]
	require.Equal(t, "arn:aws:sns:example:users", aws.ToString(input.TopicArn))
	require.Equal(t, string(payload), aws.ToString(input.Message))
	require.Nil(t, input.Subject)
}

func TestDriverPublishWrapsClientError(t *testing.T) {
	t.Parallel()

	publishErr := errors.New("unavailable")
	client := &fakeClient{publishErr: publishErr}
	driver := newDriver(t, client, map[string]string{"users.created": "arn:aws:sns:example:users"})

	err := driver.Publish(context.Background(), "users.created", nil)

	require.ErrorIs(t, err, publishErr)
}

func TestNewClonesTopicMappings(t *testing.T) {
	t.Parallel()

	client := &fakeClient{}
	topics := map[string]string{"users.created": "arn:aws:sns:example:original"}
	driver := newDriver(t, client, topics)
	topics["users.created"] = "arn:aws:sns:example:changed"

	err := driver.Publish(context.Background(), "users.created", nil)

	require.NoError(t, err)
	require.Len(t, client.publishInputs(), 1)
	require.Equal(t, "arn:aws:sns:example:original", aws.ToString(client.publishInputs()[0].TopicArn))
}

func newDriver(t *testing.T, client *fakeClient, topics map[string]string) *Driver {
	t.Helper()

	driver, err := newFrom(client, SNSPublisherConfig{Topics: topics})
	require.NoError(t, err)

	return driver
}

type fakeClient struct {
	mu         sync.Mutex
	publish    []*awssns.PublishInput
	publishErr error
}

func (client *fakeClient) Publish(_ context.Context, input *awssns.PublishInput, _ ...func(*awssns.Options)) (*awssns.PublishOutput, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.publish = append(client.publish, input)

	return &awssns.PublishOutput{}, client.publishErr
}

func (client *fakeClient) publishInputs() []*awssns.PublishInput {
	client.mu.Lock()
	defer client.mu.Unlock()

	return append([]*awssns.PublishInput(nil), client.publish...)
}
