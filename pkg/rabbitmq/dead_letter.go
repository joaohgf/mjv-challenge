package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"maps"

	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

// deadLetter republishes first and only acknowledges the source after it succeeds.
func (c *Consumer[M]) deadLetter(ctx context.Context, delivery amqp091.Delivery, cause error) (err error) {
	ctx, span := telemetry.StartSpan(ctx, "rabbitmq.dead_letter", trace.SpanKindProducer)
	defer func() {
		telemetry.End(span, err)
	}()
	message := toPublishing(ctx, delivery, cause)
	if err := c.publishToDeadLetter(ctx, message); err != nil {
		return fmt.Errorf("publishing dead-letter message: %w", err)
	}
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("acknowledging dead-lettered message: %w", err)
	}
	slog.Warn("message sent to dead-letter queue", "queue", c.config.DeadLetterName, "error", cause)
	return nil
}

// publishToDeadLetter uses the configured confirmed publisher for one failed delivery.
func (c *Consumer[M]) publishToDeadLetter(ctx context.Context, message amqp091.Publishing) error {
	if c.deadLetterPublish != nil {
		return c.deadLetterPublish(ctx, message)
	}
	return c.publishDeadLetter(ctx, message)
}

// toPublishing preserves AMQP metadata, records the failure and makes the parking message durable.
func toPublishing(ctx context.Context, delivery amqp091.Delivery, cause error) amqp091.Publishing {
	headers := amqp091.Table{}
	maps.Copy(headers, delivery.Headers)
	headers["x-failure-reason"] = cause.Error()
	headers = telemetry.InjectAMQPHeaders(ctx, headers)
	return amqp091.Publishing{
		Headers: headers, ContentType: delivery.ContentType, ContentEncoding: delivery.ContentEncoding,
		DeliveryMode: amqp091.Persistent, Priority: delivery.Priority, CorrelationId: delivery.CorrelationId,
		ReplyTo: delivery.ReplyTo, Expiration: delivery.Expiration, MessageId: delivery.MessageId,
		Timestamp: delivery.Timestamp, Type: delivery.Type, UserId: delivery.UserId, AppId: delivery.AppId,
		Body: delivery.Body,
	}
}
