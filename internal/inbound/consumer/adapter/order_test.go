package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/joaohgf/mjv-challenge/internal/core/domain"
	"github.com/joaohgf/mjv-challenge/internal/inbound/consumer/dto"
	"github.com/joaohgf/mjv-challenge/internal/inbound/consumer/mapper"
)

type (
	consumerStub struct {
		handler func(context.Context, *dto.Message[*dto.Order]) error
	}
	updaterStub struct {
		updated bool
		err     error
	}
)

func (stub *consumerStub) Consume(ctx context.Context, handler func(context.Context, *dto.Message[*dto.Order]) error) error {
	stub.handler = handler
	return handler(ctx, &dto.Message[*dto.Order]{Payload: new(dto.Order)})
}

func (stub *updaterStub) Update(_ context.Context, order *domain.Order) (*domain.Order, error) {
	stub.updated = true
	return order, stub.err
}

func TestConsumerReturnsUpdateError(t *testing.T) {
	consumer := new(consumerStub)
	updater := &updaterStub{err: errors.New("database unavailable")}
	handler := NewConsumer(consumer, &mapper.Order{}, updater)

	if err := handler.Consume(context.Background()); err == nil {
		t.Fatal("expected update error")
	}
}

func TestConsumerConsumesAndUpdatesOrder(t *testing.T) {
	consumer := new(consumerStub)
	updater := new(updaterStub)
	handler := NewConsumer(consumer, &mapper.Order{}, updater)

	if err := handler.Consume(context.Background()); err != nil {
		t.Fatal(err)
	}
	if consumer.handler == nil || !updater.updated {
		t.Fatal("expected consumer handler to update the order")
	}
}

func TestConsumerRejectsIncompleteMessage(t *testing.T) {
	tests := []struct {
		name    string
		message *dto.Message[*dto.Order]
	}{
		{name: "missing envelope"},
		{name: "missing payload", message: new(dto.Message[*dto.Order])},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updater := new(updaterStub)
			handler := NewConsumer(new(consumerStub), &mapper.Order{}, updater)

			if err := handler.update(context.Background(), test.message); err == nil {
				t.Fatal("expected incomplete message error")
			}
			if updater.updated {
				t.Fatal("expected mapper and use case to be skipped")
			}
		})
	}
}
