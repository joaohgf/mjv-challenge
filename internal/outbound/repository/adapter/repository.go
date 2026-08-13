package adapter

import (
	"context"
	"fmt"

	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// Repository maps domain objects at the persistence boundary.
type Repository[D any, M port.Identifiable] struct {
	store  port.Store[M]
	mapper port.Mapper[D, M]
}

// NewRepository combines generic storage with deterministic domain mapping.
func NewRepository[D any, M port.Identifiable](store port.Store[M], mapper port.Mapper[D, M]) *Repository[D, M] {
	return &Repository[D, M]{store: store, mapper: mapper}
}

// Save maps the domain value before storing it and maps the result back.
func (repository *Repository[D, M]) Save(ctx context.Context, source D) (D, error) {
	model, err := repository.store.Save(ctx, repository.mapper.To(source))
	if err != nil {
		return source, fmt.Errorf("saving repository model: %w", err)
	}
	return repository.mapper.From(model), nil
}

// Update replaces an existing persistence model and maps it back to the domain.
func (repository *Repository[D, M]) Update(ctx context.Context, source D) (D, error) {
	model, err := repository.store.Update(ctx, repository.mapper.To(source))
	if err != nil {
		return source, fmt.Errorf("updating repository model: %w", err)
	}
	return repository.mapper.From(model), nil
}

// Get reads a persistence model by ID and maps it back to the domain type.
func (repository *Repository[D, M]) Get(ctx context.Context, id string) (D, error) {
	model, err := repository.store.Get(ctx, id)
	if err != nil {
		var empty D
		return empty, fmt.Errorf("getting repository model: %w", err)
	}
	return repository.mapper.From(model), nil
}
