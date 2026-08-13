package adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/outbound/outbox/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Claim leases the oldest pending or expired event for one relay invocation.
func (store *Store[D, M]) Claim(ctx context.Context) (*domain.OutboxEvent[D], error) {
	now := time.Now().UTC()
	var event model.Event[M]
	result := store.collection.FindOneAndUpdate(ctx, available(now), claimUpdate(now, store.lease), claimOptions())
	if err := result.Decode(&event); errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errs.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("claiming outbox event: %w", err)
	}
	return &domain.OutboxEvent[D]{ID: event.ID, Payload: store.mapper.From(event.Payload), Attempts: event.Attempts}, nil
}

func available(now time.Time) bson.M {
	return bson.M{"$or": bson.A{
		bson.M{"status": model.Pending},
		bson.M{"status": model.Processing, "locked_until": bson.M{"$lte": now}},
	}}
}

func claimUpdate(now time.Time, lease time.Duration) bson.M {
	return bson.M{"$set": bson.M{"status": model.Processing, "locked_until": now.Add(lease), "updated_at": now}, "$inc": bson.M{"attempts": 1}}
}

func claimOptions() *options.FindOneAndUpdateOptions {
	return options.FindOneAndUpdate().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetReturnDocument(options.After)
}
