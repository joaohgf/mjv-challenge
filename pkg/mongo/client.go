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

// Connect opens the MongoDB client after Docker initializes the database schema.
func (repository *Repository[M]) Connect(ctx context.Context) error {
	operationContext, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	client, err := mongo.Connect(operationContext, options.Client().ApplyURI(repository.config.MongoDBURI))
	if err != nil {
		return fmt.Errorf("connect to mongodb: %w", err)
	}
	repository.client = client
	if err := repository.Ping(ctx); err != nil {
		_ = repository.Close(context.Background())
		repository.client = nil
		return err
	}
	return nil
}

// Ping verifies that MongoDB can serve requests within the operation budget.
func (repository *Repository[M]) Ping(ctx context.Context) error {
	if repository.client == nil {
		return fmt.Errorf("ping mongodb: client is not connected")
	}
	operationContext, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	if err := repository.client.Ping(operationContext, nil); err != nil {
		return fmt.Errorf("ping mongodb: %w", err)
	}
	return nil
}

// Close releases the MongoDB client connection.
func (repository *Repository[M]) Close(ctx context.Context) error {
	if repository.client == nil {
		return nil
	}
	operationContext, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	return repository.client.Disconnect(operationContext)
}
