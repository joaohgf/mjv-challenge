package adapter

import (
	"context"
	"fmt"

	"github.com/joaohgf/mjv-challenge/internal/core/port"
)

// Creator persists an aggregate and its outbox event in one transaction.
type Creator[D any] struct {
	transaction port.TransactionRunner
	repository  port.PersistentCreator[D]
	outbox      port.Outbox[D]
}

// NewCreator joins aggregate persistence and durable event enqueueing.
func NewCreator[D any](transaction port.TransactionRunner, repository port.PersistentCreator[D], outbox port.Outbox[D]) *Creator[D] {
	return &Creator[D]{transaction: transaction, repository: repository, outbox: outbox}
}

// Create inserts an aggregate and enqueues it atomically for later publication.
func (creator *Creator[D]) Create(ctx context.Context, source D) (D, error) {
	var created D
	err := creator.transaction.Transaction(ctx, func(transactionContext context.Context) error {
		result, err := creator.repository.Create(transactionContext, source)
		if err != nil {
			return fmt.Errorf("creating aggregate: %w", err)
		}
		if err := creator.outbox.Enqueue(transactionContext, result); err != nil {
			return fmt.Errorf("enqueueing outbox event: %w", err)
		}
		created = result
		return nil
	})
	if err != nil {
		return source, fmt.Errorf("creating aggregate transaction: %w", err)
	}
	return created, nil
}
