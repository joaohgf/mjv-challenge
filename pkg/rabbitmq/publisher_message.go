package adapter

import (
	"context"

	"github.com/joaohgf/mjv-challenge/pkg/telemetry"
	"github.com/rabbitmq/amqp091-go"
)

// publishing applies the AMQP properties required for durable order messages.
func publishing(ctx context.Context, body []byte, headers amqp091.Table) amqp091.Publishing {
	return amqp091.Publishing{
		Headers: telemetry.InjectAMQPHeaders(ctx, headers), ContentType: "application/json",
		DeliveryMode: amqp091.Persistent, Body: body,
	}
}
