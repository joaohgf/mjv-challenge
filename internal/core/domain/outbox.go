package domain

// OutboxEvent is one persisted message awaiting publication to a transport.
type OutboxEvent[D any] struct {
	ID       string
	Payload  D
	Attempts int
}
