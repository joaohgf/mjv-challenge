package dto

import "time"

// Order is the order payload serialized inside an outbound message.
type Order struct {
	ID          string     `json:"id"`
	ProductName string     `json:"product_name"`
	Status      string     `json:"status"`
	Quantity    int        `json:"quantity"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
