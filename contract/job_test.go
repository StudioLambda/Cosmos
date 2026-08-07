package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"
	contractmock "github.com/studiolambda/cosmos/contract/mock"

	"github.com/stretchr/testify/require"
)

type emailJob struct {
	ID  string `json:"id"`
	err error
}

func (emailJob) Name() string {
	return "emails.send"
}

func (emailJob) Queue() string {
	return "emails"
}

func (emailJob) RetryPolicy() contract.RetryPolicy {
	return contract.RetryPolicy{
		MaxAttempts: 3,
		Backoff: func(attempt int) time.Duration {
			return time.Duration(attempt) * time.Minute
		},
	}
}

func (job emailJob) Handle(context.Context) error {
	return job.err
}

type jobResolver func(string) (contract.Job, error)

func (resolver jobResolver) ResolveJob(name string) (contract.Job, error) {
	return resolver(name)
}

type jobConsumerDriver struct {
	delivery contract.JobDelivery
}

func (driver jobConsumerDriver) Consume(ctx context.Context, queue string, handler contract.JobDeliveryHandler) error {
	return handler(ctx, driver.delivery)
}

type jobDelivery struct {
	message      contract.JobMessage
	acknowledged bool
	retried      bool
	retryDelay   time.Duration
	rejected     bool
	failed       error
}

func (delivery *jobDelivery) Message() contract.JobMessage {
	return delivery.message
}

func (delivery *jobDelivery) Acknowledge(context.Context) error {
	delivery.acknowledged = true

	return nil
}

func (delivery *jobDelivery) Retry(_ context.Context, delay time.Duration) error {
	delivery.retried = true
	delivery.retryDelay = delay

	return nil
}

func (delivery *jobDelivery) Reject(context.Context) error {
	delivery.rejected = true

	return nil
}

func (delivery *jobDelivery) Fail(_ context.Context, err error) error {
	delivery.failed = err

	return nil
}

func TestJobDispatcherDispatchesSerializedJob(t *testing.T) {
	t.Parallel()

	driver := contractmock.NewJobDispatcherDriverMock(t)
	dispatcher := contract.NewJobDispatcher(driver)
	driver.EXPECT().Dispatch(context.Background(), contract.JobMessage{
		Name:    "emails.send",
		Queue:   "emails",
		Payload: []byte(`{"id":"email-1"}`),
	}).Return(nil)

	err := dispatcher.Dispatch(context.Background(), emailJob{ID: "email-1"})

	require.NoError(t, err)
	require.Same(t, driver, dispatcher.Driver())
}

func TestJobInterfaceForwardsApplicationValues(t *testing.T) {
	t.Parallel()

	var job contract.Job = emailJob{ID: "email-1"}

	require.Equal(t, "emails.send", job.Name())
	require.Equal(t, "emails", job.Queue())
	require.Equal(t, 3, job.RetryPolicy().MaxAttempts)
	require.NoError(t, job.Handle(context.Background()))
}

func TestJobWorkerAcknowledgesSuccessfulJob(t *testing.T) {
	t.Parallel()

	delivery := &jobDelivery{message: contract.JobMessage{
		Name:    "emails.send",
		Payload: []byte(`{"id":"email-1"}`),
	}}
	worker := contract.NewJobWorker(jobConsumerDriver{delivery: delivery}, jobResolver(func(string) (contract.Job, error) {
		return &emailJob{}, nil
	}))

	err := worker.Work(context.Background(), "emails")

	require.NoError(t, err)
	require.True(t, delivery.acknowledged)
	require.Nil(t, delivery.failed)
}

func TestJobWorkerRetriesOrdinaryErrorUsingPolicyBackoff(t *testing.T) {
	t.Parallel()

	cause := errors.New("service unavailable")
	delivery := &jobDelivery{message: contract.JobMessage{
		Name:     "emails.send",
		Payload:  []byte(`{"id":"email-1"}`),
		Attempts: 1,
	}}
	worker := contract.NewJobWorker(jobConsumerDriver{delivery: delivery}, jobResolver(func(string) (contract.Job, error) {
		return &emailJob{err: cause}, nil
	}))

	err := worker.Work(context.Background(), "emails")

	require.NoError(t, err)
	require.True(t, delivery.retried)
	require.Equal(t, 2*time.Minute, delivery.retryDelay)
}

func TestJobWorkerRetriesExplicitDelay(t *testing.T) {
	t.Parallel()

	cause := errors.New("rate limited")
	delivery := &jobDelivery{message: contract.JobMessage{
		Name:    "emails.send",
		Payload: []byte(`{"id":"email-1"}`),
	}}
	worker := contract.NewJobWorker(jobConsumerDriver{delivery: delivery}, jobResolver(func(string) (contract.Job, error) {
		return &emailJob{err: contract.RetryJobAfter(cause, time.Minute)}, nil
	}))

	err := worker.Work(context.Background(), "emails")

	require.NoError(t, err)
	require.True(t, delivery.retried)
	require.Equal(t, time.Minute, delivery.retryDelay)
}

func TestJobWorkerRejectsRejectedJob(t *testing.T) {
	t.Parallel()

	delivery := &jobDelivery{message: contract.JobMessage{
		Name:    "emails.send",
		Payload: []byte(`{"id":"email-1"}`),
	}}
	worker := contract.NewJobWorker(jobConsumerDriver{delivery: delivery}, jobResolver(func(string) (contract.Job, error) {
		return &emailJob{err: contract.RejectJob(errors.New("invalid recipient"))}, nil
	}))

	err := worker.Work(context.Background(), "emails")

	require.NoError(t, err)
	require.True(t, delivery.rejected)
}

func TestJobWorkerFailsExhaustedJob(t *testing.T) {
	t.Parallel()

	cause := errors.New("service unavailable")
	delivery := &jobDelivery{message: contract.JobMessage{
		Name:     "emails.send",
		Payload:  []byte(`{"id":"email-1"}`),
		Attempts: 3,
	}}
	worker := contract.NewJobWorker(jobConsumerDriver{delivery: delivery}, jobResolver(func(string) (contract.Job, error) {
		return &emailJob{err: cause}, nil
	}))

	err := worker.Work(context.Background(), "emails")

	require.NoError(t, err)
	require.ErrorIs(t, delivery.failed, cause)
}

func TestJobWorkerFailsUnresolvableJob(t *testing.T) {
	t.Parallel()

	cause := errors.New("unknown job")
	delivery := &jobDelivery{message: contract.JobMessage{Name: "unknown"}}
	worker := contract.NewJobWorker(jobConsumerDriver{delivery: delivery}, jobResolver(func(string) (contract.Job, error) {
		return nil, cause
	}))

	err := worker.Work(context.Background(), "emails")

	require.NoError(t, err)
	require.ErrorIs(t, delivery.failed, cause)
}

func TestRetryJobPreservesCauseAndSettlement(t *testing.T) {
	t.Parallel()

	cause := errors.New("service unavailable")
	err := contract.RetryJob(cause)
	var settlement contract.JobSettlementError

	require.ErrorAs(t, err, &settlement)
	require.Equal(t, contract.JobRetry, settlement.Settlement())
	require.ErrorIs(t, err, cause)
}

func TestRejectJobPreservesCauseAndSettlement(t *testing.T) {
	t.Parallel()

	cause := errors.New("invalid recipient")
	err := contract.RejectJob(cause)
	var settlement contract.JobSettlementError

	require.ErrorAs(t, err, &settlement)
	require.Equal(t, contract.JobReject, settlement.Settlement())
	require.ErrorIs(t, err, cause)
}

func TestFailJobPreservesCauseAndSettlement(t *testing.T) {
	t.Parallel()

	cause := errors.New("invalid template")
	err := contract.FailJob(cause)
	var settlement contract.JobSettlementError

	require.ErrorAs(t, err, &settlement)
	require.Equal(t, contract.JobFail, settlement.Settlement())
	require.ErrorIs(t, err, cause)
}

func TestRetryJobAfterPreservesCauseSettlementAndDelay(t *testing.T) {
	t.Parallel()

	cause := errors.New("rate limited")
	err := contract.RetryJobAfter(cause, time.Minute)
	var settlement contract.JobSettlementError
	var delayed contract.JobRetryAfterError

	require.ErrorAs(t, err, &settlement)
	require.Equal(t, contract.JobRetry, settlement.Settlement())
	require.ErrorAs(t, err, &delayed)
	require.Equal(t, time.Minute, delayed.RetryAfter())
	require.ErrorIs(t, err, cause)
}
