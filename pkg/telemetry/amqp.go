package telemetry

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectAMQPHeaders writes the active trace context into AMQP headers.
func InjectAMQPHeaders(ctx context.Context, headers amqp091.Table) amqp091.Table {
	if headers == nil {
		headers = amqp091.Table{}
	}
	for key, value := range InjectContext(ctx, nil) {
		headers[key] = value
	}
	return headers
}

// ExtractAMQPHeaders restores a parent trace context from AMQP headers.
func ExtractAMQPHeaders(ctx context.Context, headers amqp091.Table) context.Context {
	values := make(map[string]string, len(headers))
	for key, value := range headers {
		if text, ok := value.(string); ok {
			values[key] = text
		}
	}
	return ExtractContext(ctx, values)
}

// InjectContext serializes the active trace context for durable storage.
func InjectContext(ctx context.Context, values map[string]string) map[string]string {
	if values == nil {
		values = map[string]string{}
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(values))
	return values
}

// ExtractContext restores a parent trace context that was durably stored.
func ExtractContext(ctx context.Context, values map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(values))
}
