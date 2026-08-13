package port

import (
	"context"
)

type (
	// Publisher sends a message to an asynchronous transport.
	Publisher[D any] interface {
		Publish(context.Context, D) error
	}
	// DeadLetterPublisher parks a message that cannot be published normally.
	DeadLetterPublisher[D any] interface {
		DeadLetter(context.Context, D, error) error
	}
)
