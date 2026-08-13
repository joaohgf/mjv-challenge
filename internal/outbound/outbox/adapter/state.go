package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/enum"
	"go.mongodb.org/mongo-driver/bson"
)

// MarkPublished records that RabbitMQ confirmed receipt of the claimed event.
func (store *Store[D, M]) MarkPublished(ctx context.Context, id string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"status": enum.OutboxPublished, "updated_at": now, "published_at": now}, "$unset": bson.M{"locked_until": ""}}
	if _, err := store.collection.UpdateOne(ctx, claimed(id), update); err != nil {
		return fmt.Errorf("marking outbox event published: %w", err)
	}
	return nil
}

// MarkDeadLettered records that the event was parked after publish failures.
func (store *Store[D, M]) MarkDeadLettered(ctx context.Context, id string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"status": enum.OutboxDeadLettered, "updated_at": now}, "$unset": bson.M{"locked_until": ""}}
	if _, err := store.collection.UpdateOne(ctx, claimed(id), update); err != nil {
		return fmt.Errorf("marking outbox event dead-lettered: %w", err)
	}
	slog.Warn("outbox event sent to dead-letter queue", "event_id", id)
	return nil
}

// Release makes a failed event immediately available to the next relay attempt.
func (store *Store[D, M]) Release(ctx context.Context, id string) error {
	update := bson.M{"$set": bson.M{"status": enum.OutboxPending, "updated_at": time.Now().UTC()}, "$unset": bson.M{"locked_until": ""}}
	if _, err := store.collection.UpdateOne(ctx, claimed(id), update); err != nil {
		return fmt.Errorf("releasing outbox event: %w", err)
	}
	return nil
}

func claimed(id string) bson.M {
	return bson.M{"event_id": id, "status": enum.OutboxProcessing}
}
