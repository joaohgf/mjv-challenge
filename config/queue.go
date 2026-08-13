package config

import "time"

// LoadQueue returns the main queue and its parking-queue configuration.
func LoadQueue() *Queue {
	target := new(Queue)
	target.RabbitMQURL = value("RABBITMQ_URL", "amqp://app:app@localhost:5672/")
	target.Name = value("RABBITMQ_QUEUE_NAME", "orders")
	target.DeadLetterName = value("RABBITMQ_DEAD_LETTER_QUEUE_NAME", "orders.dlq")
	target.PublishTimeout = duration("RABBITMQ_PUBLISH_TIMEOUT", 5*time.Second)
	target.OutboxLease = duration("OUTBOX_LEASE_DURATION", 15*time.Second)
	target.OutboxRetryInterval = duration("OUTBOX_RETRY_INTERVAL", time.Second)
	return target
}
