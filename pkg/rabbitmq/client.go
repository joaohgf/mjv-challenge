package adapter

import (
	"context"
	"fmt"

	"github.com/joaohgf/mjv-challenge/config"
	"github.com/rabbitmq/amqp091-go"
)

// Client owns an AMQP connection and one channel for a publisher or consumer.
type Client struct {
	config     *config.Queue
	connection *amqp091.Connection
	channel    *amqp091.Channel
}

// NewClient builds a disconnected AMQP client with queue configuration.
func NewClient(config *config.Queue) *Client {
	target := new(Client)
	target.config = config
	return target
}

// Connect opens an AMQP connection and the channel shared by one adapter.
func (rb *Client) Connect(ctx context.Context) error {
	connection, err := amqp091.Dial(rb.config.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("connecting rabbitMQ: %w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return fmt.Errorf("creating channel: %w", err)
	}
	rb.connection = connection
	rb.channel = channel
	return nil
}
