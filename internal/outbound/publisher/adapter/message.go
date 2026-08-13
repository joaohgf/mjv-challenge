package adapter

import (
	"context"

	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// Publisher maps domain values before delegating to a generic message publisher.
type Publisher[D, M any] struct {
	publisher port.Publisher[M]
	mapper    port.To[D, M]
}

// NewPublisher wires the domain-to-message mapper and transport publisher.
func NewPublisher[D, M any](publisher port.Publisher[M], mapper port.To[D, M]) *Publisher[D, M] {
	return &Publisher[D, M]{publisher: publisher, mapper: mapper}
}

// Publish translates a domain value and sends its message representation.
func (publisher *Publisher[D, M]) Publish(ctx context.Context, source D) error {
	target := publisher.mapper.To(source)
	err := publisher.publisher.Publish(ctx, target)
	return err
}
