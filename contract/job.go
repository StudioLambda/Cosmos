package contract

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"
)

// RetryPolicy defines the code-owned retry policy for a [Job]. A zero value
// delegates retry decisions and timing to the driver or its configured default.
//
// MaxAttempts limits the total number of attempts when it is positive. A zero
// value uses the driver default. Backoff returns the delay before a retry for
// the given attempt number. A nil Backoff uses the driver default.
type RetryPolicy struct {
	// MaxAttempts is the total allowed attempts, including the initial attempt.
	// A value of zero uses the driver default.
	MaxAttempts int

	// Backoff returns the requested delay before the given retry attempt. A nil
	// value uses the driver default.
	Backoff func(attempt int) time.Duration
}

// Job is a self-describing unit of asynchronous application work. Its name,
// queue, retry policy, and handler are application code, not transport data.
type Job interface {
	// Name returns the stable application identifier used to resolve this job's
	// definition when handling a delivery.
	Name() string

	// Queue returns the queue to which this job is dispatched.
	Queue() string

	// RetryPolicy returns the policy for ordinary handler errors.
	RetryPolicy() RetryPolicy

	// Handle performs the job's work. A nil result acknowledges the delivery. An
	// ordinary error requests retry according to RetryPolicy. Use [RetryJob],
	// [RejectJob], or [FailJob] to request an explicit settlement.
	Handle(ctx context.Context) error
}

// JobMessage is the transport delivery envelope for a [Job]. Retry policy is
// intentionally excluded because a runtime resolves it from the job definition
// identified by Name.
type JobMessage struct {
	// ID identifies this delivery when the driver provides an identifier. The
	// dispatcher leaves ID empty because no standard-library identifier is a
	// stable transport choice; a driver assigns or exposes it on delivery.
	ID string

	// Name identifies the application job definition.
	Name string

	// Queue identifies the queue that received the job.
	Queue string

	// Payload is the JSON encoding of the job value.
	Payload []byte

	// Attempts is the number of delivery attempts when the driver provides that
	// information. A value of zero means it is unavailable.
	Attempts int
}

// JobDispatcherDriver dispatches raw [JobMessage] values. A successful
// dispatch confirms only that the driver accepted the message, not that a job
// handler processed it.
type JobDispatcherDriver interface {
	// Dispatch adds message to its named queue.
	Dispatch(ctx context.Context, message JobMessage) error
}

// JobDispatcher JSON-encodes [Job] values and dispatches their transport
// envelopes through a [JobDispatcherDriver].
type JobDispatcher struct {
	driver JobDispatcherDriver
}

// NewJobDispatcher creates a [JobDispatcher] that delegates to driver.
func NewJobDispatcher(driver JobDispatcherDriver) *JobDispatcher {
	return &JobDispatcher{driver: driver}
}

// Driver returns the underlying [JobDispatcherDriver].
func (dispatcher *JobDispatcher) Driver() JobDispatcherDriver {
	return dispatcher.driver
}

// Dispatch JSON-encodes job and sends its name and queue to the driver. It does
// not generate a message ID; see [JobMessage.ID].
func (dispatcher *JobDispatcher) Dispatch(ctx context.Context, job Job) error {
	payload, err := json.Marshal(job)

	if err != nil {
		return err
	}

	return dispatcher.driver.Dispatch(ctx, JobMessage{
		Name:    job.Name(),
		Queue:   job.Queue(),
		Payload: payload,
	})
}

// JobSettlement is the explicit delivery settlement requested by a job
// handler. A worker translates it to its broker's acknowledgement, retry, and
// rejection operations.
type JobSettlement int

const (
	// JobRetry requests redelivery. The worker applies the job retry policy
	// unless the error also implements [JobRetryAfterError].
	JobRetry JobSettlement = iota + 1

	// JobReject requests terminal rejection. A broker may dead-letter rejected
	// deliveries according to its configured policy.
	JobReject

	// JobFail requests terminal failure handling. A driver may archive or
	// dead-letter the delivery when its backend supports and is configured for
	// that behavior; it is not guaranteed by this contract.
	JobFail
)

// JobSettlementError is an inspectable job handler result. Use [errors.As] to
// inspect it while [errors.Is] and [errors.As] continue to traverse its cause.
type JobSettlementError interface {
	error

	// Settlement returns the requested delivery settlement.
	Settlement() JobSettlement

	// Unwrap returns the error that caused this outcome.
	Unwrap() error
}

// JobRetryAfterError is an optional extension of [JobSettlementError] returned by
// [RetryJobAfter]. RetryAfter is a requested delay, not a guarantee: drivers
// may be unable to schedule delayed retries exactly or at all.
type JobRetryAfterError interface {
	JobSettlementError

	// RetryAfter returns the requested delay before retrying.
	RetryAfter() time.Duration
}

type jobOutcomeError struct {
	cause      error
	settlement JobSettlement
}

func (jobOutcomeError *jobOutcomeError) Error() string {
	if jobOutcomeError.cause == nil {
		return "job outcome"
	}

	return jobOutcomeError.cause.Error()
}

func (jobOutcomeError *jobOutcomeError) Settlement() JobSettlement {
	return jobOutcomeError.settlement
}

func (jobOutcomeError *jobOutcomeError) Unwrap() error {
	return jobOutcomeError.cause
}

type jobRetryAfterError struct {
	*jobOutcomeError
	delay time.Duration
}

func (jobRetryAfterError *jobRetryAfterError) RetryAfter() time.Duration {
	return jobRetryAfterError.delay
}

// RetryJob requests another attempt according to the job's [RetryPolicy].
func RetryJob(err error) error {
	return &jobOutcomeError{cause: err, settlement: JobRetry}
}

// RetryJobAfter requests another attempt after delay. The delay is a request,
// not a guarantee; see [JobRetryAfterError].
func RetryJobAfter(err error, delay time.Duration) error {
	return &jobRetryAfterError{
		jobOutcomeError: &jobOutcomeError{cause: err, settlement: JobRetry},
		delay:           delay,
	}
}

// RejectJob requests terminal rejection. Broker dead-lettering, if any, is
// determined by the driver's backend configuration.
func RejectJob(err error) error {
	return &jobOutcomeError{cause: err, settlement: JobReject}
}

// FailJob requests terminal failure handling. See [JobFail].
func FailJob(err error) error {
	return &jobOutcomeError{cause: err, settlement: JobFail}
}

// JobResolver resolves a job name to a fresh [Job] value. A resolver must not
// reuse job values across deliveries because JSON decoding and handling may
// mutate them.
type JobResolver interface {
	// ResolveJob returns a fresh job value for name.
	ResolveJob(name string) (Job, error)
}

// JobDelivery is a broker-reserved job message. Its settlement methods hide
// transport details such as SQS receipt handles and AMQP delivery tags.
type JobDelivery interface {
	// Message returns the job message being delivered.
	Message() JobMessage

	// Acknowledge permanently completes this delivery.
	Acknowledge(ctx context.Context) error

	// Retry requests another delivery after delay. A zero delay delegates timing
	// to the driver's configured default.
	Retry(ctx context.Context, delay time.Duration) error

	// Reject permanently rejects this delivery. The driver's broker may
	// dead-letter rejected deliveries according to its configured policy.
	Reject(ctx context.Context) error

	// Fail requests terminal failure handling for this delivery. The driver's
	// broker may archive or dead-letter it according to its configured policy.
	Fail(ctx context.Context, err error) error
}

// JobDeliveryHandler handles one broker-reserved [JobDelivery]. Returning an
// error reports a worker or settlement failure to the consumer driver.
type JobDeliveryHandler func(context.Context, JobDelivery) error

// JobConsumerDriver receives deliveries from a named queue. It must not
// acknowledge a delivery before handler returns successfully.
type JobConsumerDriver interface {
	// Consume receives deliveries from queue until ctx is canceled or an error
	// occurs. The handler is responsible for settling every accepted delivery.
	Consume(ctx context.Context, queue string, handler JobDeliveryHandler) error
}

// JobWorker resolves, handles, and settles deliveries received through a
// [JobConsumerDriver].
type JobWorker struct {
	driver   JobConsumerDriver
	resolver JobResolver
}

// NewJobWorker creates a [JobWorker] that consumes through driver and resolves
// jobs through resolver.
func NewJobWorker(driver JobConsumerDriver, resolver JobResolver) *JobWorker {
	return &JobWorker{driver: driver, resolver: resolver}
}

// Driver returns the underlying [JobConsumerDriver].
func (worker *JobWorker) Driver() JobConsumerDriver {
	return worker.driver
}

// Resolver returns the underlying [JobResolver].
func (worker *JobWorker) Resolver() JobResolver {
	return worker.resolver
}

// Work consumes and handles deliveries from queue until ctx is canceled or the
// driver returns an error.
func (worker *JobWorker) Work(ctx context.Context, queue string) error {
	return worker.driver.Consume(ctx, queue, worker.handle)
}

func (worker *JobWorker) handle(ctx context.Context, delivery JobDelivery) error {
	message := delivery.Message()
	job, err := worker.resolver.ResolveJob(message.Name)

	if err != nil {
		return delivery.Fail(ctx, err)
	}

	if job.Name() != message.Name {
		return delivery.Fail(ctx, fmt.Errorf("job resolver returned %q for message %q", job.Name(), message.Name))
	}

	err = json.Unmarshal(message.Payload, job)
	if err != nil {
		return delivery.Fail(ctx, err)
	}

	err = job.Handle(ctx)
	if err == nil {
		return delivery.Acknowledge(ctx)
	}

	return worker.settle(ctx, delivery, job.RetryPolicy(), err)
}

func (worker *JobWorker) settle(ctx context.Context, delivery JobDelivery, policy RetryPolicy, err error) error {
	var settlement JobSettlementError
	if errors.As(err, &settlement) {
		switch settlement.Settlement() {
		case JobReject:
			return delivery.Reject(ctx)
		case JobFail:
			return delivery.Fail(ctx, err)
		case JobRetry:
			return worker.retry(ctx, delivery, policy, err)
		default:
			return delivery.Fail(ctx, fmt.Errorf("unknown job settlement %d: %w", settlement.Settlement(), err))
		}
	}

	return worker.retry(ctx, delivery, policy, err)
}

func (worker *JobWorker) retry(ctx context.Context, delivery JobDelivery, policy RetryPolicy, err error) error {
	message := delivery.Message()
	if policy.MaxAttempts > 0 && message.Attempts >= policy.MaxAttempts {
		return delivery.Fail(ctx, err)
	}

	var delayed JobRetryAfterError
	if errors.As(err, &delayed) {
		return delivery.Retry(ctx, delayed.RetryAfter())
	}

	if policy.Backoff == nil {
		return delivery.Retry(ctx, 0)
	}

	attempt := message.Attempts + 1
	if attempt == 1 {
		attempt = 2
	}

	return delivery.Retry(ctx, policy.Backoff(attempt))
}
