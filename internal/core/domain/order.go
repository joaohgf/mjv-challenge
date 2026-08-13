package domain

import (
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/enum"
)

// Order is the domain representation of a submitted order and its lifecycle.
type Order struct {
	ID          string
	ProductName string
	Status      enum.Status
	Quantity    int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// NewOrder creates an empty domain order for a mapper or use case to populate.
func NewOrder() *Order {
	target := new(Order)
	return target
}
