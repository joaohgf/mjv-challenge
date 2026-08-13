package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/joaohgf/mjv-challenge/config"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestStartDisabledReturnsNoopShutdown(t *testing.T) {
	shutdown, err := Start(context.Background(), &config.Telemetry{Enabled: false})
	if err != nil {
		t.Fatalf("expected disabled telemetry to start, got %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("expected noop shutdown, got %v", err)
	}
}

func TestShutdownStopsTraceProvider(t *testing.T) {
	if err := shutdown(trace.NewTracerProvider())(context.Background()); err != nil {
		t.Fatalf("expected trace provider to stop, got %v", err)
	}
}

func TestCloseDoesNotPropagateShutdownError(t *testing.T) {
	Close(context.Background(), func(context.Context) error { return errors.New("exporter unavailable") })
	Close(context.Background(), shutdown(trace.NewTracerProvider()))
}
