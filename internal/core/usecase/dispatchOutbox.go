package usecase

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// DispatchOutbox publishes one claimed event and records its confirmed delivery.
type DispatchOutbox[D any] struct {
	outbox    port.Outbox[D]
	publisher port.Publisher[D]
}

// NewDispatchOutbox wires durable events to their asynchronous transport.
func NewDispatchOutbox[D any](outbox port.Outbox[D], publisher port.Publisher[D]) *DispatchOutbox[D] {
	return &DispatchOutbox[D]{outbox: outbox, publisher: publisher}
}

// Dispatch returns false when no event is available and true after claiming one.
func (dispatcher *DispatchOutbox[D]) Dispatch(ctx context.Context) (bool, error) {
	event, err := dispatcher.outbox.Claim(ctx)
	if errors.Is(err, errs.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claiming outbox event: %w", err)
	}
	if err := dispatcher.publisher.Publish(ctx, event.Payload); err != nil {
		return true, dispatcher.release(ctx, event.ID, err)
	}
	if err := dispatcher.outbox.MarkPublished(ctx, event.ID); err != nil {
		return true, fmt.Errorf("marking outbox event published: %w", err)
	}
	return true, nil
}

func (dispatcher *DispatchOutbox[D]) release(ctx context.Context, id string, publishErr error) error {
	if err := dispatcher.outbox.Release(ctx, id); err != nil {
		return errors.Join(fmt.Errorf("publishing outbox event: %w", publishErr), fmt.Errorf("releasing outbox event: %w", err))
	}
	return fmt.Errorf("publishing outbox event: %w", publishErr)
}
