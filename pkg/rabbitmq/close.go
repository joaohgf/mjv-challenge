package adapter

import (
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// Close releases AMQP resources and accepts connections already closed remotely.
func (rb *Client) Close() error {
	channel, connection := rb.channel, rb.connection
	rb.channel, rb.connection = nil, nil
	if err := closeChannel(channel); err != nil {
		return err
	}
	return closeConnection(connection)
}

// IsClosed reports whether the AMQP resources are unavailable for use.
func (rb *Client) IsClosed() bool {
	return rb.channel == nil || rb.connection == nil || rb.channel.IsClosed() || rb.connection.IsClosed()
}

func closeChannel(channel *amqp091.Channel) error {
	if channel == nil || channel.IsClosed() {
		return nil
	}
	if err := channel.Close(); err != nil && !closedError(err) {
		return fmt.Errorf("closing rabbitmq channel: %w", err)
	}
	return nil
}

func closeConnection(connection *amqp091.Connection) error {
	if connection == nil || connection.IsClosed() {
		return nil
	}
	if err := connection.Close(); err != nil && !closedError(err) {
		return fmt.Errorf("closing rabbitmq connection: %w", err)
	}
	return nil
}

func closedError(err error) bool {
	var target *amqp091.Error
	return errors.As(err, &target) && target.Code == amqp091.ChannelError && target.Reason == amqp091.ErrClosed.Reason
}
