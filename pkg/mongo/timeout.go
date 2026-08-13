package adapter

import (
	"context"
)

// withSaveTimeout bounds a persistence operation by the configured write budget.
func (repository *Repository[M]) withSaveTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, repository.config.SaveTimeout)
}
