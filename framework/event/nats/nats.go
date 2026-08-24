package nats

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/studiolambda/cosmos/contract"
	core "github.com/studiolambda/cosmos/framework/event/internal/event"

	"github.com/nats-io/nats.go"
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

// NATSBroker implements [contract.EventPublisherDriver] and
// [contract.EventSubscriberDriver] using NATS messaging
// system.
// It provides a lightweight, high-performance event broker with built-in
// fan-out support and wildcard subscriptions.
// NATS handles message routing natively, making this implementation simpler
// than brokers that require manual handler tracking.
//
// Wildcard patterns: '*' matches a single dot-separated token (NATS native).
// '#' is translated to NATS subscriptions that together match zero or more
// tokens and must be the last token.
type NATSBroker struct {
	// conn is the underlying NATS connection.
	// It handles all communication with the NATS server including publishing,
	// subscribing, and maintaining the connection lifecycle.
	conn   *nats.Conn
	owned  bool
	logger *contract.Logger
}

// NATSBrokerConfig configures a NATS broker connection.
// It provides comprehensive control over connection behavior,
// authentication, and reliability features.
// All fields are optional; sensible defaults are applied when
// using NewNATSBrokerWith.
//
// WARNING: Credential fields (Username, Password, Token,
// NKeySeed) are stored as plain strings in memory for the
// lifetime of this struct. Callers should:
//  1. Always use TLS (set TLSConfig) to protect credentials
//     in transit.
//  2. Load credentials from environment variables or a secret
//     manager rather than hard-coding them.
//  3. Prefer short-lived credentials or NKey/JWT auth via
//     CredentialsFile where supported.
type NATSBrokerConfig struct {
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

	// RootCAs is a list of paths to root CA certificate files.
	// Used to verify the NATS server's certificate when using TLS.
	RootCAs []string

	// Logger records recovered handler panics. A nil logger discards records.
	Logger *contract.Logger
}

// NATSBrokerRuntime holds runtime-only NATS broker settings.
type NATSBrokerRuntime struct {
	// TLSConfig enables TLS encryption for the NATS connection.
	TLSConfig *tls.Config
}

// DefaultNATSBrokerConfig returns the default NATS broker configuration.
func DefaultNATSBrokerConfig() NATSBrokerConfig {
	return NATSBrokerConfig{
		URLs:          []string{DefaultNATSURL},
		MaxReconnects: DefaultNATSMaxReconnects,
		ReconnectWait: DefaultNATSReconnectWait,
	}
}

// FromConfiguration populates the declarative broker configuration from configuration.
func (config *NATSBrokerConfig) FromConfiguration(configuration *contract.Configuration) {
	logger := config.Logger
	*config = DefaultNATSBrokerConfig()
	config.Logger = logger
	config.URLs = configuration.GetOr("urls", config.URLs)
	config.Name = configuration.GetOr("name", config.Name)
	config.MaxReconnects = configuration.GetOr("max_reconnects", config.MaxReconnects)
	config.ReconnectWait = configuration.GetOr("reconnect_wait", config.ReconnectWait)
	config.Timeout = configuration.GetOr("timeout", config.Timeout)
	config.Username = configuration.GetOr("username", config.Username)
	config.Password = configuration.GetOr("password", config.Password)
	config.Token = configuration.GetOr("token", config.Token)
	config.NKeySeed = configuration.GetOr("nkey_seed", config.NKeySeed)
	config.CredentialsFile = configuration.GetOr("credentials_file", config.CredentialsFile)
	config.RootCAs = configuration.GetOr("root_cas", config.RootCAs)
}

// NewNATSBroker creates a new NATS broker with custom configuration.
// It applies sensible defaults for any unspecified configuration fields.
func NewNATSBroker(config NATSBrokerConfig) (*NATSBroker, error) {
	return NewNATSBrokerWith(config, NATSBrokerRuntime{})
}

// NewNATSBrokerWith creates a new NATS broker with custom configuration
// and runtime options. Returns an error if connection to the NATS server fails.
func NewNATSBrokerWith(config NATSBrokerConfig, runtime NATSBrokerRuntime) (*NATSBroker, error) {
	var opts []nats.Option

	if config.Name != "" {
		opts = append(opts, nats.Name(config.Name))
	}

	maxReconnects := DefaultNATSMaxReconnects

	if config.MaxReconnects != 0 {
		maxReconnects = config.MaxReconnects
	}

	opts = append(opts, nats.MaxReconnects(maxReconnects))

	reconnectWait := DefaultNATSReconnectWait

	if config.ReconnectWait != 0 {
		reconnectWait = config.ReconnectWait
	}

	opts = append(opts, nats.ReconnectWait(reconnectWait))

	if config.Timeout != 0 {
		opts = append(opts, nats.Timeout(config.Timeout))
	}

	if config.Username != "" && config.Password != "" {
		opts = append(opts, nats.UserInfo(config.Username, config.Password))
	}

	if config.Token != "" {
		opts = append(opts, nats.Token(config.Token))
	}

	if config.NKeySeed != "" {
		opt, err := nats.NkeyOptionFromSeed(config.NKeySeed)

		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	if config.CredentialsFile != "" {
		opts = append(opts, nats.UserCredentials(config.CredentialsFile))
	}

	if runtime.TLSConfig != nil {
		opts = append(opts, nats.Secure(runtime.TLSConfig))
	}

	if len(config.RootCAs) > 0 {
		opts = append(opts, nats.RootCAs(config.RootCAs...))
	}

	urls := config.URLs

	if len(urls) == 0 {
		urls = []string{DefaultNATSURL}
	}

	conn, err := nats.Connect(strings.Join(urls, ","), opts...)

	if err != nil {
		return nil, err
	}

	return &NATSBroker{conn: conn, owned: true, logger: eventLogger(config.Logger)}, nil
}

// NewNATSBrokerFrom creates a new NATS broker from an existing connection.
// This is useful when you need full control over connection creation or want
// to share a connection across multiple components. The caller retains
// ownership of conn.
func NewNATSBrokerFrom(conn *nats.Conn) *NATSBroker {
	return &NATSBroker{
		conn:   conn,
		logger: eventLogger(nil),
	}
}

// Publish sends raw payload bytes to all subscribers of the named event.
func (broker *NATSBroker) Publish(
	ctx context.Context,
	event string,
	payload []byte,
) error {
	if err := core.ValidateName(event); err != nil {
		return err
	}

	return broker.conn.Publish(event, payload)
}

// Subscribe registers a handler for events matching the given pattern.
// The event pattern supports portable wildcards:
//   - "*" matches a single token (e.g., "users.*.created")
//   - "#" matches zero or more trailing tokens (e.g., "users.#")
//
// The "#" wildcard is translated to an exact NATS subscription plus a '>'
// subscription because NATS '>' does not match zero tokens.
// Multiple subscribers to the same subject all receive messages (fan-out).
//
// Returns an unsubscribe function that removes this specific handler.
// The context is used only for the subscription setup, not for the handler
// lifecycle.
func (broker *NATSBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	if err := core.ValidatePattern(event); err != nil {
		return nil, err
	}

	subjects := natsSubjects(event)
	subs := make([]*nats.Subscription, 0, len(subjects))

	for _, subject := range subjects {
		sub, err := broker.conn.Subscribe(subject, func(msg *nats.Msg) {
			defer func() {
				if r := recover(); r != nil {
					broker.logger.Error("panic in nats event handler", "subject", subject, "panic", fmt.Sprint(r))
				}
			}()

			handler(msg.Data)
		})

		if err != nil {
			for _, subscribed := range subs {
				_ = subscribed.Unsubscribe()
			}

			return nil, err
		}

		subs = append(subs, sub)
	}

	return func() error {
		var errs []error

		for _, sub := range subs {
			errs = append(errs, sub.Unsubscribe())
		}

		return errors.Join(errs...)
	}, nil
}

func eventLogger(logger *contract.Logger) *contract.Logger {
	if logger == nil {
		return contract.NewLogger(nil)
	}

	return logger
}

// Ping verifies that the NATS connection is still alive.
func (broker *NATSBroker) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, nats.DefaultTimeout)
		defer cancel()
	}

	return broker.conn.FlushWithContext(ctx)
}

// Close drains and closes the NATS connection only when the broker created it.
// A broker created with [NewNATSBrokerFrom] leaves its caller-owned connection
// open.
func (broker *NATSBroker) Close() error {
	if !broker.owned {
		return nil
	}

	if err := broker.conn.Drain(); err != nil {
		return err
	}

	broker.conn.Close()

	return nil
}

// convertSubject converts event patterns to NATS subject format.
// It replaces the multi-level wildcard "#" with NATS's ">" wildcard.
// Single-level wildcards "*" are already compatible with NATS.
func convertSubject(event string) string {
	return strings.ReplaceAll(event, "#", ">")
}

// natsSubjects translates a portable pattern to one or more NATS subjects.
func natsSubjects(pattern string) []string {
	if pattern == "#" {
		return []string{">"}
	}

	if !strings.HasSuffix(pattern, "#") {
		return []string{convertSubject(pattern)}
	}

	prefix := strings.TrimSuffix(pattern, ".#")

	return []string{prefix, prefix + ".>"}
}
