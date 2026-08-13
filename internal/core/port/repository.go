package port

import (
	"context"
)

type (
	// Identifiable exposes the stable identifier and its persistence field name.
	Identifiable interface {
		GetID() string
		GetIDField() string
	}
	// PersistentCreator inserts a new value without replacing an existing one.
	PersistentCreator[D any] interface {
		Create(context.Context, D) (D, error)
	}
	// PersistentUpdater replaces an existing value without creating a new one.
	PersistentUpdater[D any] interface {
		Update(context.Context, D) (D, error)
	}
	// Getter loads one value using its application identifier.
	Getter[D any] interface {
		Get(context.Context, string) (D, error)
	}
	// Store combines the persistence operations required by an order repository.
	Store[D any] interface {
		PersistentCreator[D]
		PersistentUpdater[D]
		Getter[D]
	}
)
