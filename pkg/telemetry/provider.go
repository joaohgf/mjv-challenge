package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Shutdown flushes telemetry data and releases exporter resources.
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
	meterProvider, err := newMeterProvider(ctx, settings, resource)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return shutdown(tracerProvider, meterProvider), nil
}

// newMeterProvider configures metric export only for a compatible OTLP backend.
func newMeterProvider(ctx context.Context, settings *config.Telemetry, resource *resource.Resource) (*metric.MeterProvider, error) {
	if !settings.MetricsEnabled {
		return metric.NewMeterProvider(metric.WithResource(resource)), nil
	}
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(settings.Endpoint), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("creating OTLP metric exporter: %w", err)
	}
	return metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exporter)), metric.WithResource(resource)), nil
}

// Close logs exporter shutdown failures without masking the process result.
func Close(ctx context.Context, shutdown Shutdown) {
	if err := shutdown(ctx); err != nil {
		slog.Error("closing telemetry", "error", err)
	}
}

// shutdown flushes the metric and trace providers in the reverse setup order.
func shutdown(traces *trace.TracerProvider, metrics *metric.MeterProvider) Shutdown {
	return func(ctx context.Context) error {
		return errors.Join(metrics.Shutdown(ctx), traces.Shutdown(ctx))
	}
}
