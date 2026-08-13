package mapper

import (
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
	"github.com/joaohgf/mjv-challenge/internal/outbound/repository/model"
)

// OrderMapper converts domain orders to and from MongoDB persistence models.
type OrderMapper struct{}

// To builds the persistence model used by the MongoDB adapter.
func (*OrderMapper) To(source *domain.Order) *model.Order {
	target := model.NewOrder()
	if source.ID != "" {
		target.ID = source.ID
	}
	target.ProductName = source.ProductName
	target.Status = string(source.Status)
	target.Quantity = source.Quantity
	target.CreatedAt = source.CreatedAt
	if source.UpdatedAt != nil {
		target.UpdatedAt = source.UpdatedAt
	}
	return target
}

// From rebuilds the domain order returned by a MongoDB query or save.
func (*OrderMapper) From(source *model.Order) *domain.Order {
	target := domain.NewOrder()
	target.ID = source.ID
	target.ProductName = source.ProductName
	target.Status = enum.Status(source.Status)
	target.Quantity = source.Quantity
	target.CreatedAt = source.CreatedAt
	if source.UpdatedAt != nil {
		target.UpdatedAt = source.UpdatedAt
	}
	return target
}
