package adapter

import (
	"context"
)

// WithOperationTimeout bounds one MongoDB driver operation by the configured budget.
func (repository *Repository[M]) WithOperationTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, repository.config.OperationTimeout)
}
