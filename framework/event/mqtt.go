package event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/studiolambda/cosmos/contract"
)

// MQTTBroker implements the EventBroker interface using MQTT v5
// protocol for publish/subscribe messaging. It uses the Eclipse
// Paho Go client with automatic reconnection support and always
// operates with clean sessions for simplicity.
//
// Event names are automatically converted to MQTT topic format
// where dots become slashes and asterisks become plus signs for
// single-level wildcard matching. Hash symbols remain unchanged
// for multi-level wildcards.
//
// The broker supports fan-out messaging where multiple handlers
// can subscribe to the same topic and all will receive messages.
type MQTTBroker struct {
	// client is the autopaho connection manager with auto-reconnection.
	client *autopaho.ConnectionManager

	// qos is the quality of service level for publish and subscribe.
	qos byte

	// mu protects the handlers and subscriptions maps during
	// concurrent access.
	mu sync.RWMutex

	// handlers stores all event handlers keyed by topic and
	// handler ID for supporting multiple handlers per topic.
	handlers map[string]map[string]contract.EventHandler

	// subscriptions tracks active MQTT broker subscriptions
	// by topic to enable proper cleanup and unsubscribe.
	subscriptions map[string]bool

	// nextID generates unique handler identifiers for
	// proper unsubscribe behavior.
	nextID atomic.Uint64
}

// MQTTBrokerOptions configures the creation of a new MQTTBroker,
// allowing customization of connection URLs, QoS level, and
// authentication credentials.
type MQTTBrokerOptions struct {
	// URLs is a slice of MQTT broker URLs to connect to.
	// Format: mqtt://host:port or mqtts://host:port for TLS.
	// Multiple URLs enable automatic failover between brokers.
	URLs []string

	// QoS is the quality of service level (0, 1, or 2).
	// 0: At most once delivery (fire and forget).
	// 1: At least once delivery (recommended default).
	// 2: Exactly once delivery (highest overhead).
	// Default: 1
	QoS byte

	// Username for MQTT broker authentication (optional).
	Username string

	// Password for MQTT broker authentication (optional).
	Password string

	// KeepAlive is the interval in seconds for keep-alive pings
	// to maintain the connection. Default: 30
	KeepAlive uint16
}

// DefaultMQTTQoS is the default quality of service level used
// for MQTT publish and subscribe operations when not specified.
const DefaultMQTTQoS = 1

// DefaultMQTTKeepAlive is the default keep-alive interval in
// seconds used to maintain the MQTT connection.
const DefaultMQTTKeepAlive = 30

// convertTopic converts an event name to MQTT topic format by
// replacing dots with slashes (topic separator) and asterisks
// with plus signs (single-level wildcard). Multi-level wildcards
// (hash) are left unchanged as they match MQTT conventions.
func convertTopic(event string) string {
	topic := strings.ReplaceAll(event, ".", "/")
	topic = strings.ReplaceAll(topic, "*", "+")

	return topic
}

// matchTopic checks if a message topic matches a subscription
// pattern, supporting MQTT wildcard semantics for single-level
// plus and multi-level hash wildcards.
func matchTopic(pattern, topic string) bool {
	if pattern == topic {
		return true
	}

	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	return matchParts(patternParts, topicParts)
}

// matchParts recursively matches topic parts against pattern
// parts, handling plus for single-level and hash for multi-level
// wildcard matching according to MQTT topic filter rules.
func matchParts(pattern, topic []string) bool {
	if len(pattern) == 0 {
		return len(topic) == 0
	}

	if len(topic) == 0 {
		return pattern[0] == "#"
	}

	if pattern[0] == "#" {
		return true
	}

	if pattern[0] == "+" || pattern[0] == topic[0] {
		return matchParts(pattern[1:], topic[1:])
	}

	return false
}

// NewMQTTBroker creates a new MQTTBroker by connecting to the
// MQTT broker at the given URL with default settings. The broker
// uses clean sessions, QoS 1, and automatic reconnection. It must
// be closed when no longer needed to release resources.
//
// The URL should be in the format:
// mqtt://host:port or mqtts://host:port for TLS
func NewMQTTBroker(url string) (*MQTTBroker, error) {
	return NewMQTTBrokerWith(&MQTTBrokerOptions{
		URLs: []string{url},
		QoS:  DefaultMQTTQoS,
	})
}

// NewMQTTBrokerWith creates a new MQTTBroker using the provided
// options for connection URLs, QoS level, and authentication.
// Multiple URLs enable automatic failover between brokers. The
// broker uses clean sessions and automatic reconnection.
func NewMQTTBrokerWith(options *MQTTBrokerOptions) (*MQTTBroker, error) {
	qos := options.QoS
	if qos == 0 && len(options.URLs) > 0 {
		qos = DefaultMQTTQoS
	}

	keepAlive := options.KeepAlive
	if keepAlive == 0 {
		keepAlive = DefaultMQTTKeepAlive
	}

	urls := make([]*url.URL, len(options.URLs))
	for i, urlStr := range options.URLs {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid url %q: %w", urlStr, err)
		}
		urls[i] = u
	}

	broker := &MQTTBroker{
		qos:           qos,
		handlers:      make(map[string]map[string]contract.EventHandler),
		subscriptions: make(map[string]bool),
	}

	cfg := autopaho.ClientConfig{
		ServerUrls:                    urls,
		KeepAlive:                     keepAlive,
		CleanStartOnInitialConnection: true,
		SessionExpiryInterval:         0,
		ClientConfig: paho.ClientConfig{
			ClientID: "",
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					broker.route(pr.Packet)

					return true, nil
				},
			},
		},
	}

	if options.Username != "" {
		cfg.ConnectUsername = options.Username
		cfg.ConnectPassword = []byte(options.Password)
	}

	ctx := context.Background()
	cm, err := autopaho.NewConnection(ctx, cfg)
	if err != nil {
		return nil, err
	}

	broker.client = cm

	return broker, nil
}

// NewMQTTBrokerFrom creates a new MQTTBroker from an existing
// autopaho ConnectionManager and QoS level. This constructor is
// useful for advanced scenarios where the user needs full control
// over the MQTT connection configuration.
func NewMQTTBrokerFrom(
	client *autopaho.ConnectionManager,
	qos byte,
) *MQTTBroker {
	return &MQTTBroker{
		client:        client,
		qos:           qos,
		handlers:      make(map[string]map[string]contract.EventHandler),
		subscriptions: make(map[string]bool),
	}
}

// route delivers an incoming MQTT message to all matching
// handlers based on topic pattern matching. This implements
// fan-out behavior where multiple handlers can receive the
// same message if they subscribed to matching patterns.
func (b *MQTTBroker) route(pb *paho.Publish) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for pattern, handlers := range b.handlers {
		if matchTopic(pattern, pb.Topic) {
			for _, handler := range handlers {
				handler(func(dest any) error {
					return json.Unmarshal(pb.Payload, dest)
				})
			}
		}
	}
}

// Publish sends an event with the given name and payload to all
// subscribers listening for that event. The payload is serialized
// to JSON and the event name is converted to MQTT topic format.
//
// Publishing is thread-safe and respects context cancellation.
// The operation uses the configured QoS level for delivery
// guarantees.
func (b *MQTTBroker) Publish(
	ctx context.Context,
	event string,
	payload any,
) error {
	encoded, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	topic := convertTopic(event)

	_, err = b.client.Publish(ctx, &paho.Publish{
		Topic:   topic,
		QoS:     b.qos,
		Payload: encoded,
		Properties: &paho.PublishProperties{
			ContentType: "application/json",
		},
	})

	return err
}

// Subscribe registers a handler to receive events with the given
// name or pattern. The event name is converted to MQTT topic
// format where dots become slashes and asterisks become plus
// signs for single-level wildcard matching. Hash symbols remain
// unchanged for multi-level wildcard matching.
//
// Multiple handlers can subscribe to the same topic and all will
// receive messages (fan-out). Each handler is tracked individually
// so unsubscribing one handler does not affect others.
//
// The returned unsubscribe function removes the specific handler
// and unsubscribes from the MQTT broker only when the last handler
// for the topic is removed.
func (b *MQTTBroker) Subscribe(
	ctx context.Context,
	event string,
	handler contract.EventHandler,
) (contract.EventUnsubscribeFunc, error) {
	topic := convertTopic(event)
	handlerID := strconv.FormatUint(b.nextID.Add(1), 10)

	b.mu.Lock()
	isFirst := !b.subscriptions[topic]

	if isFirst {
		b.subscriptions[topic] = true
	}

	if b.handlers[topic] == nil {
		b.handlers[topic] = make(map[string]contract.EventHandler)
	}

	b.handlers[topic][handlerID] = handler
	b.mu.Unlock()

	if isFirst {
		_, err := b.client.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{
					Topic: topic,
					QoS:   b.qos,
				},
			},
		})

		if err != nil {
			b.mu.Lock()
			delete(b.handlers[topic], handlerID)

			if len(b.handlers[topic]) == 0 {
				delete(b.handlers, topic)
				delete(b.subscriptions, topic)
			}

			b.mu.Unlock()

			return nil, err
		}
	}

	return func() error {
		b.mu.Lock()
		delete(b.handlers[topic], handlerID)
		shouldUnsubscribe := len(b.handlers[topic]) == 0

		if shouldUnsubscribe {
			delete(b.handlers, topic)
			delete(b.subscriptions, topic)
		}

		b.mu.Unlock()

		if shouldUnsubscribe {
			_, err := b.client.Unsubscribe(ctx, &paho.Unsubscribe{
				Topics: []string{topic},
			})

			return err
		}

		return nil
	}, nil
}

// Close gracefully disconnects from the MQTT broker and releases
// all resources. This will terminate all active subscriptions and
// close the underlying connection.
func (b *MQTTBroker) Close() error {
	return b.client.Disconnect(context.Background())
}
