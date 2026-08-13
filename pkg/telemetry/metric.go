package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// RecordHTTP captures HTTP request volume and elapsed time.
func RecordHTTP(ctx context.Context, method, route string, status int, elapsed time.Duration) {
	attrs := []attribute.KeyValue{attribute.String("http.request.method", method), attribute.String("http.route", route), attribute.Int("http.response.status_code", status)}
	record(ctx, "mjv.http.requests", 1, attrs...)
	record(ctx, "mjv.http.duration.ms", elapsed.Milliseconds(), attrs...)
}

// RecordOperation captures the result of one infrastructure operation.
func RecordOperation(ctx context.Context, system, operation string, err error) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	attrs := []attribute.KeyValue{attribute.String("system", system), attribute.String("operation", operation), attribute.String("outcome", outcome)}
	record(ctx, "mjv.operations", 1, attrs...)
}

// record writes one counter sample when the active meter supports it.
func record(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) {
	counter, err := otel.Meter(instrumentationName).Int64Counter(name)
	if err == nil {
		counter.Add(ctx, value, metric.WithAttributes(attrs...))
	}
}
