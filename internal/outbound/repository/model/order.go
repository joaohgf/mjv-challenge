package model

import (
	"time"
)

// Order is the MongoDB document representation of a domain order.
type Order struct {
	ID          string     `bson:"order_id"`
	ProductName string     `bson:"product"`
	Status      string     `bson:"status"`
	Quantity    int        `bson:"quantity"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   *time.Time `bson:"updatedAt"`
}

// NewOrder creates an empty persistence model for a mapper to populate.
func NewOrder() *Order {
	target := new(Order)
	return target
}

// GetID exposes the business identifier used by the generic Mongo repository.
func (order *Order) GetID() string {
	return order.ID
}

// GetIDField identifies the MongoDB field used to find an order document.
func (*Order) GetIDField() string {
	return "order_id"
}
