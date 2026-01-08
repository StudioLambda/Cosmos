package event

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/studiolambda/cosmos/contract"
)

const (
	// DefaultNATSURL is the default connection URL for NATS server.
	// It points to a local NATS server running on the standard port.
	DefaultNATSURL = nats.DefaultURL

	// DefaultNATSMaxReconnects is the default maximum number of reconnect
	// attempts.
	// A value of -1 allows unlimited reconnection attempts.
	DefaultNATSMaxReconnects = -1

	// DefaultNATSReconnectWait is the default time to wait between reconnect
	// attempts.
	// This provides a reasonable backoff when the NATS server is temporarily
	// unavailable.
	DefaultNATSReconnectWait = 2 * time.Second
)

// NATSBroker implements the EventBroker interface using NATS messaging
// system.
// It provides a lightweight, high-performance event broker with built-in
// fan-out support and wildcard subscriptions.
// NATS handles message routing natively, making this implementation simpler
// than brokers that require manual handler tracking.
type NATSBroker struct {
	// conn is the underlying NATS connection.
	// It handles all communication with the NATS server including publishing,
	// subscribing, and maintaining the connection lifecycle.
	conn *nats.Conn
}

// NATSBrokerOptions configures a NATS broker connection.
// It provides comprehensive control over connection behavior, authentication,
// and reliability features.
// All fields are optional; sensible defaults are applied when using
// NewNATSBrokerWith.
type NATSBrokerOptions struct {
	// URLs is a list of NATS server URLs to connect to.
	// Multiple URLs enable automatic failover in clustered deployments.
	// If empty, defaults to DefaultNATSURL.
	URLs []string

	// Name identifies this client connection in NATS server logs and
	// monitoring.
	// Useful for debugging and tracing connection issues.
	Name string

	// MaxReconnects is the maximum number of reconnection attempts.
	// Use -1 for unlimited reconnects (default), 0 to disable reconnection.
	MaxReconnects int

	// ReconnectWait is the time to wait between reconnection attempts.
	// Defaults to DefaultNATSReconnectWait (2 seconds).
	ReconnectWait time.Duration

	// Timeout is the connection timeout for initial connection and
	// operations.
	// If zero, NATS uses its default timeout.
	Timeout time.Duration

	// Username is the username for basic authentication.
	// Used in combination with Password when the NATS server requires auth.
	Username string

	// Password is the password for basic authentication.
	// Used in combination with Username when the NATS server requires auth.
	Password string

	// Token is a bearer token for token-based authentication.
	// Alternative to username/password authentication.
	Token string

	// NKeySeed is the seed for NKey authentication.
	// NKey provides cryptographic authentication without transmitting secrets.
	NKeySeed string

	// CredentialsFile is the path to a credentials file containing JWT and
	// NKey.
	// This is the recommended authentication method for production
	// deployments.
	CredentialsFile string

	// TLSConfig enables TLS encryption for the NATS connection.
	// When set, all communication with the NATS server is encrypted.
	TLSConfig *tls.Config

	// RootCAs is a list of paths to root CA certificate files.
	// Used to verify the NATS server's certificate when using TLS.
	RootCAs []string
}

// NewNATSBroker creates a new NATS broker connected to the specified URL.
// It applies sensible defaults for reconnection behavior.
// This constructor is suitable for simple use cases with a single NATS
// server.
//
// For clustered deployments or custom authentication, use
// NewNATSBrokerWith instead.
func NewNATSBroker(url string) (*NATSBroker, error) {
	return NewNATSBrokerWith(&NATSBrokerOptions{
		URLs: []string{url},
	})
}

// NewNATSBrokerWith creates a new NATS broker with custom configuration.
// It provides full control over connection behavior, authentication, and
// reliability.
// Applies sensible defaults for any unspecified options.
//
// Returns an error if connection to the NATS server fails.
func NewNATSBrokerWith(options *NATSBrokerOptions) (*NATSBroker, error) {
	opts := []nats.Option{}

	if options.Name != "" {
		opts = append(opts, nats.Name(options.Name))
	}

	maxReconnects := DefaultNATSMaxReconnects

	if options.MaxReconnects != 0 {
		maxReconnects = options.MaxReconnects
	}

	opts = append(opts, nats.MaxReconnects(maxReconnects))

	reconnectWait := DefaultNATSReconnectWait

	if options.ReconnectWait != 0 {
		reconnectWait = options.ReconnectWait
	}

	opts = append(opts, nats.ReconnectWait(reconnectWait))

	if options.Timeout != 0 {
		opts = append(opts, nats.Timeout(options.Timeout))
	}

	if options.Username != "" && options.Password != "" {
		opts = append(opts, nats.UserInfo(options.Username, options.Password))
	}

	if options.Token != "" {
		opts = append(opts, nats.Token(options.Token))
	}

	if options.NKeySeed != "" {
		opt, err := nats.NkeyOptionFromSeed(options.NKeySeed)

		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	if options.CredentialsFile != "" {
		opts = append(opts, nats.UserCredentials(options.CredentialsFile))
	}

	if options.TLSConfig != nil {
		opts = append(opts, nats.Secure(options.TLSConfig))
	}

	if len(options.RootCAs) > 0 {
		opts = append(opts, nats.RootCAs(options.RootCAs...))
	}

	urls := options.URLs

	if len(urls) == 0 {
		urls = []string{DefaultNATSURL}
	}

	conn, err := nats.Connect(strings.Join(urls, ","), opts...)

	if err != nil {
		return nil, err
	}

	return NewNATSBrokerFrom(conn), nil
}

// NewNATSBrokerFrom creates a new NATS broker from an existing connection.
// This is useful when you need full control over connection creation or want
// to share a connection across multiple components.
// The broker takes ownership of the connection and will close it when Close
// is called.
func NewNATSBrokerFrom(conn *nats.Conn) *NATSBroker {
	return &NATSBroker{
		conn: conn,
	}
}

// Publish sends an event with the given payload to all subscribers.
// The payload is JSON-encoded before transmission.
// The event name is used as the NATS subject.
//
// Returns an error if JSON encoding fails or if the publish operation fails.
// The context is used for operation timeout and cancellation.
func (b *NATSBroker) Publish(
	ctx context.Context,
	event string,
	payload any,
) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	return b.conn.Publish(event, encoded)
}

// Subscribe registers a handler for events matching the given pattern.
// The event pattern supports NATS wildcards:
//   - "*" matches a single token (e.g., "users.*.created")
//   - ">" matches multiple tokens (e.g., "users.>")
//
// The "#" wildcard from other brokers is automatically converted to ">".
// Multiple subscribers to the same subject all receive messages (fan-out).
//
// Returns an unsubscribe function that removes this specific handler.
// The context is used only for the subscription setup, not for the handler
// lifecycle.
func (b *NATSBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	subject := b.convertSubject(event)

	sub, err := b.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(func(dest any) error {
			return json.Unmarshal(msg.Data, dest)
		})
	})

	if err != nil {
		return nil, err
	}

	return func() error {
		return sub.Unsubscribe()
	}, nil
}

// Close gracefully shuts down the NATS connection.
// It drains all pending messages before closing, ensuring no messages are
// lost.
// After Close is called, the broker cannot be reused.
func (b *NATSBroker) Close() error {
	if err := b.conn.Drain(); err != nil {
		return err
	}

	b.conn.Close()

	return nil
}

// convertSubject converts event patterns to NATS subject format.
// It replaces the multi-level wildcard "#" with NATS's ">" wildcard.
// Single-level wildcards "*" are already compatible with NATS.
func (b *NATSBroker) convertSubject(event string) string {
	return strings.ReplaceAll(event, "#", ">")
}
