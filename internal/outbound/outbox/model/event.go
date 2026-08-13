package model

import (
	"time"

	"github.com/joaohgf/mjv-challenge/internal/enum"
)

// Event is the MongoDB document that durably records an outbound message.
type Event[D any] struct {
	ID           string            `bson:"event_id"`
	Payload      D                 `bson:"payload"`
	Status       enum.OutboxStatus `bson:"status"`
	Attempts     int               `bson:"attempts"`
	TraceContext map[string]string `bson:"trace_context,omitempty"`
	CreatedAt    time.Time         `bson:"created_at"`
	UpdatedAt    time.Time         `bson:"updated_at"`
	LockedUntil  *time.Time        `bson:"locked_until,omitempty"`
	PublishedAt  *time.Time        `bson:"published_at,omitempty"`
}
