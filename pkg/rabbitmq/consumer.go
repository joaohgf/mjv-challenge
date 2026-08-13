package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

// Consumer decodes deliveries and coordinates manual acknowledgement.
type Consumer[M any] struct {
	*Client
	// returns receives routing failures for mandatory dead-letter publications.
	returns <-chan amqp091.Return
	// deadLetterPublish enables the confirmed publication path to be unit tested.
	deadLetterPublish func(context.Context, amqp091.Publishing) error
}

// NewConsumer creates a typed consumer for the configured main queue.
func NewConsumer[M any](config *config.Queue) *Consumer[M] {
	target := &Consumer[M]{Client: NewClient(config)}
	target.deadLetterPublish = target.publishDeadLetter
	return target
}

// Connect opens AMQP resources and enables confirms for dead-letter publishing.
func (c *Consumer[M]) Connect(ctx context.Context) error {
	if err := c.Client.Connect(ctx); err != nil {
		return err
	}
	if err := enableConfirms(c.channel); err != nil {
		_ = c.Client.Close()
		return fmt.Errorf("enabling publisher confirms: %w", err)
	}
	c.returns = c.channel.NotifyReturn(make(chan amqp091.Return, 1))
	return nil
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
