package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

// process acknowledges success or parks a decoding or handling failure in the DLQ.
func (c *Consumer[M]) process(ctx context.Context, delivery amqp091.Delivery, handle func(context.Context, M) error) (err error) {
	ctx = telemetry.ExtractAMQPHeaders(ctx, delivery.Headers)
	ctx, span := telemetry.StartSpan(ctx, "rabbitmq.consume", trace.SpanKindConsumer)
	defer func() { telemetry.End(span, err) }()
	source, err := c.decode(delivery.Body)
	if err != nil {
		return c.deadLetter(ctx, delivery, err)
	}
	if err := callHandler(ctx, source, handle); err != nil {
		return c.deadLetter(ctx, delivery, err)
	}
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("acknowledging message: %w", err)
	}
	return nil
}

// decode converts one JSON body to the generic consumer message type.
func (c *Consumer[M]) decode(body []byte) (M, error) {
	var message M
	if err := json.Unmarshal(body, &message); err != nil {
		return message, fmt.Errorf("decoding message: %w", err)
	}
	return message, nil
}

// callHandler converts an unexpected handler panic into a parking-queue failure.
func callHandler[M any](ctx context.Context, message M, handle func(context.Context, M) error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("handling message panic: %v", recovered)
		}
	}()
	return handle(ctx, message)
}
