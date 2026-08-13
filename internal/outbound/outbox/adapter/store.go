package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
	"github.com/joaohgf/mjv-challenge/internal/enum"
	"github.com/joaohgf/mjv-challenge/internal/outbound/outbox/model"
	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"go.mongodb.org/mongo-driver/mongo"
)

// Store persists and leases mapped events in the MongoDB outbox collection.
type Store[D, M any] struct {
	collection *mongo.Collection
	mapper     port.Mapper[D, M]
	lease      time.Duration
}

// NewStore creates a MongoDB outbox store with its event lease duration.
func NewStore[D, M any](collection *mongo.Collection, mapper port.Mapper[D, M], lease time.Duration) *Store[D, M] {
	return &Store[D, M]{collection: collection, mapper: mapper, lease: lease}
}

// Enqueue inserts a pending event in the same transaction as its aggregate.
func (store *Store[D, M]) Enqueue(ctx context.Context, payload D) error {
	now := time.Now().UTC()
	event := &model.Event[M]{
		ID: uuid.NewString(), Payload: store.mapper.To(payload), Status: enum.OutboxPending,
		TraceContext: telemetry.InjectContext(ctx, nil), CreatedAt: now, UpdatedAt: now,
	}
	if _, err := store.collection.InsertOne(ctx, event); err != nil {
		return fmt.Errorf("inserting outbox event: %w", err)
	}
	return nil
}
