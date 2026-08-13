package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

// publish serializes, delivers and confirms a message in the selected queue.
func (pb *Publisher[M]) publish(ctx context.Context, message M, name string, headers amqp091.Table) (err error) {
	ctx, span := telemetry.StartSpan(ctx, name, trace.SpanKindProducer)
	defer func() {
		telemetry.RecordOperation(ctx, "rabbitmq", name, err)
		telemetry.End(span, err)
	}()
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, pb.config.PublishTimeout)
	defer cancel()
	if err := pb.ensureConnected(ctx); err != nil {
		return fmt.Errorf("reconnecting publisher: %w", err)
	}
	confirmation, err := pb.channel.PublishWithDeferredConfirmWithContext(ctx, "", pb.queue, false, false, publishing(ctx, body, headers))
	if err != nil {
		return pb.recover(ctx, fmt.Errorf("publishing message: %w", err))
	}
	if confirmation == nil {
		return errors.New("publisher confirms are not enabled")
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

// ensureConnected reconnects the publisher before the next delivery if needed.
func (pb *Publisher[M]) ensureConnected(ctx context.Context) error {
	if !pb.IsClosed() {
		return nil
	}
	return pb.reconnect(ctx)
}

// recover reconnects after a closed connection while preserving the publish error.
func (pb *Publisher[M]) recover(ctx context.Context, publishErr error) error {
	if !pb.IsClosed() {
		return publishErr
	}
	if err := pb.reconnect(ctx); err != nil {
		return errors.Join(publishErr, fmt.Errorf("reconnecting publisher: %w", err))
	}
	return publishErr
}
