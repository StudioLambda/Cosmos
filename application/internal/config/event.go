package config

import (
	"github.com/knadh/koanf/v2"
	"github.com/samber/do/v2"
	"github.com/studiolambda/cosmos/framework/event"
)

func NewEventMemory(i do.Injector) (event.MemoryBrokerConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := event.MemoryBrokerConfig{
		MaxConcurrentDeliveries: k.Int("event.memory.max_concurrent_deliveries"),
	}

	return config, nil
}

func NewEventAMQP(i do.Injector) (event.AMQPBrokerConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := event.AMQPBrokerConfig{
		URL:      k.String("event.amqp.url"),
		Exchange: k.String("event.amqp.exchange"),
	}

	return config, nil
}

func NewEventMQTT(i do.Injector) (event.MQTTBrokerConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := event.MQTTBrokerConfig{
		URLs:      k.Strings("event.mqtt.urls"),
		QoS:       byte(k.Int("event.mqtt.qos")),
		Username:  k.String("event.mqtt.username"),
		Password:  k.String("event.mqtt.password"),
		KeepAlive: uint16(k.Int("event.mqtt.keep_alive")),
	}

	return config, nil
}

func NewEventNATS(i do.Injector) (event.NATSBrokerConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := event.NATSBrokerConfig{
		URLs:            k.Strings("event.nats.urls"),
		Name:            k.String("event.nats.name"),
		MaxReconnects:   k.Int("event.nats.max_reconnects"),
		ReconnectWait:   k.Duration("event.nats.reconnect_wait"),
		Timeout:         k.Duration("event.nats.timeout"),
		Username:        k.String("event.nats.username"),
		Password:        k.String("event.nats.password"),
		Token:           k.String("event.nats.token"),
		NKeySeed:        k.String("event.nats.nkey_seed"),
		CredentialsFile: k.String("event.nats.credentials_file"),
		RootCAs:         k.Strings("event.nats.root_cas"),
	}

	return config, nil
}

func NewEventRedis(i do.Injector) (event.RedisBrokerConfig, error) {
	k := do.MustInvoke[*koanf.Koanf](i)

	config := event.RedisBrokerConfig{
		Network:  k.String("event.redis.network"),
		Addr:     k.String("event.redis.addr"),
		Username: k.String("event.redis.username"),
		Password: k.String("event.redis.password"),
		DB:       k.Int("event.redis.db"),
	}

	return config, nil
}
