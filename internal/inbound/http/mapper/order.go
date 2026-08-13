package mapper

import (
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/inbound/http/dto"
)

// Order translates between HTTP order DTOs and the domain order.
type Order struct{}

// To converts a creation request into a domain order without lifecycle fields.
func (*Order) To(source *dto.OrderCreate) *domain.Order {
	target := domain.NewOrder()
	target.ProductName = source.ProductName
	target.Quantity = source.Quantity
	return target
}

// From converts a domain order into the response sent to HTTP clients.
func (*Order) From(source *domain.Order) *dto.OrderResponse {
	target := dto.NewOrderResponse()
	target.ID = source.ID
	target.ProductName = source.ProductName
	target.Status = string(source.Status)
	target.Quantity = source.Quantity
	target.CreatedAt = source.CreatedAt
	target.UpdatedAt = source.UpdatedAt
	return target
}
