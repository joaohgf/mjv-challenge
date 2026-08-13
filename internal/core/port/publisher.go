package port

import (
	"context"
)

type (
	// Publisher sends a message to an asynchronous transport.
	Publisher[D any] interface {
		Publish(context.Context, D) error
	}
)
