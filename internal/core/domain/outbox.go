package domain

import "context"

// OutboxEvent is one persisted message awaiting publication to a transport.
type OutboxEvent[D any] struct {
	ID string
	// LeaseToken identifies the relay that currently owns this event lease.
	LeaseToken string
	Payload    D
	Attempts   int
	// Context is restored by the persistence adapter and is never persisted itself.
	Context context.Context
}
