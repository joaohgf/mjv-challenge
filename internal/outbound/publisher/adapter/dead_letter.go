package adapter

import (
	"context"

	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// DeadLetterPublisher maps a domain value before parking it in a failure queue.
type DeadLetterPublisher[D, M any] struct {
	publisher port.DeadLetterPublisher[M]
	mapper    port.To[D, M]
}

// NewDeadLetterPublisher wires a mapper and a parking-queue message publisher.
func NewDeadLetterPublisher[D, M any](publisher port.DeadLetterPublisher[M], mapper port.To[D, M]) *DeadLetterPublisher[D, M] {
	return &DeadLetterPublisher[D, M]{publisher: publisher, mapper: mapper}
}

// DeadLetter translates a domain value and parks its message representation.
func (publisher *DeadLetterPublisher[D, M]) DeadLetter(ctx context.Context, source D, cause error) error {
	return publisher.publisher.DeadLetter(ctx, publisher.mapper.To(source), cause)
}
