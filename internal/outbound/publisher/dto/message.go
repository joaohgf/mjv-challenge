package dto

import (
	"time"
)

// Message is the generic JSON envelope published to RabbitMQ.
type Message[D any] struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Payload   D          `json:"payload"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// NewMessage creates an empty message envelope for a publisher mapper.
func NewMessage[D any]() *Message[D] {
	target := new(Message[D])
	return target
}
