package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

// Publisher serializes a typed message and publishes it to the configured queue.
type Publisher[M any] struct {
	*RabbitMQ
}

// NewPublisher creates a typed publisher for the configured main queue.
func NewPublisher[M any](settings *config.Queue) *Publisher[M] {
	return &Publisher[M]{RabbitMQ: NewRabbitMQ(settings)}
}

// Connect opens AMQP resources and enables broker acknowledgements for publishes.
func (pb *Publisher[M]) Connect(ctx context.Context) error {
	if err := pb.RabbitMQ.Connect(ctx); err != nil {
		return err
	}
	if err := pb.channel.Confirm(false); err != nil {
		_ = pb.RabbitMQ.Close()
		return fmt.Errorf("enabling publisher confirms: %w", err)
	}
	return nil
}

// Publish encodes a persistent JSON message and waits for the broker confirmation.
func (pb *Publisher[M]) Publish(ctx context.Context, message M) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, pb.config.PublishTimeout)
	defer cancel()
	if err := pb.ensureConnected(ctx); err != nil {
		return fmt.Errorf("reconnecting publisher: %w", err)
	}
	confirmation, err := pb.channel.PublishWithDeferredConfirmWithContext(ctx, "", pb.config.Name, false, false, publishing(body))
	if err != nil {
		return pb.recover(ctx, fmt.Errorf("publishing message: %w", err))
	}
	if confirmation == nil {
		return fmt.Errorf("publisher confirms are not enabled")
	}
	ack, err := confirmation.WaitContext(ctx)
	if err != nil {
		return pb.recover(ctx, fmt.Errorf("waiting publisher confirmation: %w", err))
	}
	if !ack {
		return pb.recover(ctx, errors.New("message rejected by rabbitmq"))
	}
	return nil
}

// publishing applies the AMQP properties required for durable order messages.
func publishing(body []byte) amqp091.Publishing {
	return amqp091.Publishing{ContentType: "application/json", DeliveryMode: amqp091.Persistent, Body: body}
}

func (pb *Publisher[M]) recover(ctx context.Context, publishErr error) error {
	if !pb.IsClosed() {
		return publishErr
	}
	if err := pb.reconnect(ctx); err != nil {
		return errors.Join(publishErr, fmt.Errorf("reconnecting publisher: %w", err))
	}
	return publishErr
}

func (pb *Publisher[M]) ensureConnected(ctx context.Context) error {
	if !pb.IsClosed() {
		return nil
	}
	return pb.reconnect(ctx)
}

func (pb *Publisher[M]) reconnect(ctx context.Context) error {
	if err := pb.RabbitMQ.Close(); err != nil {
		return fmt.Errorf("closing disconnected publisher: %w", err)
	}
	if err := pb.Connect(ctx); err != nil {
		return err
	}
	slog.Info("rabbitmq publisher reconnected", "queue", pb.config.Name)
	return nil
}
