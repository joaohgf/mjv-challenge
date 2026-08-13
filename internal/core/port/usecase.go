package port

import "context"

type (
	// Creator creates a new domain value.
	Creator[D any] interface {
		Create(context.Context, D) (D, error)
	}
	// Updater advances an existing domain value.
	Updater[D any] interface {
		Update(context.Context, D) (D, error)
	}
)
