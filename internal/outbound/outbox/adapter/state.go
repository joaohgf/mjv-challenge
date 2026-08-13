package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/outbound/outbox/model"
	"go.mongodb.org/mongo-driver/bson"
)

// MarkPublished records that RabbitMQ confirmed receipt of the claimed event.
func (store *Store[D, M]) MarkPublished(ctx context.Context, id string) error {
	now := time.Now().UTC()
	update := bson.M{"$set": bson.M{"status": model.Published, "updated_at": now, "published_at": now}, "$unset": bson.M{"locked_until": ""}}
	if _, err := store.collection.UpdateOne(ctx, claimed(id), update); err != nil {
		return fmt.Errorf("marking outbox event published: %w", err)
	}
	return nil
}

// Release makes a failed event immediately available to the next relay attempt.
func (store *Store[D, M]) Release(ctx context.Context, id string) error {
	update := bson.M{"$set": bson.M{"status": model.Pending, "updated_at": time.Now().UTC()}, "$unset": bson.M{"locked_until": ""}}
	if _, err := store.collection.UpdateOne(ctx, claimed(id), update); err != nil {
		return fmt.Errorf("releasing outbox event: %w", err)
	}
	return nil
}

func claimed(id string) bson.M {
	return bson.M{"event_id": id, "status": model.Processing}
}
