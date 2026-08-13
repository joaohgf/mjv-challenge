package adapter

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// Transaction runs an operation atomically using the repository MongoDB client.
func (repository *Repository[M]) Transaction(ctx context.Context, operation func(context.Context) error) error {
	ctx, cancel := repository.WithOperationTimeout(ctx)
	defer cancel()
	session, err := repository.client.StartSession()
	if err != nil {
		return fmt.Errorf("starting mongodb session: %w", err)
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(sessionContext mongo.SessionContext) (interface{}, error) {
		return nil, operation(sessionContext)
	})
	if err != nil {
		return fmt.Errorf("running mongodb transaction: %w", err)
	}
	return nil
}
