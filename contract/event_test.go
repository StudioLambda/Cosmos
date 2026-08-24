package contract_test

import (
	"context"
	"testing"

	"github.com/studiolambda/cosmos/contract"
	contractmock "github.com/studiolambda/cosmos/contract/mock"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventPublisherPublishesJSON(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewEventPublisherDriverMock(t)
	publisher := contract.NewEventPublisher(driver)
	driver.EXPECT().Publish(context.Background(), "users.created", []byte(`{"id":1}`)).Return(nil)

	err := publisher.Publish(context.Background(), "users.created", struct {
		ID int `json:"id"`
	}{ID: 1})

	require.NoError(t, err)
	require.Same(t, driver, publisher.Driver())
}

func TestEventSubscriberDecodesJSON(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewEventSubscriberDriverMock(t)
	subscriber := contract.NewEventSubscriber(driver)
	unsubscribe := func() error { return nil }

	driver.EXPECT().Subscribe(context.Background(), "users.created", mock.Anything).RunAndReturn(
		func(_ context.Context, _ string, handler contract.EventHandler) (contract.EventUnsubscribeFunc, error) {
			handler([]byte(`{"id":1}`))

			return unsubscribe, nil
		},
	)

	var received int
	actualUnsubscribe, err := subscriber.Subscribe(context.Background(), "users.created", func(decode contract.EventDecoder[struct {
		ID int `json:"id"`
	}]) {
		message, err := decode()
		require.NoError(t, err)

		received = message.ID
	})

	require.NoError(t, err)
	require.Equal(t, 1, received)
	require.NoError(t, actualUnsubscribe())
	require.Same(t, driver, subscriber.Driver())
}
