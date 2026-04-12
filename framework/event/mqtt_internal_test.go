package event

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/studiolambda/cosmos/contract"

	"github.com/eclipse/paho.golang/paho"
	"github.com/stretchr/testify/require"
)

func TestConvertTopicDotsToSlashes(t *testing.T) {
	t.Parallel()

	result := convertTopic("user.created")

	require.Equal(t, "user/created", result)
}

func TestConvertTopicStarToPlus(t *testing.T) {
	t.Parallel()

	result := convertTopic("user.*.created")

	require.Equal(t, "user/+/created", result)
}

func TestConvertTopicHashUnchanged(t *testing.T) {
	t.Parallel()

	result := convertTopic("logs.#")

	require.Equal(t, "logs/#", result)
}

func TestConvertTopicCombined(t *testing.T) {
	t.Parallel()

	result := convertTopic("user.*.#")

	require.Equal(t, "user/+/#", result)
}

func TestConvertTopicNoSeparators(t *testing.T) {
	t.Parallel()

	result := convertTopic("simple")

	require.Equal(t, "simple", result)
}

func TestConvertTopicEmptyString(t *testing.T) {
	t.Parallel()

	result := convertTopic("")

	require.Equal(t, "", result)
}

func TestMatchTopicExactMatch(t *testing.T) {
	t.Parallel()

	result := matchTopic("user/created", "user/created")

	require.True(t, result)
}

func TestMatchTopicNoMatch(t *testing.T) {
	t.Parallel()

	result := matchTopic("user/created", "order/created")

	require.False(t, result)
}

func TestMatchTopicPlusMatchesSingleLevel(t *testing.T) {
	t.Parallel()

	result := matchTopic("user/+/created", "user/123/created")

	require.True(t, result)
}

func TestMatchTopicPlusDoesNotMatchMultipleLevels(t *testing.T) {
	t.Parallel()

	result := matchTopic("user/+/created", "user/1/2/created")

	require.False(t, result)
}

func TestMatchTopicHashMatchesZeroTrailing(t *testing.T) {
	t.Parallel()

	result := matchTopic("logs/#", "logs")

	require.True(t, result)
}

func TestMatchTopicHashMatchesOneTrailing(t *testing.T) {
	t.Parallel()

	result := matchTopic("logs/#", "logs/error")

	require.True(t, result)
}

func TestMatchTopicHashMatchesMultipleTrailing(t *testing.T) {
	t.Parallel()

	result := matchTopic("logs/#", "logs/error/db")

	require.True(t, result)
}

func TestMatchTopicHashAloneMatchesEverything(t *testing.T) {
	t.Parallel()

	result := matchTopic("#", "anything/here")

	require.True(t, result)
}

func TestMatchPartsBothEmptySlices(t *testing.T) {
	t.Parallel()

	result := matchParts([]string{}, []string{})

	require.True(t, result)
}

func TestMatchPartsEmptyPatternNonEmptyTopic(t *testing.T) {
	t.Parallel()

	result := matchParts([]string{}, []string{"a"})

	require.False(t, result)
}

func TestMatchPartsHashPatternEmptyTopic(t *testing.T) {
	t.Parallel()

	result := matchParts([]string{"#"}, []string{})

	require.True(t, result)
}

// newTestMQTTBroker creates an MQTTBroker with initialized
// semaphore for use in route() tests that don't need a real
// MQTT connection.
func newTestMQTTBroker(handlers map[string]map[string]contract.EventHandler) *MQTTBroker {
	return &MQTTBroker{
		handlers: handlers,
		sem:      make(chan struct{}, DefaultMaxConcurrentDeliveries),
	}
}

func TestMQTTBrokerRouteDeliversToMatchingHandler(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"user/created": {
			"1": func(payload contract.EventPayload) {
				called.Store(true)
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"hello"`),
	})

	broker.routeWg.Wait()

	require.True(t, called.Load())
}

func TestMQTTBrokerRouteDeliversToWildcardHandler(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"user/+": {
			"1": func(payload contract.EventPayload) {
				called.Store(true)
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "user/123",
		Payload: []byte(`"data"`),
	})

	broker.routeWg.Wait()

	require.True(t, called.Load())
}

func TestMQTTBrokerRouteDoesNotDeliverToNonMatching(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"user/created": {
			"1": func(payload contract.EventPayload) {
				called.Store(true)
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "order/created",
		Payload: []byte(`"data"`),
	})

	broker.routeWg.Wait()

	require.False(t, called.Load())
}

func TestMQTTBrokerRouteDeliversToMultipleHandlersSamePattern(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"user/created": {
			"1": func(payload contract.EventPayload) {
				count.Add(1)
			},
			"2": func(payload contract.EventPayload) {
				count.Add(1)
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"hello"`),
	})

	broker.routeWg.Wait()

	require.Equal(t, int32(2), count.Load())
}

func TestMQTTBrokerRouteFanOutAcrossPatterns(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"user/created": {
			"1": func(payload contract.EventPayload) {
				count.Add(1)
			},
		},
		"user/+": {
			"2": func(payload contract.EventPayload) {
				count.Add(1)
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"hello"`),
	})

	broker.routeWg.Wait()

	require.Equal(t, int32(2), count.Load())
}

func TestMQTTBrokerRouteUnmarshalsJSONPayload(t *testing.T) {
	t.Parallel()

	var received atomic.Value

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"user/created": {
			"1": func(payload contract.EventPayload) {
				var value string

				_ = payload(&value)

				received.Store(value)
			},
		},
	})

	data, err := json.Marshal("hello world")

	require.NoError(t, err)

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: data,
	})

	broker.routeWg.Wait()

	require.Equal(t, "hello world", received.Load())
}

func TestMQTTBrokerRouteDispatchesAsynchronously(t *testing.T) {
	t.Parallel()

	var started sync.WaitGroup
	var finished sync.WaitGroup

	started.Add(2)
	finished.Add(2)

	gate := make(chan struct{})

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"events/test": {
			"1": func(payload contract.EventPayload) {
				started.Done()
				<-gate
				finished.Done()
			},
			"2": func(payload contract.EventPayload) {
				started.Done()
				<-gate
				finished.Done()
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "events/test",
		Payload: []byte(`"data"`),
	})

	// Both handlers must start concurrently. If route() were
	// synchronous, the second handler could never start while
	// the first is blocked on the gate.
	done := make(chan struct{})

	go func() {
		started.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Both handlers started concurrently — success.
	case <-time.After(2 * time.Second):
		t.Fatal("handlers did not start concurrently within timeout")
	}

	close(gate)

	finished.Wait()
	broker.routeWg.Wait()
}

func TestMQTTBrokerRouteRecoversPanic(t *testing.T) {
	t.Parallel()

	var secondCalled atomic.Bool

	broker := newTestMQTTBroker(map[string]map[string]contract.EventHandler{
		"events/panic": {
			"1": func(payload contract.EventPayload) {
				panic("test panic")
			},
		},
		"events/+": {
			"2": func(payload contract.EventPayload) {
				secondCalled.Store(true)
			},
		},
	})

	broker.route(&paho.Publish{
		Topic:   "events/panic",
		Payload: []byte(`"data"`),
	})

	broker.routeWg.Wait()

	require.True(t, secondCalled.Load())
}
