package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/enum"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MarkPublished records that RabbitMQ confirmed receipt of the claimed event.
func (store *Store[D, M]) MarkPublished(ctx context.Context, id, leaseToken string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"status": enum.OutboxPublished, "updated_at": now, "published_at": now}, "$unset": unlockFields()}
	if err := store.transition(ctx, id, leaseToken, update); err != nil {
		return fmt.Errorf("marking outbox event published: %w", err)
	}
	return nil
}

// MarkDeadLettered records that the event was parked after publish failures.
func (store *Store[D, M]) MarkDeadLettered(ctx context.Context, id, leaseToken string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"status": enum.OutboxDeadLettered, "updated_at": now}, "$unset": unlockFields()}
	if err := store.transition(ctx, id, leaseToken, update); err != nil {
		return fmt.Errorf("marking outbox event dead-lettered: %w", err)
	}
	slog.Warn("outbox event sent to dead-letter queue", "event_id", id)
	return nil
}

// Release makes a failed event immediately available to the next relay attempt.
func (store *Store[D, M]) Release(ctx context.Context, id, leaseToken string) error {
	update := bson.M{"$set": bson.M{"status": enum.OutboxPending, "updated_at": time.Now().UTC()}, "$unset": unlockFields()}
	if err := store.transition(ctx, id, leaseToken, update); err != nil {
		return fmt.Errorf("releasing outbox event: %w", err)
	}
	return nil
}

func (store *Store[D, M]) transition(ctx context.Context, id, leaseToken string, update bson.M) error {
	operationContext, cancel := store.withTimeout(ctx)
	defer cancel()
	result, err := store.collection.UpdateOne(operationContext, claimed(id, leaseToken), update)
	if err != nil {
		return err
	}
	return requireLease(result)
}

func claimed(id, leaseToken string) bson.M {
	return bson.M{"event_id": id, "status": enum.OutboxProcessing, "lease_token": leaseToken}
}

func unlockFields() bson.M {
	return bson.M{"locked_until": "", "lease_token": ""}
}

func requireLease(result *mongo.UpdateResult) error {
	if result == nil || result.MatchedCount == 0 {
		return errs.ErrLeaseLost
	}
	return nil
}
