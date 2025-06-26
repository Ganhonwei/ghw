package mq

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type NatsConsumerOption func(c *jetstream.ConsumerConfig)

type NatsRequestOption func(c *NatsRequestOptions)

type NatsRequestOptions struct {
	Timeout time.Duration
}

// WithConsumerNoneAck jetstream 消息不ack高qps
func WithConsumerNoneAck() NatsConsumerOption {
	return func(c *jetstream.ConsumerConfig) {
		c.AckPolicy = jetstream.AckNonePolicy
	}
}

func WithConsumerMaxDeliver(maxDeliver int) NatsConsumerOption {
	return func(c *jetstream.ConsumerConfig) {
		c.MaxDeliver = maxDeliver
	}
}

func WithConsumerMaxAckPending(maxAckPending int) NatsConsumerOption {
	return func(c *jetstream.ConsumerConfig) {
		c.MaxAckPending = maxAckPending
	}
}

func WithRequestTimeout(timeout time.Duration) NatsRequestOption {
	return func(c *NatsRequestOptions) {
		c.Timeout = timeout
	}
}
