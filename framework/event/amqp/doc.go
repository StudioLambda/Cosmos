// Package amqp provides AMQP event publisher and subscriber drivers.
//
// [NewAMQPBroker] owns the connection it creates. [NewAMQPBrokerFrom] accepts
// a caller-owned connection and never closes it.
package amqp
