package telemetry

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestAMQPHeadersPropagateTraceContext(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator()) })
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled,
	})
	headers := InjectAMQPHeaders(trace.ContextWithSpanContext(context.Background(), parent), amqp091.Table{})
	restored := trace.SpanContextFromContext(ExtractAMQPHeaders(context.Background(), headers))

	if headers["traceparent"] == "" || restored.TraceID() != parent.TraceID() || restored.SpanID() != parent.SpanID() {
		t.Fatalf("expected trace context in AMQP headers, got %#v", headers)
	}
}

func TestContextPropagatesTraceContextThroughDurableMap(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator()) })
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}, TraceFlags: trace.FlagsSampled,
	})
	stored := InjectContext(trace.ContextWithSpanContext(context.Background(), parent), nil)
	restored := trace.SpanContextFromContext(ExtractContext(context.Background(), stored))

	if stored["traceparent"] == "" || restored.TraceID() != parent.TraceID() {
		t.Fatalf("expected trace context in durable map, got %#v", stored)
	}
}
