package middlewares

import "time"

// Delivery 消息分配器接口
type Delivery interface {
	Payload() []byte
	Ack() bool
}

// Consumer 消费者接口
type Consumer interface {
	Consume(delivery Delivery)
}

// Queue 异步队列接口
type Queue interface {
	Publish(payload []byte) bool
	AddConsumer(tag string, consumer Consumer) string
	StartConsuming(prefetchLimit int, pollDuration time.Duration) bool
	StopConsuming() bool
	Close() bool
}
