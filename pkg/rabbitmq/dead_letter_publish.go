package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// confirmation waits for RabbitMQ's acknowledgement of one publication.
type confirmation interface {
	WaitContext(context.Context) (bool, error)
}

// publishDeadLetter waits until RabbitMQ persists a mandatory parking-queue publication.
func (c *Consumer[M]) publishDeadLetter(ctx context.Context, message amqp091.Publishing) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.PublishTimeout)
	defer cancel()
	confirmation, err := c.channel.PublishWithDeferredConfirmWithContext(
		ctx, "", c.config.DeadLetterName, true, false, message,
	)
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}
	if err := waitConfirmation(ctx, confirmation); err != nil {
		return err
	}
	return c.returnedMessage()
}

// waitConfirmation validates the broker acknowledgement for one publication.
func waitConfirmation(ctx context.Context, result confirmation) error {
	if result == nil {
		return errors.New("publisher confirms are not enabled")
	}
	acknowledged, err := result.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("waiting publisher confirmation: %w", err)
	}
	if !acknowledged {
		return errors.New("message rejected by rabbitmq")
	}
	return nil
}

// returnedMessage reports a mandatory publication that RabbitMQ could not route.
func (c *Consumer[M]) returnedMessage() error {
	select {
	case returned, open := <-c.returns:
		if !open {
			return errors.New("rabbitmq return listener closed")
		}
		return fmt.Errorf("message returned by rabbitmq: %d %s", returned.ReplyCode, returned.ReplyText)
	default:
		return nil
	}
}
