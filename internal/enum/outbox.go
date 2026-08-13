package enum

// OutboxStatus represents the delivery lifecycle of an outbox event.
type OutboxStatus string

const (
	// OutboxPending means the relay may claim the event for publication.
	OutboxPending OutboxStatus = "PENDING"
	// OutboxProcessing means a relay holds the event until its lease expires.
	OutboxProcessing OutboxStatus = "PROCESSING"
	// OutboxPublished means RabbitMQ confirmed receipt of the event.
	OutboxPublished OutboxStatus = "PUBLISHED"
	// OutboxDeadLettered means the event was parked after reaching the attempt limit.
	OutboxDeadLettered OutboxStatus = "DEAD_LETTERED"
)
