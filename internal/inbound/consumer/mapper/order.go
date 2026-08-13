package mapper

import (
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
	"github.com/joaohgf/mjv-challenge/internal/inbound/consumer/dto"
)

// Order maps a consumed queue payload to the domain representation.
type Order struct{}

// To preserves the order identity and lifecycle from the queue message.
func (*Order) To(source *dto.Order) *domain.Order {
	target := domain.NewOrder()
	target.ID = source.ID
	target.ProductName = source.ProductName
	target.Status = enum.Status(source.Status)
	target.Quantity = source.Quantity
	target.CreatedAt = source.CreatedAt
	target.UpdatedAt = source.UpdatedAt
	return target
}
