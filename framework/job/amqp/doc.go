// Package amqp provides durable, at-least-once RabbitMQ work queues for jobs.
//
// It is intentionally distinct from framework/event/amqp: jobs use durable,
// competing consumers with manual acknowledgement, rather than ephemeral event
// fan-out queues.
//
// Delayed retries use durable per-delay queues with per-message TTL and dead-letter
// routing back to the work exchange. RabbitMQ TTL expiry is not an exact scheduler,
// particularly when older messages block a retry queue. Confirmed retry publishing
// happens before acknowledging the original delivery, so failures in that window can
// produce duplicates but do not intentionally lose work.
package amqp
