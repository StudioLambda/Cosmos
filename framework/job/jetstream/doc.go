// Package jetstream provides a NATS JetStream-backed job driver with durable,
// at-least-once delivery. It requires JetStream streams and durable pull
// consumers; Core NATS alone is not supported.
//
// [New] creates and owns its NATS connection, which [Driver.Close] gracefully
// drains. Callers retain ownership of a connection passed to [NewFrom]; its
// Close method does not drain or close that connection. By default, streams and
// consumers must be provisioned operationally. Set [Config.Provision] only when
// the application is permitted to create missing JetStream resources.
//
// Reject acknowledges the original message because JetStream has no universal
// reject or dead-letter operation. Fail publishes an application-level failure
// envelope to [Config.FailureSubject], when configured, before acknowledging
// the original message. A failure subject is not a JetStream DLQ guarantee.
package jetstream
