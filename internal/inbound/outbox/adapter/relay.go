package adapter

import (
	"context"
	"log/slog"
	"time"

	"github.com/joaohgf/mjv-challenge/internal/core/port"
	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
)

// Relay periodically dispatches durable outbox events from the worker process.
type Relay struct {
	dispatcher port.Dispatcher
	interval   time.Duration
}

// NewRelay builds the worker loop for the supplied dispatcher and interval.
func NewRelay(dispatcher port.Dispatcher, interval time.Duration) *Relay {
	return &Relay{dispatcher: dispatcher, interval: interval}
}

// Run dispatches available events and waits between idle or failed attempts.
func (relay *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(relay.interval)
	defer ticker.Stop()
	for {
		dispatched, err := relay.dispatch(ctx)
		if err != nil {
			slog.Error("dispatching outbox event", "error", err)
		}
		if dispatched && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (relay *Relay) dispatch(ctx context.Context) (dispatched bool, err error) {
	ctx, span := telemetry.StartSpan(ctx, "outbox.dispatch", trace.SpanKindProducer)
	defer func() {
		telemetry.End(span, err)
	}()
	return relay.dispatcher.Dispatch(ctx)
}
