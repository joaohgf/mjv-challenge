package dto

import "time"

type (
	// Message is the RabbitMQ payload envelope consumed by the worker.
	Message[D any] struct {
		ID        string     `json:"id"`
		Type      string     `json:"type"`
		Payload   D          `json:"payload"`
		CreatedAt time.Time  `json:"createdAt"`
		UpdatedAt *time.Time `json:"updatedAt"`
	}
	// Order is the order representation carried inside a queue message.
	Order struct {
		ID          string     `json:"id"`
		ProductName string     `json:"product_name"`
		Status      string     `json:"status"`
		Quantity    int        `json:"quantity"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   *time.Time `json:"updated_at"`
	}
)
