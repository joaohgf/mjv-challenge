package model

import "time"

const (
	// Pending means the relay may claim the event for publication.
	Pending = "PENDING"
	// Processing means a relay holds the event until its lease expires.
	Processing = "PROCESSING"
	// Published means RabbitMQ confirmed receipt of the event.
	Published = "PUBLISHED"
)

// Event is the MongoDB document that durably records an outbound message.
type Event[D any] struct {
	ID          string     `bson:"event_id"`
	Payload     D          `bson:"payload"`
	Status      string     `bson:"status"`
	Attempts    int        `bson:"attempts"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"`
	LockedUntil *time.Time `bson:"locked_until,omitempty"`
	PublishedAt *time.Time `bson:"published_at,omitempty"`
}
