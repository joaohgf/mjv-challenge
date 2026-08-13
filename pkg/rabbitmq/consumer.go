package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

// Consumer decodes deliveries and coordinates manual acknowledgement.
type Consumer[M any] struct {
	*RabbitMQ
}

// NewConsumer creates a typed consumer for the configured main queue.
func NewConsumer[M any](config *config.Queue) *Consumer[M] {
	return &Consumer[M]{RabbitMQ: NewRabbitMQ(config)}
}

// Consume processes one unacknowledged delivery at a time until context ends.
func (c *Consumer[M]) Consume(ctx context.Context, handle func(context.Context, M) error) error {
	if err := c.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("configuring consumer prefetch: %w", err)
	}
	deliveries, err := c.channel.ConsumeWithContext(ctx, c.config.Name, "worker", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("starting consumer: %w", err)
	}
	for delivery := range deliveries {
		if err := c.process(ctx, delivery, handle); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}
	return errors.New("rabbitmq deliveries channel closed")
}

// process acknowledges success or parks a decoding or handling failure in the DLQ.
func (c *Consumer[M]) process(ctx context.Context, delivery amqp091.Delivery, handle func(context.Context, M) error) error {
	source, err := c.decode(delivery.Body)
	if err != nil {
		return c.deadLetter(ctx, delivery, err)
	}
	if err := handle(ctx, source); err != nil {
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
