package domain

import "context"

// OutboxEvent is one persisted message awaiting publication to a transport.
type OutboxEvent[D any] struct {
	ID       string
	Payload  D
	Attempts int
	// Context is restored by the persistence adapter and is never persisted itself.
	Context context.Context
}
