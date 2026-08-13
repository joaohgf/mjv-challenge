package adapter

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Save upserts a document using its business ID instead of creating duplicates.
func (repository *Repository[M]) Save(ctx context.Context, model M) (M, error) {
	if model.GetID() == "" {
		return model, fmt.Errorf("saving mongo document: identifier is required")
	}
	ctx, cancel := context.WithTimeout(ctx, repository.config.SaveTimeout)
	defer cancel()
	_, err := repository.collection().ReplaceOne(ctx, repository.filter(model.GetID()), model, options.Replace().SetUpsert(true))
	if err != nil {
		return model, fmt.Errorf("upsert mongo document: %w", err)
	}
	return model, nil
}

// Update replaces an existing document and reports ErrNotFound when it is absent.
func (repository *Repository[M]) Update(ctx context.Context, model M) (M, error) {
	if model.GetID() == "" {
		return model, fmt.Errorf("updating mongo document: identifier is required")
	}
	result, err := repository.collection().ReplaceOne(ctx, repository.filter(model.GetID()), model)
	if err != nil {
		return model, fmt.Errorf("update mongo document: %w", err)
	}
	if result.MatchedCount == 0 {
		return model, errs.ErrNotFound
	}
	return model, nil
}

// Get returns ErrNotFound when no document matches the application identifier.
func (repository *Repository[M]) Get(ctx context.Context, id string) (M, error) {
	var model M
	err := repository.collection().FindOne(ctx, repository.filter(id)).Decode(&model)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model, errs.ErrNotFound
	}
	if err != nil {
		return model, fmt.Errorf("get mongo document: %w", err)
	}
	return model, nil
}
