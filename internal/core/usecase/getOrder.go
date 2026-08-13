package usecase

import (
	"context"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// GetOrder retrieves one order without exposing the persistence implementation.
type GetOrder struct {
	repository port.Getter[*domain.Order]
}

// NewGetOrder creates the read use case for the supplied repository.
func NewGetOrder(repository port.Getter[*domain.Order]) *GetOrder {
	return &GetOrder{repository: repository}
}

// Get delegates lookup to the repository so inbound adapters stay storage-agnostic.
func (getter *GetOrder) Get(ctx context.Context, id string) (*domain.Order, error) {
	return getter.repository.Get(ctx, id)
}
