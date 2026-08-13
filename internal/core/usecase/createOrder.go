package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/enum"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// CreateOrder initializes a new order before its transactional persistence.
type CreateOrder struct {
	creator port.Creator[*domain.Order]
}

// NewCreateOrder wires the transactional creation port.
func NewCreateOrder(creator port.Creator[*domain.Order]) *CreateOrder {
	return &CreateOrder{creator: creator}
}

// Create delegates persistence and outbox enqueueing to one transactional boundary.
func (co *CreateOrder) Create(ctx context.Context, source *domain.Order) (*domain.Order, error) {
	source.ID = uuid.NewString()
	source.CreatedAt = time.Now().UTC()
	source.Status = enum.Criado
	created, err := co.creator.Create(ctx, source)
	if err != nil {
		return nil, err
	}
	return created, nil
}
