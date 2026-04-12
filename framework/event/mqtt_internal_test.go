package event

import (
	"encoding/json"
	"sync/atomic"
	"testing"

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

func TestMQTTBrokerRouteDeliversToMatchingHandler(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					called.Store(true)
				},
			},
		},
	}

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"hello"`),
	})

	require.True(t, called.Load())
}

func TestMQTTBrokerRouteDeliversToWildcardHandler(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/+": {
				"1": func(payload contract.EventPayload) {
					called.Store(true)
				},
			},
		},
	}

	broker.route(&paho.Publish{
		Topic:   "user/123",
		Payload: []byte(`"data"`),
	})

	require.True(t, called.Load())
}

func TestMQTTBrokerRouteDoesNotDeliverToNonMatching(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					called.Store(true)
				},
			},
		},
	}

	broker.route(&paho.Publish{
		Topic:   "order/created",
		Payload: []byte(`"data"`),
	})

	require.False(t, called.Load())
}

func TestMQTTBrokerRouteDeliversToMultipleHandlersSamePattern(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					count.Add(1)
				},
				"2": func(payload contract.EventPayload) {
					count.Add(1)
				},
			},
		},
	}

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"hello"`),
	})

	require.Equal(t, int32(2), count.Load())
}

func TestMQTTBrokerRouteFanOutAcrossPatterns(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
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
		},
	}

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"hello"`),
	})

	require.Equal(t, int32(2), count.Load())
}

func TestMQTTBrokerRouteUnmarshalsJSONPayload(t *testing.T) {
	t.Parallel()

	var received string

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					_ = payload(&received)
				},
			},
		},
	}

	data, err := json.Marshal("hello world")

	require.NoError(t, err)

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: data,
	})

	require.Equal(t, "hello world", received)
}

func TestMQTTBrokerRoutePanicDoesNotPropagate(t *testing.T) {
	t.Parallel()

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					panic("handler panic")
				},
			},
		},
	}

	require.NotPanics(t, func() {
		broker.route(&paho.Publish{
			Topic:   "user/created",
			Payload: []byte(`"data"`),
		})
	})
}

func TestMQTTBrokerRoutePanicDoesNotAffectOtherHandlers(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					panic("handler panic")
				},
				"2": func(payload contract.EventPayload) {
					called.Store(true)
				},
			},
		},
	}

	broker.route(&paho.Publish{
		Topic:   "user/created",
		Payload: []byte(`"data"`),
	})

	require.True(t, called.Load())
}

func TestMQTTBrokerHandlePublishRoutesMessage(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					called.Store(true)
				},
			},
		},
	}

	handled, err := broker.HandlePublish(paho.PublishReceived{
		Packet: &paho.Publish{
			Topic:   "user/created",
			Payload: []byte(`"hello"`),
		},
	})

	require.NoError(t, err)
	require.True(t, handled)
	require.True(t, called.Load())
}

func TestMQTTBrokerHandlePublishRecoversPanic(t *testing.T) {
	t.Parallel()

	broker := &MQTTBroker{
		handlers: map[string]map[string]contract.EventHandler{
			"user/created": {
				"1": func(payload contract.EventPayload) {
					panic("handler panic")
				},
			},
		},
	}

	require.NotPanics(t, func() {
		handled, err := broker.HandlePublish(paho.PublishReceived{
			Packet: &paho.Publish{
				Topic:   "user/created",
				Payload: []byte(`"data"`),
			},
		})

		require.NoError(t, err)
		require.True(t, handled)
	})
}
