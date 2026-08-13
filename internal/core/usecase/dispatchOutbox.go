package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// DispatchOutbox publishes one claimed event and records its confirmed delivery.
type DispatchOutbox[D any] struct {
	outbox      port.Outbox[D]
	publisher   port.Publisher[D]
	deadLetter  port.DeadLetterPublisher[D]
	maxAttempts int
}

// NewDispatchOutbox wires durable events to their asynchronous transport.
func NewDispatchOutbox[D any](outbox port.Outbox[D], publisher port.Publisher[D], deadLetter port.DeadLetterPublisher[D], maxAttempts int) *DispatchOutbox[D] {
	return &DispatchOutbox[D]{outbox: outbox, publisher: publisher, deadLetter: deadLetter, maxAttempts: maxAttempts}
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
	messageContext := eventContext(ctx, event.Context)
	if event.Attempts > dispatcher.maxAttempts {
		return true, dispatcher.deadLetterEvent(ctx, messageContext, event, attemptLimitError(dispatcher.maxAttempts))
	}
	if err := dispatcher.publisher.Publish(messageContext, event.Payload); err != nil {
		if event.Attempts == dispatcher.maxAttempts {
			return true, dispatcher.deadLetterEvent(ctx, messageContext, event, err)
		}
		return true, dispatcher.release(ctx, event, err)
	}
	if err := dispatcher.outbox.MarkPublished(ctx, event.ID, event.LeaseToken); err != nil {
		return true, fmt.Errorf("marking outbox event published: %w", err)
	}
	return true, nil
}

func (dispatcher *DispatchOutbox[D]) deadLetterEvent(ctx context.Context, eventContext context.Context, event *domain.OutboxEvent[D], cause error) error {
	if err := dispatcher.deadLetter.DeadLetter(eventContext, event.Payload, cause); err != nil {
		return dispatcher.release(ctx, event, err)
	}
	if err := dispatcher.outbox.MarkDeadLettered(ctx, event.ID, event.LeaseToken); err != nil {
		return fmt.Errorf("marking outbox event dead-lettered: %w", err)
	}
	return nil
}

func attemptLimitError(maxAttempts int) error {
	return fmt.Errorf("outbox event exceeded %d publish attempts", maxAttempts)
}

// eventContext preserves restored propagation data and falls back for legacy events.
func eventContext(fallback, restored context.Context) context.Context {
	if restored != nil {
		return restored
	}
	return fallback
}

func (dispatcher *DispatchOutbox[D]) release(ctx context.Context, event *domain.OutboxEvent[D], publishErr error) error {
	if err := dispatcher.outbox.Release(ctx, event.ID, event.LeaseToken); err != nil {
		return errors.Join(fmt.Errorf("publishing outbox event: %w", publishErr), fmt.Errorf("releasing outbox event: %w", err))
	}
	return fmt.Errorf("publishing outbox event: %w", publishErr)
}
