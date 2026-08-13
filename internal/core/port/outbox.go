package port

import (
	"context"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
)

type (
	// TransactionRunner executes an operation atomically in a persistence transaction.
	TransactionRunner interface {
		Transaction(context.Context, func(context.Context) error) error
	}
	// Outbox stores messages durably until a relay publishes them.
	Outbox[D any] interface {
		Enqueue(context.Context, D) error
		Claim(context.Context) (*domain.OutboxEvent[D], error)
		MarkPublished(context.Context, string, string) error
		MarkDeadLettered(context.Context, string, string) error
		Release(context.Context, string, string) error
	}
	// Dispatcher publishes at most one available outbox event.
	Dispatcher interface {
		Dispatch(context.Context) (bool, error)
	}
)
