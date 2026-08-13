package port

import "context"

type (
	// Consumer delivers decoded messages to a caller-provided handler.
	Consumer[D any] interface {
		Consume(context.Context, func(context.Context, D) error) error
	}
)
