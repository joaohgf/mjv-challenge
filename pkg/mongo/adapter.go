package adapter

import (
	"context"
	"fmt"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository provides generic MongoDB persistence; mapping belongs to outbound adapters.
type Repository[M port.Identifiable] struct {
	config *config.Database
	client *mongo.Client
}

// NewRepository creates a generic MongoDB repository for an identifiable model.
func NewRepository[M port.Identifiable](config *config.Database) *Repository[M] {
	return &Repository[M]{config: config}
}

// Connect opens the MongoDB client and creates the unique application ID index.
func (repository *Repository[M]) Connect(ctx context.Context) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(repository.config.MongoDBURI))
	if err != nil {
		return fmt.Errorf("connect to mongodb: %w", err)
	}
	repository.client = client
	if err := repository.ensureUniqueIDIndex(ctx); err != nil {
		_ = client.Disconnect(ctx)
		repository.client = nil
		return fmt.Errorf("creating mongodb identifier index: %w", err)
	}
	return nil
}

// Close releases the MongoDB client connection.
func (repository *Repository[M]) Close(ctx context.Context) error {
	return repository.client.Disconnect(ctx)
}
