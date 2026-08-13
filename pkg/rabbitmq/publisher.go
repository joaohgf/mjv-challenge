package adapter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

// Publisher serializes a typed message and publishes it to the configured queue.
type Publisher[M any] struct {
	*Client
	queue string
}

// NewPublisher creates a typed publisher for the configured main queue.
func NewPublisher[M any](settings *config.Queue) *Publisher[M] {
	return newPublisher[M](settings, settings.Name)
}

// NewDeadLetterPublisher creates a typed publisher for the parking queue.
func NewDeadLetterPublisher[M any](settings *config.Queue) *Publisher[M] {
	return newPublisher[M](settings, settings.DeadLetterName)
}

// Connect opens AMQP resources and enables broker acknowledgements for publishes.
func (pb *Publisher[M]) Connect(ctx context.Context) error {
	if err := pb.Client.Connect(ctx); err != nil {
		return err
	}
	if err := pb.channel.Confirm(false); err != nil {
		_ = pb.Client.Close()
		return fmt.Errorf("enabling publisher confirms: %w", err)
	}
	return nil
}

// Publish encodes a persistent JSON message and waits for the broker confirmation.
func (pb *Publisher[M]) Publish(ctx context.Context, message M) (err error) {
	return pb.publish(ctx, message, "rabbitmq.publish", nil)
}

// DeadLetter publishes a failed message with its cause in the parking queue.
func (pb *Publisher[M]) DeadLetter(ctx context.Context, message M, cause error) error {
	return pb.publish(ctx, message, "rabbitmq.dead_letter", amqp091.Table{"x-failure-reason": cause.Error()})
}

func newPublisher[M any](settings *config.Queue, queue string) *Publisher[M] {
	return &Publisher[M]{Client: NewClient(settings), queue: queue}
}

func (pb *Publisher[M]) reconnect(ctx context.Context) error {
	if err := pb.Client.Close(); err != nil {
		return fmt.Errorf("closing disconnected publisher: %w", err)
	}
	if err := pb.Connect(ctx); err != nil {
		return err
	}
	slog.Info("rabbitmq publisher reconnected", "queue", pb.queue)
	return nil
}
