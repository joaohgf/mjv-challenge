package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Shutdown flushes trace data and releases exporter resources.
type Shutdown func(context.Context) error

// Start configures OpenTelemetry providers when telemetry is enabled.
func Start(ctx context.Context, settings *config.Telemetry) (Shutdown, error) {
	if !settings.Enabled {
		return func(context.Context) error { return nil }, nil
	}
	traces, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(settings.Endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}
	resource := resource.NewWithAttributes("", attribute.String("service.name", settings.ServiceName))
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(traces), trace.WithResource(resource))
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return shutdown(tracerProvider), nil
}

// Close logs trace shutdown failures without masking the process result.
func Close(ctx context.Context, shutdown Shutdown) {
	if err := shutdown(ctx); err != nil {
		slog.Error("closing telemetry", "error", err)
	}
}

// shutdown flushes and stops the trace provider.
func shutdown(traces *trace.TracerProvider) Shutdown {
	return traces.Shutdown
}
