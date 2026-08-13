package adapter

import (
	"context"
	"log/slog"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/core/port"
	"github.com/joaohgf/mjv-challenge/internal/inbound/consumer/dto"
)

// Consumer maps queue payloads before invoking the order update use case.
type Consumer struct {
	consumer port.Consumer[*dto.Message[*dto.Order]]
	mapper   port.To[*dto.Order, *domain.Order]
	usecase  port.Updater[*domain.Order]
}

// NewConsumer wires a generic queue consumer to the order update use case.
func NewConsumer(
	consumer port.Consumer[*dto.Message[*dto.Order]],
	mapper port.To[*dto.Order, *domain.Order],
	usecase port.Updater[*domain.Order],
) *Consumer {
	target := new(Consumer)
	target.consumer = consumer
	target.mapper = mapper
	target.usecase = usecase
	return target
}

// Consume delegates delivery control to the transport adapter.
func (handler *Consumer) Consume(ctx context.Context) error {
	return handler.consumer.Consume(ctx, handler.update)
}

// update maps a message payload and advances the corresponding order state.
func (handler *Consumer) update(ctx context.Context, message *dto.Message[*dto.Order]) error {
	if err := validateMessage(message); err != nil {
		return err
	}
	slog.Debug("processing order message", "message_id", message.ID)
	order := handler.mapper.To(message.Payload)
	_, err := handler.usecase.Update(ctx, order)
	if err == nil {
		slog.Debug("order processed", "order_id", order.ID)
	}
	return err
}
