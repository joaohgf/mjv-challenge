package mapper

import (
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/outbound/publisher/dto"
)

// Order maps a domain order into the outbound RabbitMQ message envelope.
type Order struct{}

// To creates a typed order message while preserving lifecycle metadata.
func (*Order) To(source *domain.Order) *dto.Message[*dto.Order] {
	target := dto.NewMessage[*dto.Order]()
	target.ID = source.ID
	target.Payload = &dto.Order{
		ID: source.ID, ProductName: source.ProductName, Status: string(source.Status),
		Quantity: source.Quantity, CreatedAt: source.CreatedAt, UpdatedAt: source.UpdatedAt,
	}
	target.Type = "order"
	target.CreatedAt = source.CreatedAt
	if source.UpdatedAt != nil {
		target.UpdatedAt = source.UpdatedAt
	}
	return target
}
