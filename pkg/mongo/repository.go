package adapter

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/trace"
)

// Create inserts a new document and relies on its unique business ID index.
func (repository *Repository[M]) Create(ctx context.Context, model M) (modelResult M, err error) {
	ctx, span := telemetry.StartSpan(ctx, "mongodb.create", trace.SpanKindClient)
	defer func() {
		telemetry.End(span, err)
	}()
	if model.GetID() == "" {
		return model, fmt.Errorf("creating mongo document: identifier is required")
	}
	ctx, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	_, err = repository.collection().InsertOne(ctx, model)
	if err != nil {
		return model, fmt.Errorf("inserting mongo document: %w", err)
	}
	return model, nil
}

// Update replaces an existing document and reports ErrNotFound when it is absent.
func (repository *Repository[M]) Update(ctx context.Context, model M) (modelResult M, err error) {
	ctx, span := telemetry.StartSpan(ctx, "mongodb.update", trace.SpanKindClient)
	defer func() {
		telemetry.End(span, err)
	}()
	if model.GetID() == "" {
		return model, fmt.Errorf("updating mongo document: identifier is required")
	}
	ctx, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	updateResult, err := repository.collection().ReplaceOne(ctx, repository.filter(model.GetID()), model)
	if err != nil {
		return model, fmt.Errorf("update mongo document: %w", err)
	}
	if updateResult.MatchedCount == 0 {
		return model, errs.ErrNotFound
	}
	return model, nil
}

// Get returns ErrNotFound when no document matches the application identifier.
func (repository *Repository[M]) Get(ctx context.Context, id string) (modelResult M, err error) {
	ctx, span := telemetry.StartSpan(ctx, "mongodb.get", trace.SpanKindClient)
	defer func() {
		telemetry.End(span, err)
	}()
	ctx, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	var model M
	err = repository.collection().FindOne(ctx, repository.filter(id)).Decode(&model)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model, errs.ErrNotFound
	}
	if err != nil {
		return model, fmt.Errorf("get mongo document: %w", err)
	}
	return model, nil
}
